// Package gounions generate JSON wrapper code for union types and
// types using them
package gounions

import (
	"fmt"
	"go/types"
	"reflect"
	"strings"

	an "github.com/benoitkugler/gomacro/analysis"
	gen "github.com/benoitkugler/gomacro/generator"
)

// Generate walks through all the types in `Source`
// and generate JSON routines for unions.
func Generate(an *an.Analysis) []gen.Declaration {
	var out []gen.Declaration

	out = append(out, gen.Declaration{
		ID: "aa_header",
		Content: fmt.Sprintf(`package %s
	
		import "encoding/json"
		
		// Code generated by gomacro/generator/gounions. DO NOT EDIT
		`, an.Root.Name),
		Priority: true,
	})

	ctx := context{cache: make(gen.Cache), srcPkg: an.Root.Types}
	for _, typ := range an.Source {
		out = append(out, ctx.generate(an.Types[typ])...)
	}

	return out
}

type context struct {
	cache  gen.Cache
	srcPkg *types.Package
}

func (ctx context) generate(typ an.Type) []gen.Declaration {
	if ctx.cache.Check(typ) {
		return nil
	}

	switch typ := typ.(type) {
	case *an.Basic, *an.Time, *an.Extern, *an.Enum:
		return nil
	case *an.Pointer:
		return ctx.generate(typ.Elem)
	case *an.Array:
		if _, isElemUnion := typ.Elem.(*an.Union); isElemUnion {
			panic("anonymous arrays containing unions are not supported")
		}
		return nil
	case *an.Map:
		if _, isElemUnion := typ.Elem.(*an.Union); isElemUnion {
			panic("anonymous maps containing unions are not supported")
		}
		return nil
	case *an.Named:
		return ctx.codeForNamed(typ)
	case *an.Union:
		return ctx.codeForUnion(typ)
	case *an.Struct:
		return ctx.codeForStruct(typ)
	default:
		panic(an.ExhaustiveTypeSwitch)
	}
}

func (ctx context) isLocal(ty *types.Named) bool {
	return ty.Obj().Pkg() == ctx.srcPkg
}

func (ctx context) codeForUnion(u *an.Union) []gen.Declaration {
	// only add wrapper code for "internal" unions
	if !ctx.isLocal(u.Type().(*types.Named)) {
		return nil
	}

	// the concret types do no need code generation, just returns
	// the JSON routines
	return []gen.Declaration{{ID: an.LocalName(u), Content: jsonForUnion(u)}}
}

// return the go code implementing JSON convertions
func jsonForUnion(u *an.Union) string {
	var casesFrom, casesTo, kindDecls []string

	name := an.LocalName(u)
	wrapperName := name + "Wrapper"

	for _, member := range u.Members {
		memberName := an.LocalName(member)
		kindValue := memberName
		kindVarName := memberName + name[0:2] + "Kind"
		kindDecls = append(kindDecls, fmt.Sprintf("%s = %q", kindVarName, kindValue))

		casesFrom = append(casesFrom, fmt.Sprintf(`case %q:
			var data %s
			err = json.Unmarshal(wr.Data, &data)
			out.Data = data
	`, kindValue, memberName))

		caseTo := fmt.Sprintf(`case %s:
			wr = wrapper{Kind: %q, Data: data}
		`, memberName, kindValue)
		casesTo = append(casesTo, caseTo)
	}

	codeKinds := fmt.Sprintf(`
	const (
		%s
	)
	`, strings.Join(kindDecls, "\n"))

	codeFrom := fmt.Sprintf(`func (out *%s) UnmarshalJSON(src []byte) error {
		var wr struct {
			Kind string
			Data json.RawMessage
		}
		err := json.Unmarshal(src, &wr)
		if err != nil {
			return err
		}
		switch wr.Kind {
			%s
		default:
			panic("exhaustive switch")
		}
		return err
	}
	`, wrapperName, strings.Join(casesFrom, ""))

	codeTo := fmt.Sprintf(`func (item %s) MarshalJSON() ([]byte, error) {
		type wrapper struct {
			Data interface{}
			Kind string
		}
		var wr wrapper
		switch data := item.Data.(type) {
		%s
		default:
			panic("exhaustive switch")
		}
		return json.Marshal(wr)
	}
	`, wrapperName, strings.Join(casesTo, ""))

	return fmt.Sprintf(`
	// %s may be used as replacements for %s 
	// when working with JSON
	type %s struct{
		Data %s
	}

	%s 

	%s

	%s
	`, wrapperName, name, wrapperName, name, codeFrom, codeTo, codeKinds)
}

func (ctx context) codeForNamed(typ *an.Named) (out []gen.Declaration) {
	if !ctx.isLocal(typ.Type().(*types.Named)) {
		return nil
	}
	switch under := typ.Underlying.(type) {
	case *an.Array:
		if elem, isElemUnion := under.Elem.(*an.Union); isElemUnion {
			out = append(out, ctx.codeForUnion(elem)...)
			out = append(out, gen.Declaration{ID: an.LocalName(typ), Content: jsonForArray(typ)})
		}
		return out
	case *an.Map:
		if elem, isElemUnion := under.Elem.(*an.Union); isElemUnion {
			out = append(out, ctx.codeForUnion(elem)...)
			out = append(out, gen.Declaration{ID: an.LocalName(typ), Content: jsonForMap(typ)})
		}
		return out
	case *an.Basic, *an.Time: // nothing to do
		return nil
	default:
		panic(an.ExhaustiveAnonymousTypeSwitch)
	}
}

// assume typ.Underlying is Array,
// and array elements are union
func jsonForArray(typ *an.Named) string {
	ar := typ.Underlying.(*an.Array)
	elem := ar.Elem.(*an.Union)

	name := an.LocalName(typ)
	elemName := an.LocalName(elem)

	return fmt.Sprintf(`func (list %s) MarshalJSON() ([]byte, error) {
		tmp := make([]%sWrapper, len(list))
		for i, v := range list {
			tmp[i].Data = v
		}
		return json.Marshal(tmp)
	}
	
	func (list *%s) UnmarshalJSON(data []byte) error {
		var tmp []%sWrapper
		err := json.Unmarshal(data, &tmp)
		*list = make(%s, len(tmp))
		for i, v := range tmp {
			(*list)[i] = v.Data
		}
		return err
	}`, name, elemName, name, elemName, name)
}

// assume typ.Underlying is Map,
// and map elements are union
func jsonForMap(typ *an.Named) string {
	ar := typ.Underlying.(*an.Map)
	elem := ar.Elem.(*an.Union)

	name := an.LocalName(typ)
	// since we write methods, we can safely assume the output package is the same as `typ`
	keyName := types.TypeString(ar.Key.Type(), gen.NameRelativeTo(typ.Type().(*types.Named).Obj().Pkg()))
	elemName := an.LocalName(elem)

	return fmt.Sprintf(`func (dict %[1]s) MarshalJSON() ([]byte, error) {
		tmp := make(map[%[2]s]%[3]sWrapper)
		for k, v := range dict {
			tmp[k] = %[3]sWrapper{v}
		}
		return json.Marshal(tmp)
	}

	func (dict *%[1]s) UnmarshalJSON(src []byte) error {
		var wr map[%[2]s]%[3]sWrapper
		err := json.Unmarshal(src, &wr)
		if err != nil {
			return err
		}

		*dict = make(%[1]s)
		for i, v := range wr {
			(*dict)[i] = v.Data
		}
		return nil
	}
	`, name, keyName, elemName)
}

// returns true if the field type in an union
func requireWrapper(field an.StructField) bool {
	_, isUnion := field.Type.(*an.Union)
	return isUnion
}

func (ctx context) codeForStruct(st *an.Struct) (out []gen.Declaration) {
	var structRequireWrapper bool
	for _, field := range st.Fields {
		if reflect.StructTag(field.Tag).Get("gomacro") != "ignore" {
			// recurse
			out = append(out, ctx.generate(field.Type)...)
		}
		if requireWrapper(field) {
			structRequireWrapper = true
		}
	}

	// check for field requiring the wrapper
	if !structRequireWrapper {
		return out
	}

	var fieldsDef, fieldsAssignToWrapper, fieldsAssignFromWrapper []string
	for _, field := range st.Fields {
		fieldName := field.Field.Name()

		fieldType := types.TypeString(field.Field.Type(), gen.NameRelativeTo(st.Type().(*types.Named).Obj().Pkg()))

		fieldAssignToW := fmt.Sprintf("item.%s", fieldName)
		fieldAssignFromW := fmt.Sprintf("wr.%s", fieldName)
		if requireWrapper(field) { // wrapper field
			fieldType = fieldType + "Wrapper"
			fieldAssignToW = fmt.Sprintf("%s{item.%s}", fieldType, fieldName)
			fieldAssignFromW = fmt.Sprintf("wr.%s.Data", fieldName)
		}

		fieldsDef = append(fieldsDef, fieldName+" "+fieldType)
		fieldsAssignToWrapper = append(fieldsAssignToWrapper, fmt.Sprintf("%s: %s,", fieldName, fieldAssignToW))
		fieldsAssignFromWrapper = append(fieldsAssignFromWrapper, fmt.Sprintf("item.%s = %s", fieldName, fieldAssignFromW))
	}

	name := an.LocalName(st)
	code := fmt.Sprintf(`
		func (item %s) MarshalJSON() ([]byte, error) {
			type wrapper struct {
				%s
			}
			wr := wrapper{
				%s
			}
			return json.Marshal(wr)
		}

		func (item *%s) UnmarshalJSON(src []byte) error {
			type wrapper struct {
				%s
			}
			var wr wrapper 
			err := json.Unmarshal(src, &wr)
			if err != nil {
				return err
			}
			%s
			return nil
		}
		`,
		name, strings.Join(fieldsDef, "\n"), strings.Join(fieldsAssignToWrapper, "\n"),
		name, strings.Join(fieldsDef, "\n"), strings.Join(fieldsAssignFromWrapper, "\n"),
	)
	out = append(out, gen.Declaration{
		ID:      name + "_json",
		Content: code,
	})
	return out
}

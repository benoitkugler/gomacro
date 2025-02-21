package randdata

import (
	"fmt"
	"go/types"
	"strings"

	an "github.com/benoitkugler/gomacro/analysis"
	gen "github.com/benoitkugler/gomacro/generator"
)

// Generate generates the code for random data generation of
// types defined in the analysis `Source`.
func Generate(ana *an.Analysis) []gen.Declaration {
	return generateWithTarget(ana, ana.Root.Types)
}

func generateWithTarget(ana *an.Analysis, targetPackage *types.Package) []gen.Declaration {
	var (
		out []gen.Declaration
		ctx = context{cache: make(gen.Cache), targetPackage: targetPackage}
	)

	for _, typ := range ana.Source {
		out = append(out, ctx.generate(ana.Types[typ])...)
	}

	imports := ctx.cache.Imports()

	out = append(out, gen.Declaration{
		ID: "__header",
		Content: fmt.Sprintf(`
		package %s

		import (
			%s
		)
			
		// Code generated by gomacro/generator/go/randdata. DO NOT EDIT.

		`, targetPackage.Name(), strings.Join(imports, "\n")),
		Priority: true,
	})

	return out
}

type context struct {
	targetPackage *types.Package
	cache         gen.Cache
}

// return the name or description in Go of the type `typ`
func (ctx context) typeName(typ an.Type) string {
	return types.TypeString(typ.Type(), gen.NameRelativeTo(ctx.targetPackage))
}

func (ctx context) functionID(ty an.Type) string {
	switch ty := ty.(type) {
	case *an.Basic:
		if ty.B.Name() == "byte" {
			return "uint8" // prefer not to use the alias
		} else if ty.B.Name() == "rune" {
			return "int32"
		}
		return ty.B.Name()
	case *an.Time:
		if ty.IsDate {
			return "tDate"
		}
		return "tTime"
	case *an.Pointer:
		return ctx.functionID(ty.Elem) + "Ptr"
	case *an.Array:
		if L := ty.Len; L != -1 {
			return fmt.Sprintf("Ar%d_%s", L, ctx.functionID(ty.Elem))
		}
		return fmt.Sprintf("Slice%s", ctx.functionID(ty.Elem))
	case *an.Map:
		return fmt.Sprintf("Map%s%s", ctx.functionID(ty.Key), ctx.functionID(ty.Elem))
	case *an.Named, *an.Enum, *an.Struct, *an.Union:
		// build a string usable in function names
		obj := ty.Type().(*types.Named).Obj()
		packageName := obj.Pkg().Name()
		localName := obj.Name()
		if obj.Pkg() == ctx.targetPackage {
			return localName
		}
		return packageName[:3] + "_" + localName
	default:
		panic(an.ExhaustiveTypeSwitch)
	}
}

func (ctx context) generate(ty an.Type) []gen.Declaration {
	if ctx.cache.Check(ty) {
		return nil
	}

	switch ty := ty.(type) {
	case *an.Basic:
		return []gen.Declaration{ctx.codeForBasic(ty)}
	case *an.Time:
		return []gen.Declaration{ctx.codeForTime(ty)}
	case *an.Pointer:
		return ctx.codeForPointer(ty)
	case *an.Named:
		return ctx.codeForNamed(ty)
	case *an.Array:
		return ctx.codeForArray(ty)
	case *an.Map:
		return ctx.codeForMap(ty)
	case *an.Enum:
		return []gen.Declaration{ctx.codeForEnum(ty)}
	case *an.Struct:
		return ctx.codeForStruct(ty)
	case *an.Union:
		return ctx.codeForUnion(ty)
	default:
		panic(an.ExhaustiveTypeSwitch)
	}
}

func (ctx context) codeForBasic(bs *an.Basic) gen.Declaration {
	var code string
	switch bs.B.Kind() {
	case types.Bool:
		code = fnBool()
	case types.Int:
		code = fnInt("int")
	case types.Int32:
		code = fnInt("int32")
	case types.Int64:
		code = fnInt("int64")
	case types.Uint8:
		code = fnInt("uint8")
	case types.Int8:
		code = fnInt("int8")
	case types.Int16:
		code = fnInt("int16")
	case types.Uint16:
		code = fnInt("uint16")
	case types.Float64:
		code = fnFloat64()
	case types.String:
		code = fnString()
	default:
		panic(fmt.Sprintf("basic type %v not supported", bs.B))
	}
	return gen.Declaration{ID: ctx.functionID(bs), Content: code}
}

func fnBool() string {
	return `
	func randbool() bool {
		i := rand.Int31n(2)
		return i == 1
	}`
}

func fnInt(intType string) string {
	return fmt.Sprintf(`
	func rand%s() %s {
		return %s(rand.Intn(1000000))
	}`, intType, intType, intType)
}

func fnFloat64() string {
	return `
	func randfloat64() float64 {
		return rand.Float64() * float64(rand.Int31())
	}`
}

func fnString() string {
	return `
	var letterRunes2  = []rune("azertyuiopqsdfghjklmwxcvbn123456789é@!?&èïab ")

	func randstring() string {
		b := make([]rune, 10)
		maxLength := len(letterRunes2)		
		for i := range b {
			b[i] = letterRunes2[rand.Intn(maxLength)]
		}
		return string(b)
	}`
}

func (ctx context) codeForTime(ty *an.Time) gen.Declaration {
	id := ctx.functionID(ty)
	content := fmt.Sprintf(`
	func rand%s() time.Time {
		return time.Unix(int64(rand.Int31()), 5)
	}
	`, id)
	return gen.Declaration{ID: id, Content: content}
}

func (ctx context) codeForPointer(ty *an.Pointer) []gen.Declaration {
	out := ctx.generate(ty.Elem) // recurse
	elemName := ctx.typeName(ty.Elem)
	id := ctx.functionID(ty)
	decl := gen.Declaration{
		ID: id, Content: fmt.Sprintf(`
		func rand%s() *%s {
			data := rand%s()
			return &data
		}`, id, elemName, ctx.functionID(ty.Elem)),
	}

	out = append(out, decl)
	return out
}

func (ctx context) codeForArray(ty *an.Array) []gen.Declaration {
	decls := ctx.generate(ty.Elem) // recurse for deps

	elemString := ctx.typeName(ty.Elem)
	elemID := ctx.functionID(ty.Elem)
	id := ctx.functionID(ty)
	var code string
	if ty.Len != -1 {
		code = fmt.Sprintf(`
		func rand%s() [%d]%s {
			var out [%d]%s
			for i := range out {
				out[i] = rand%s()
			}
			return out
		}`, id, ty.Len, elemString, ty.Len, elemString, elemID)
	} else {
		code = fmt.Sprintf(`
		func rand%s() []%s {
			l := 3 + rand.Intn(5)
			out := make([]%s, l)
			for i := range out {
				out[i] = rand%s()
			}
			return out
		}`, id, elemString, elemString, elemID)
	}
	decls = append(decls, gen.Declaration{ID: id, Content: code})
	return decls
}

func (ctx context) codeForMap(ty *an.Map) []gen.Declaration {
	decls := append(ctx.generate(ty.Key), ctx.generate(ty.Elem)...) // recurse for deps

	keyString := ctx.typeName(ty.Key)
	elemString := ctx.typeName(ty.Elem)
	id, keyID, elemID := ctx.functionID(ty), ctx.functionID(ty.Key), ctx.functionID(ty.Elem)
	decl := gen.Declaration{
		ID: id, Content: fmt.Sprintf(`
	func rand%s() map[%s]%s {
		l := 40 + rand.Intn(10)
		out := make(map[%s]%s, l)
		for i := 0; i < l; i++ {
			out[rand%s()] = rand%s()
		}
		return out
	}`, id, keyString, elemString, keyString, elemString, keyID, elemID),
	}

	decls = append(decls, decl)
	return decls
}

func (ctx context) codeForStruct(ty *an.Struct) (decls []gen.Declaration) {
	fieldsCode := ""
	for _, field := range ty.Fields {
		if !field.Field.Exported() || field.Tag.Get("gomacro-data") == "ignore" {
			continue
		}

		decls = append(decls, ctx.generate(field.Type)...) // recurse
		fieldsCode += fmt.Sprintf("s.%s = rand%s()\n", field.Field.Name(), ctx.functionID(field.Type))
	}
	id, name := ctx.functionID(ty), ctx.typeName(ty)
	code := gen.Declaration{
		ID: id, Content: fmt.Sprintf(`
	func rand%s() %s {
		var s %s
		%s
		return s
	}`, id, name, name, fieldsCode),
	}

	decls = append(decls, code)
	return decls
}

func (ctx context) codeForNamed(ty *an.Named) []gen.Declaration {
	decls := ctx.generate(ty.Underlying)

	name, id := ctx.typeName(ty), ctx.functionID(ty)
	decl := gen.Declaration{
		ID: id,
		Content: fmt.Sprintf(`
	func rand%s() %s {
		return %s(rand%s())
	}`, id, name, name, ctx.functionID(ty.Underlying)),
	}

	decls = append(decls, decl)
	return decls
}

func (ctx context) codeForEnum(ty *an.Enum) gen.Declaration {
	id, name := ctx.functionID(ty), ctx.typeName(ty)

	choices := make([]string, len(ty.Members))
	for i, val := range ty.Members {
		if !val.Const.Exported() {
			continue
		}

		fullString := types.ObjectString(val.Const, gen.NameRelativeTo(ctx.targetPackage))
		choices[i] = strings.Fields(fullString)[1]
	}

	return gen.Declaration{
		ID: id, Content: fmt.Sprintf(`
	func rand%s() %s {
		choix := [...]%s{%s}
		i := rand.Intn(len(choix))
		return choix[i]
	}`, id, name, name, strings.Join(choices, ", ")),
	}
}

func (ctx context) codeForUnion(ty *an.Union) []gen.Declaration {
	var (
		choix []string
		out   []gen.Declaration
	)
	for _, member := range ty.Members {
		choix = append(choix, fmt.Sprintf("rand%s(),\n", ctx.functionID(member)))
		out = append(out, ctx.generate(member)...) // recurse
	}

	id, name := ctx.functionID(ty), ctx.typeName(ty)
	out = append(out, gen.Declaration{
		ID: id,
		Content: fmt.Sprintf(`
		func rand%s() %s {
			choix := [...]%s{
				%s
			}
			i := rand.Intn(%d)
			return choix[i]
		}`,
			id, name, name,
			strings.Join(choix, ""),
			len(ty.Members)),
	})

	return out
}

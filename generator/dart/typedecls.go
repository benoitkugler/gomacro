// Package implements a code generator outputting
// Dart code for type definitions and JSON routines.
package dart

import (
	"fmt"
	"strings"

	an "github.com/benoitkugler/gomacro/analysis"
	gen "github.com/benoitkugler/gomacro/generator"
)

// lowerFirst converts field names to the Dart convention
func lowerFirst(s string) string {
	return strings.ToLower(s[0:1]) + s[1:]
}

// typeName returns the Dart string used to refer to this type
// (not to be confused with the type declaration)
func typeName(typ an.Type) string {
	switch typ := typ.(type) {
	case *an.Pointer:
		panic("pointers not handled by the Dart generator")
	case *an.Basic: // may be named or not
		switch typ.Kind() {
		case an.BKBool:
			return "bool"
		case an.BKInt:
			return "int"
		case an.BKFloat:
			return "double"
		case an.BKString:
			return "String"
		default:
			panic(an.ExhaustiveBasicKindSwitch)
		}
	case *an.Time:
		return "DateTime"
	case *an.Array:
		return fmt.Sprintf("List<%s>", typeName(typ.Elem))
	case *an.Map:
		return fmt.Sprintf("Map<%s,%s>", typeName(typ.Key), typeName(typ.Elem))
	case *an.Named, *an.Extern, *an.Struct, *an.Enum, *an.Union: // these types are always named
		return strings.Title(an.LocalName(typ)) // Dart convention
	default:
		panic(an.ExhaustiveTypeSwitch + fmt.Sprintf(": %T", typ))
	}
}

func (buf buffer) codeForNamed(typ *an.Named) (gen.Declaration, string) {
	// type wrapper
	code := fmt.Sprintf(`
	// %s
	typedef %s = %s;
	
	%s
	`, gen.Origin(typ), typeName(typ), typeName(typ.Underlying),
		jsonForNamed(typ),
	)
	out := gen.Declaration{ID: typeName(typ), Content: code}

	// recurse for the underlying code
	importFile := buf.generate(typ.Underlying, buf.linker.GetOutput(typ.Type()))

	return out, importFile
}

func codeForBasic(typ *an.Basic) gen.Declaration {
	// nothing to declare, since basic types are predeclared
	return gen.Declaration{ID: typeName(typ) + "_json", Content: jsonForBasic(typ)}
}

func codeForTime(typ *an.Time) gen.Declaration {
	return gen.Declaration{ID: "__DateTime_json", Content: jsonForTime()}
}

func (buf buffer) codeForArray(typ *an.Array, parentOutputFile string) (gen.Declaration, string) {
	out := gen.Declaration{ID: jsonID(typ), Content: jsonForArray(typ)}

	// recurse for the element
	importS := buf.generate(typ.Elem, parentOutputFile)
	return out, importS
}

func (buf buffer) codeForMap(typ *an.Map, parentOutputFile string) (gen.Declaration, string, string) {
	out := gen.Declaration{ID: jsonID(typ), Content: jsonForMap(typ)}

	// recurse for the key and element
	importKey := buf.generate(typ.Key, parentOutputFile)
	importElem := buf.generate(typ.Elem, parentOutputFile)

	return out, importKey, importElem
}

func codeForEnum(typ *an.Enum) gen.Declaration {
	var names, values, comments []string
	for _, v := range typ.Members {
		if !v.Const.Exported() {
			continue
		}
		names = append(names, lowerFirst(v.Const.Name()))
		comments = append(comments, fmt.Sprintf("%q", v.Comment))
		values = append(values, v.Const.Val().String())
	}

	name := typeName(typ)

	var fromValue string
	if typ.IsIota { // we can just use Dart builtin enums
		fromValue = fmt.Sprintf(`static %s fromValue(int i) {
			return %s.values[i];
		}
		
		int toValue() {
			return index;
		}
		`, name, name)
	} else { // add lookup array
		valueType := "String"
		if typ.IsInteger() {
			valueType = "int"
		}
		fromValue = fmt.Sprintf(`
		static const _values = [
			%s
		];
		static %s fromValue(%s s) {
			return %s.values[_values.indexOf(s)];
		}
	
		%s toValue() {
			return _values[index];
		}
		`, strings.Join(values, ", "), name, valueType, name, valueType)
	}

	enumDecl := fmt.Sprintf(`enum  %s {
		%s
	}
	
	extension _%sExt on %s {
		%s
	}
	`, name, strings.Join(names, ", "), name, name, fromValue)

	content := "// " + gen.Origin(typ) + "\n" + enumDecl
	content += "\n" + jsonForEnum(typ)

	return gen.Declaration{ID: name, Content: content}
}

func (buf buffer) codeForUnion(typ *an.Union) (gen.Declaration, []string) {
	var importMembers []string
	// recurse
	for _, member := range typ.Members {
		imp := buf.generate(member, buf.linker.GetOutput(typ.Type()))
		importMembers = append(importMembers, imp)
	}

	name := typeName(typ)
	content := fmt.Sprintf(`
	/// %s
	abstract class %s {}
	`, gen.Origin(typ), name)

	content += jsonForUnion(typ)

	out := gen.Declaration{ID: name, Content: content}
	return out, importMembers
}

func (buf buffer) codeForStruct(typ *an.Struct) (gen.Declaration, []string) {
	var fields, initFields, interpolatedFields, importForFields []string
	for _, field := range typ.Fields {
		if !field.JSONExported() {
			continue
		}

		// recurse
		importField := buf.generate(field.Type, buf.linker.GetOutput(typ.Name))
		importForFields = append(importForFields, importField)

		dartFieldName := lowerFirst(field.JSONName()) // convert to dart convention

		fields = append(fields, fmt.Sprintf("final %s %s;", typeName(field.Type), dartFieldName))
		initFields = append(initFields, fmt.Sprintf("this.%s", dartFieldName))
		interpolatedFields = append(interpolatedFields, fmt.Sprintf("$%s", dartFieldName))
	}

	implements := make([]string, 0, len(typ.Implements))
	for _, imp := range typ.Implements {
		if !imp.IsExported() {
			continue
		}
		implements = append(implements, an.LocalName(imp))
	}
	var implementCode string
	if len(implements) != 0 {
		implementCode = "implements " + strings.Join(implements, ", ")
	}

	name := typeName(typ)

	out := gen.Declaration{
		ID: name, Content: fmt.Sprintf(`
		// %s
		class %s %s {
		%s

		const %s(%s);

		@override
		String toString() {
			return "%s(%s)";
		}
		}
		
		%s
	`, gen.Origin(typ), name, implementCode,
			strings.Join(fields, "\n"), name, strings.Join(initFields, ", "),
			name, strings.Join(interpolatedFields, ", "),
			jsonForStruct(typ),
		),
	}

	return out, importForFields
}

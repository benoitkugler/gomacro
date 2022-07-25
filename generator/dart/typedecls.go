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
		return typ.Name().Obj().Name()
	default:
		panic(an.ExhaustiveTypeSwitch + fmt.Sprintf(": %T", typ))
	}
}

func codeForNamed(typ *an.Named, cache gen.Cache) (out []gen.Declaration) {
	// type wrapper
	code := fmt.Sprintf(`
	// %s
	typedef %s = %s;
	
	`, typ.Name().String(), typeName(typ), typeName(typ.Underlying))
	out = append(out, gen.Declaration{ID: typeName(typ), Content: code})

	// recurse for the underlying code
	out = append(out, generate(typ.Underlying, cache)...)
	return out
}

func codeForBasic(typ *an.Basic) gen.Declaration {
	// nothing to declare, since basic types are predeclared
	return gen.Declaration{ID: typeName(typ) + "_json", Content: jsonForBasic(typ)}
}

func codeForTime(typ *an.Time) gen.Declaration {
	return gen.Declaration{ID: "__DateTime_json", Content: jsonForTime()}
}

func codeForArray(typ *an.Array, cache gen.Cache) (out []gen.Declaration) {
	out = append(out, gen.Declaration{ID: jsonID(typ), Content: jsonForArray(typ)})

	// recurse for the element
	out = append(out, generate(typ.Elem, cache)...)
	return out
}

func codeForMap(typ *an.Map, cache gen.Cache) (out []gen.Declaration) {
	out = append(out, gen.Declaration{ID: jsonID(typ), Content: jsonForMap(typ)})

	// recurse for the key and element
	out = append(out, generate(typ.Key, cache)...)
	out = append(out, generate(typ.Elem, cache)...)
	return out
}

func codeForExtern(typ *an.Extern) gen.Declaration {
	importLine := fmt.Sprintf("import '%s';", typ.ExternalFiles["dart"])
	return gen.Declaration{
		ID:       importLine,
		Content:  importLine,
		Priority: true, // must appear before other decls
	}
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

	content := "// " + gen.Origin(typ.Name()) + "\n" + enumDecl
	content += "\n" + jsonForEnum(typ)

	return gen.Declaration{ID: name, Content: content}
}

func codeForUnion(typ *an.Union, cache gen.Cache) (out []gen.Declaration) {
	// recurse
	for _, member := range typ.Members {
		out = append(out, generate(member, cache)...)
	}

	name := typeName(typ)
	content := fmt.Sprintf(`
	/// %s
	abstract class %s {}
	`, gen.Origin(typ.Name()), name)

	content += jsonForUnion(typ)

	out = append(out, gen.Declaration{ID: name, Content: content})
	return out
}

func codeForStruct(typ *an.Struct, cache gen.Cache) (out []gen.Declaration) {
	var fields, initFields, interpolatedFields []string
	for _, field := range typ.Fields {
		if !field.Field.Exported() {
			continue
		}

		// recurse
		out = append(out, generate(field.Type, cache)...)

		dartFieldName := lowerFirst(field.Field.Name()) // convert to dart convention

		fields = append(fields, fmt.Sprintf("final %s %s;", typeName(field.Type), dartFieldName))
		initFields = append(initFields, fmt.Sprintf("this.%s", dartFieldName))
		interpolatedFields = append(interpolatedFields, fmt.Sprintf("$%s", dartFieldName))
	}

	implements := make([]string, len(typ.Implements))
	for i, imp := range typ.Implements {
		implements[i] = imp.Name().Obj().Name()
	}
	var implementCode string
	if len(implements) != 0 {
		implementCode = "implements " + strings.Join(implements, ", ")
	}

	name := typeName(typ)

	decl := gen.Declaration{
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
	`, gen.Origin(typ.Name()), name, implementCode,
			strings.Join(fields, "\n"), name, strings.Join(initFields, ", "),
			name, strings.Join(interpolatedFields, ", "),
			jsonForStruct(typ),
		),
	}
	out = append(out, decl)

	return out
}

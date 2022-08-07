package typescript

import (
	"fmt"
	"strings"

	an "github.com/benoitkugler/gomacro/analysis"
	gen "github.com/benoitkugler/gomacro/generator"
)

// This file defines how to print TS types.

// Generate generates the code for the types in `ana.Source`.
func Generate(ana *an.Analysis) []gen.Declaration {
	var allTypes []an.Type
	for _, ty := range ana.Source {
		allTypes = append(allTypes, ana.Types[ty])
	}
	return generateTypes(allTypes)
}

func generateTypes(types []an.Type) []gen.Declaration {
	var (
		decls []gen.Declaration
		cache = make(gen.Cache)
	)
	for _, ty := range types {
		decls = append(decls, generate(ty, cache)...)
	}
	return decls
}

func typeName(ty an.Type) string {
	switch ty := ty.(type) {
	case *an.Pointer:
		panic("pointers not handled by Typescript generator")
	case *an.Basic:
		switch ty.Kind() {
		case an.BKString:
			return "string"
		case an.BKInt, an.BKFloat:
			return "number"
		case an.BKBool:
			return "boolean"
		default:
			panic(an.ExhaustiveBasicKindSwitch)
		}
	case *an.Time:
		if ty.IsDate {
			return "Date_"
		}
		return "Time"
	case *an.Map:
		return fmt.Sprintf("({ [key: %s]: %s } | null)", typeName(ty.Key), typeName(ty.Elem))
	case *an.Array:
		if ty.Len >= 1 { // not nullable
			return fmt.Sprintf("%s[]", typeName(ty.Elem))
		}
		return fmt.Sprintf("( %s[] | null)", typeName(ty.Elem))
	case *an.Extern, *an.Named, *an.Enum, *an.Struct, *an.Union:
		return an.LocalName(ty)
	default:
		panic(an.ExhaustiveTypeSwitch)
	}
}

func generate(ty an.Type, cache gen.Cache) []gen.Declaration {
	if cache.Check(ty) {
		return nil
	}

	switch ty := ty.(type) {
	case *an.Basic: // nothing to do
		return nil
	case *an.Pointer:
		panic("pointers not handled by Typescript generator")
	case *an.Extern:
		return []gen.Declaration{codeForExtern(ty)}
	case *an.Time:
		return []gen.Declaration{codeForTime(ty)}
	case *an.Array:
		return codeForArray(ty, cache)
	case *an.Map:
		return codeForMap(ty, cache)
	case *an.Named:
		return codeForNamed(ty, cache)
	case *an.Enum:
		return []gen.Declaration{codeForEnum(ty)}
	case *an.Struct:
		return codeForStruct(ty, cache)
	case *an.Union:
		return codeForUnion(ty, cache)
	default:
		panic(an.ExhaustiveTypeSwitch)
	}
}

var (
	timeDecl = gen.Declaration{
		ID: "__time_def",
		Content: `
	class TimeTag {
		private _ :"T" = "T"
	}
	
	// ISO date-time string
	export type Time = string & TimeTag
	`,
	}

	dateDecl = gen.Declaration{
		ID: "__date_def",
		Content: `
	class DateTag {
		private _ :"D" = "D"
	}
	
	// AAAA-MM-YY date format
	export type Date_ = string & DateTag
	`,
	}
)

func codeForTime(t *an.Time) gen.Declaration {
	// special case for date and time
	if t.IsDate {
		return dateDecl
	}
	return timeDecl
}

func codeForExtern(ty *an.Extern) gen.Declaration {
	// make sure .ts is stripped if provied
	extern := strings.TrimSuffix(ty.ExternalFiles["ts"], ".ts")
	importLine := fmt.Sprintf("import {%s} from %q;", an.LocalName(ty), extern)
	return gen.Declaration{
		ID:       importLine,
		Content:  importLine,
		Priority: true, // must appear before other decls
	}
}

func codeForNamed(named *an.Named, cache gen.Cache) []gen.Declaration {
	deps := generate(named.Underlying, cache) // recurse

	code := fmt.Sprintf(`// %s
	export type %s = %s`, gen.Origin(named), typeName(named), typeName(named.Underlying))

	deps = append(deps, gen.Declaration{ID: typeName(named), Content: code})
	return deps
}

func codeForMap(ty *an.Map, cache gen.Cache) []gen.Declaration {
	// the map itself has no additional declarations
	return append(generate(ty.Key, cache), generate(ty.Elem, cache)...)
}

func codeForArray(ty *an.Array, cache gen.Cache) []gen.Declaration {
	// the array itself has no additional declarations
	return generate(ty.Elem, cache)
}

func codeForEnum(enum *an.Enum) gen.Declaration {
	name := an.LocalName(enum)
	var valueDefs, valueLabels []string
	for _, val := range enum.Members {
		varName := val.Const.Name()
		valueDefs = append(valueDefs, fmt.Sprintf("%s = %s,", varName, val.Const.Val().String()))
		valueLabels = append(valueLabels, fmt.Sprintf("[%s.%s]: %q,", name, varName, val.Comment))
	}
	return gen.Declaration{
		ID: name,
		Content: fmt.Sprintf(`// %s
			export enum %s {
				%s
			};

			export const %sLabels: { [key in %s]: string } = {
				%s
			};
			`, gen.Origin(enum), name, strings.Join(valueDefs, "\n"),
			name, name, strings.Join(valueLabels, "\n"),
		),
	}
}

func codeForStruct(t *an.Struct, cache gen.Cache) (decls []gen.Declaration) {
	var fields []string
	for _, field := range t.Fields {
		if !field.JSONExported() {
			continue
		}

		decls = append(decls, generate(field.Type, cache)...) // recurse
		fields = append(fields, fmt.Sprintf("\t%s: %s,", field.JSONName(), typeName(field.Type)))
	}

	name := typeName(t)

	out := "// " + gen.Origin(t) + "\n"
	if isEmpty := len(t.Fields) == 0; isEmpty {
		// TS does not like empty interface
		out += fmt.Sprintf("export type %s = Record<string, never>", name)
	} else { // prefer interface syntax
		out += fmt.Sprintf(`export interface %s {
			%s
		}`, name, strings.Join(fields, "\n"))
	}

	decls = append(decls, gen.Declaration{ID: name, Content: out})
	return decls
}

func codeForUnion(u *an.Union, cache gen.Cache) (out []gen.Declaration) {
	var (
		members  []string
		kindEnum []string
	)
	name := typeName(u)
	enumKindName := name + "Kind"
	for _, m := range u.Members {
		memberName := typeName(m)
		members = append(members, memberName)
		kindEnum = append(kindEnum, fmt.Sprintf("%s = %q", memberName, an.LocalName(m)))

		out = append(out, generate(m, cache)...) // recurse
	}
	code := fmt.Sprintf(`
	export enum %s {
		%s
	}
	
	// %s
	export interface %s {
		Kind: %s
		Data: %s
	}`, enumKindName, strings.Join(kindEnum, ",\n"), gen.Origin(u), name, enumKindName, strings.Join(members, " | "))

	out = append(out, gen.Declaration{ID: name, Content: code})

	return out
}

package dart

import (
	an "github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/generator"
)

// Generate converts the given types to their Dart equivalent,
// also adding JSON convertion functions.
func Generate(an *an.Analysis) []generator.Declaration {
	var (
		out           []generator.Declaration
		generateCache = make(generator.Cache)
	)

	out = append(out, generator.Declaration{
		ID:       "aa_header",
		Content:  "// Code generated by structgen. DO NOT EDIT\n",
		Priority: true,
	}, generator.Declaration{
		ID:      "aa_header_json",
		Content: "typedef JSON = Map<String, dynamic>; // alias to shorten JSON convertors",
	})

	for _, typ := range an.Outline {
		out = append(out, generate(an.Types[typ], generateCache)...)
	}

	return out
}

func generate(typ an.Type, cache generator.Cache) []generator.Declaration {
	if named := typ.Name(); named != nil {
		if cache[named] {
			return nil
		}
		cache[named] = true
	}

	switch typ := typ.(type) {
	case *an.Named:
		return codeForNamed(typ, cache)
	case *an.Basic:
		return []generator.Declaration{codeForBasic(typ)}
	case *an.Time:
		return []generator.Declaration{codeForTime(typ)}
	case *an.Array:
		return codeForArray(typ, cache)
	case *an.Map:
		return codeForMap(typ, cache)
	case *an.Extern:
		return []generator.Declaration{codeForExtern(typ)}
	case *an.Struct:
		return codeForStruct(typ, cache)
	case *an.Enum:
		return []generator.Declaration{codeForEnum(typ)}
	case *an.Union:
		return codeForUnion(typ, cache)
	default:
		panic(an.ExhaustiveTypeSwitch)
	}
}

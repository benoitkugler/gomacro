package sql

import (
	"fmt"
	"go/types"
	"strings"

	an "github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/analysis/sql"
	gen "github.com/benoitkugler/gomacro/generator"
)

func jsonValidations(ty sql.JSON) ([]gen.Declaration, string) {
	cache := make(gen.Cache)
	decls := codeFor(ty.Type(), cache)
	fnName := functionName(ty.Type())
	return decls, fnName
}

func codeFor(ty an.Type, cache gen.Cache) []gen.Declaration {
	if cache.Check(ty) {
		return nil
	}

	switch ty := ty.(type) {
	case *an.Basic, *an.Time:
		return []gen.Declaration{codeForBasicOrTime(ty)}
	case *an.Extern:
		panic("Extern types not implemented in SQL JSON generator")
	case *an.Named:
		return codeFor(ty.Underlying, cache)
	case *an.Enum:
		return []gen.Declaration{codeForEnum(ty)}
	case *an.Map:
		return codeForMap(ty, cache)
	case *an.Array:
		return codeForArray(ty, cache)
	case *an.Struct:
		return codeForStruct(ty, cache)
	case *an.Union:
		return codeForUnion(ty, cache)
	default:
		panic(an.ExhaustiveTypeSwitch)
	}
}

func nameFromKind(kind an.BasicKind) string {
	switch kind {
	case an.BKBool:
		return "boolean"
	case an.BKInt, an.BKFloat:
		return "number"
	case an.BKString:
		return "string"
	default:
		panic(an.ExhaustiveBasicKindSwitch)
	}
}

// typeID returns an identifier for `ty`
// usable in function names
func typeID(ty an.Type) string {
	switch ty := ty.(type) {
	case *an.Pointer:
		panic("pointers not handled by the SQL generator")
	case *an.Basic: // may be named or not
		return nameFromKind(ty.Kind())
	case *an.Time:
		return "string" // saved as ISO string
	case *an.Array:
		as := "array_"
		if ty.Len >= 0 {
			as += fmt.Sprintf("%d_", ty.Len)
		}
		return as + typeID(ty.Elem)
	case *an.Map:
		return "map_" + typeID(ty.Elem) // JSON map keys are always strings
	case *an.Named, *an.Extern, *an.Struct, *an.Enum, *an.Union: // these types are always named
		return idFromNamed(ty.Type().(*types.Named))
	default:
		panic(an.ExhaustiveTypeSwitch + fmt.Sprintf(": %T", ty))
	}
}

func idFromNamed(typ *types.Named) string {
	pkg := typ.Obj().Pkg().Name()
	if len(pkg) > 4 {
		pkg = pkg[:4]
	}
	return pkg + "_" + typ.Obj().Name()
}

// functionName returns the name of the validation function
// associated with `ty`
func functionName(ty an.Type) string {
	return "gomacro_validate_json_" + typeID(ty)
}

const (
	vBasic = `
	CREATE OR REPLACE FUNCTION %s (data jsonb)
		RETURNS boolean
		AS $$
	DECLARE
		is_valid boolean := jsonb_typeof(data) = '%s';
	BEGIN
		IF NOT is_valid THEN 
			RAISE WARNING '%% is not a %s', data;
		END IF;
		RETURN is_valid;
	END;
	$$
	LANGUAGE 'plpgsql'
	IMMUTABLE;`
)

// ty should be Basic or Time
func codeForBasicOrTime(ty an.Type) gen.Declaration {
	name := typeID(ty)
	s := gen.Declaration{
		ID:      functionName(ty),
		Content: fmt.Sprintf(vBasic, functionName(ty), name, name),
	}
	return s
}

const vEnum = `
CREATE OR REPLACE FUNCTION %s (data jsonb)
	RETURNS boolean
	AS $$
DECLARE
	is_valid boolean := jsonb_typeof(data) = '%s' AND %s IN %s;
BEGIN
	IF NOT is_valid THEN 
		RAISE WARNING '%% is not a %s', data;
	END IF;
	RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;`

func codeForEnum(ty *an.Enum) gen.Declaration {
	s := gen.Declaration{ID: functionName(ty)}
	typeCast := `data#>>'{}'`
	if ty.IsInteger() {
		typeCast = "data::int"
	}
	kind, _ := an.NewBasicKind(ty.Underlying().Info())
	s.Content = fmt.Sprintf(vEnum, functionName(ty), nameFromKind(kind), typeCast, enumTuple(ty), typeID(ty))
	return s
}

const vArray = `
	CREATE OR REPLACE FUNCTION %s (data jsonb)
		RETURNS boolean
		AS $$
	BEGIN
		%s
		IF jsonb_typeof(data) != 'array' THEN RETURN FALSE; END IF;
		%s 
		RETURN (SELECT bool_and( %s(value) )  FROM jsonb_array_elements(data)) 
			%s;
	END;
	$$
	LANGUAGE 'plpgsql'
	IMMUTABLE;`

func codeForArray(ty *an.Array, cache gen.Cache) []gen.Declaration {
	critereLength, acceptZeroLength := "", ""
	if ty.Len >= 0 {
		critereLength = fmt.Sprintf("AND jsonb_array_length(data) = %d", ty.Len)
	} else {
		acceptZeroLength = "IF jsonb_array_length(data) = 0 THEN RETURN TRUE; END IF;"
	}
	gardNull := ""
	if ty.Len == -1 { // accepts null
		gardNull = "IF jsonb_typeof(data) = 'null' THEN RETURN TRUE; END IF;"
	}

	out := codeFor(ty.Elem, cache) // recursion

	fn, elemFuncName := functionName(ty), functionName(ty.Elem)
	content := fmt.Sprintf(vArray, fn, gardNull, acceptZeroLength, elemFuncName, critereLength)
	out = append(out, gen.Declaration{ID: fn, Content: content})

	return out
}

const vMap = `
	CREATE OR REPLACE FUNCTION %s (data jsonb)
		RETURNS boolean
		AS $$
	BEGIN
		IF jsonb_typeof(data) = 'null' THEN -- accept null value coming from nil maps 
			RETURN TRUE;
		END IF;
		RETURN jsonb_typeof(data) = 'object'
			AND (SELECT bool_and( %s(value) ) FROM jsonb_each(data));
	END;
	$$
	LANGUAGE 'plpgsql'
	IMMUTABLE;`

func codeForMap(ty *an.Map, cache gen.Cache) []gen.Declaration {
	out := codeFor(ty.Elem, cache) // recursion
	fn, elemFuncName := functionName(ty), functionName(ty.Elem)
	content := fmt.Sprintf(vMap, fn, elemFuncName)

	out = append(out, gen.Declaration{ID: fn, Content: content})
	return out
}

const vStruct = `
	CREATE OR REPLACE FUNCTION %s (data jsonb)
		RETURNS boolean
		AS $$
	DECLARE 
		is_valid boolean;
	BEGIN
		IF jsonb_typeof(data) != 'object' THEN 
			RETURN FALSE;
		END IF;
		is_valid := (SELECT bool_and( 
			%s
		) FROM jsonb_each(data))  
		%s;

		RETURN is_valid;
	END;
	$$
	LANGUAGE 'plpgsql'
	IMMUTABLE;`

func codeForStruct(ty *an.Struct, cache gen.Cache) (out []gen.Declaration) {
	var keys, checks []string
	for _, f := range ty.Fields {
		if !f.JSONExported() {
			continue
		}

		out = append(out, codeFor(f.Type, cache)...) // recursion
		fieldName := f.JSONName()
		keys = append(keys, fmt.Sprintf("'%s'", fieldName))
		checks = append(checks, fmt.Sprintf("AND %s(data->'%s')", functionName(f.Type), fieldName))
	}
	keyList := "key IN (" + strings.Join(keys, ", ") + ")"
	if len(keys) == 0 {
		keyList = "TRUE"
	}
	checkList := strings.Join(checks, "\n")
	fn := functionName(ty)
	content := fmt.Sprintf(vStruct, fn, keyList, checkList)
	out = append(out, gen.Declaration{ID: fn, Content: content})

	return out
}

const vUnion = `
	CREATE OR REPLACE FUNCTION %s (data jsonb)
		RETURNS boolean
		AS $$
	BEGIN
		IF jsonb_typeof(data) != 'object' OR jsonb_typeof(data->'Kind') != 'string' OR jsonb_typeof(data->'Data') = 'null' THEN 
			RETURN FALSE;
		END IF;
		CASE 
			%s
		END CASE;
	END;
	$$
	LANGUAGE 'plpgsql'
	IMMUTABLE;`

func codeForUnion(ty *an.Union, cache gen.Cache) (out []gen.Declaration) {
	var cases []string
	for _, member := range ty.Members {
		kind := member.Type().(*types.Named).Obj().Name()
		cases = append(cases, fmt.Sprintf("WHEN data->>'Kind' = '%s' THEN \n RETURN %s(data->'Data');", kind, functionName(member)))

		// generate validation function for members
		out = append(out, codeFor(member, cache)...)
	}

	cases = append(cases, "ELSE RETURN FALSE;") // unknown Kind type
	fn := functionName(ty)
	out = append(out, gen.Declaration{
		ID:      fn,
		Content: fmt.Sprintf(vUnion, fn, strings.Join(cases, "\n")),
	})
	return out
}

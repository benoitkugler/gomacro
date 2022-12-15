package sql

import (
	"fmt"
	"go/types"

	an "github.com/benoitkugler/gomacro/analysis"
)

const ExhaustiveSQLTypeSwitch = "ExhaustiveSQLTypeSwitch"

// this file defines SQL helpers types
// to handle the Go -> SQL type convertion
// for instance a []byte is not an SQL array, but 'bytea'
// or sql.NullInt64 is not a struct but a (nullable) integer

type Type interface {
	Type() an.Type

	// Name returns the SQL description of the type
	Name() string
}

func (ty Builtin) Type() an.Type { return ty.t }
func (ty Enum) Type() an.Type    { return ty.E }
func (ty Array) Type() an.Type   { return ty.A }
func (ty JSON) Type() an.Type    { return ty.t }

func basicTypeName(kind an.BasicKind) string {
	switch kind {
	case an.BKBool:
		return "boolean"
	case an.BKInt:
		return "integer"
	case an.BKFloat:
		return "real"
	case an.BKString:
		return "text"
	default:
		panic(an.ExhaustiveBasicKindSwitch)
	}
}

// return true for []byte
func isBinary(ty *an.Array) bool {
	bas, ok := ty.Elem.(*an.Basic)
	return ok && bas.B.Kind() == types.Byte
}

// Builtin is a SQL type which does not require
// additional definitions
type Builtin struct {
	t    an.Type
	name string
}

func (b Builtin) Name() string { return b.name }

// IsNullable returns `true` if the type comes
// from a sql.NullXXX type.
func (b Builtin) IsNullable() bool {
	named, ok := b.t.Type().(*types.Named)
	return ok && IsNullXXX(named) != nil
}

type Enum struct {
	E *an.Enum
}

func (e Enum) Name() string {
	kind, _ := an.NewBasicKind(e.E.Underlying().Info())
	return basicTypeName(kind)
}

type Array struct {
	A *an.Array // Elem is *Basic
}

func (ar Array) Name() string {
	return fmt.Sprintf("%s[]", basicTypeName(ar.A.Elem.(*an.Basic).Kind()))
}

type JSON struct {
	t an.Type
}

func (JSON) Name() string { return "jsonb" }

// newType converts a Go type to its matching
// SQL one.
func newType(ty an.Type) Type {
	switch ty := ty.(type) {
	case *an.Basic:
		return Builtin{t: ty, name: basicTypeName(ty.Kind())}
	case *an.Time:
		if ty.IsDate {
			return Builtin{t: ty, name: "date"}
		}
		return Builtin{t: ty, name: "timestamp (0) with time zone"}
	case *an.Array:
		// special case for []byte
		if isBinary(ty) {
			return Builtin{t: ty, name: "bytea"}
		}
		// check if the elem is a basic type, else use JSON
		if _, isElemBasic := ty.Elem.(*an.Basic); isElemBasic {
			return Array{A: ty}
		}
		return JSON{t: ty}
	case *an.Enum:
		return Enum{E: ty}
	case *an.Map, *an.Union: // use the general JSON type
		return JSON{t: ty}
	case *an.Struct:
		// special case for NullXXX types
		if elem := IsNullXXX(ty.Name); elem != nil {
			if basic, isBasic := elem.Type().Underlying().(*types.Basic); isBasic {
				if kind, ok := an.NewBasicKind(basic.Info()); ok {
					return Builtin{t: ty, name: basicTypeName(kind)}
				}
			}
		}
		return JSON{t: ty}
	case *an.Pointer:
		panic("Pointer types not supported in SQL generator")
	case *an.Named:
		return newType(ty.Underlying)
	default:
		panic(an.ExhaustiveTypeSwitch)
	}
}

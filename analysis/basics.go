package analysis

import (
	"go/types"
	"strings"
)

// AnonymousType are the types which may be
// seen without associated names.
// Contrary to the Go language, this package
// does not support anonymous structs, enums and unions.
type AnonymousType interface {
	Type

	isAnonymous()
}

func (*Basic) isAnonymous() {}
func (*Time) isAnonymous()  {}
func (*Array) isAnonymous() {}
func (*Map) isAnonymous()   {}

func (*Basic) Name() *types.Named { return nil }
func (*Time) Name() *types.Named  { return nil }
func (*Array) Name() *types.Named { return nil }
func (*Map) Name() *types.Named   { return nil }

// Named is a named type pointing
// to an `AnonymousType`. Structs, enums,
// and unions are NOT `Named`
type Named struct {
	name       *types.Named
	Underlying AnonymousType
}

func (na *Named) Name() *types.Named { return na.name }

var (
	String = &Basic{B: types.Typ[types.String]}
	Bool   = &Basic{B: types.Typ[types.Bool]}
	Int    = &Basic{B: types.Typ[types.Int]}
	Float  = &Basic{B: types.Typ[types.Float64]}
)

// BasicKind is a simplified information of the kind
// of a basic type, typically shared by Dart, JSON and TypeScript generator
type BasicKind uint8

const (
	BKString BasicKind = iota
	BKInt
	BKFloat
	BKBool
)

// Basic represents all simple types.
// Enums are special cased, meaning they are not of type `Basic`.
type Basic struct {
	B *types.Basic
}

func (b *Basic) Kind() BasicKind {
	info := b.B.Underlying().(*types.Basic).Info()
	if info&types.IsBoolean != 0 {
		return BKBool
	} else if info&types.IsInteger != 0 {
		return BKInt
	} else if info&types.IsFloat != 0 {
		return BKFloat
	} else if info&types.IsString != 0 {
		return BKString
	} else {
		panic("unsupported basic kind")
	}
}

// Time is a special case for *time.Time,
// which is not handled like a regular "pointer to named struct"
type Time struct {
	// IsDate is true if only year/month/day are actually
	// of interest for this type.
	// The following heuristic is used to compute it :
	//	- a type containing Date in its name (case insensitive match) will have IsDate = true
	IsDate bool
}

var (
	timeT = &Time{}
	dateT = &Time{IsDate: true}
)

// newTime returns `true` is the underlying type
// of `typ` is time.Time
func newTime(typ types.Type) (Type, bool) {
	// since we can't acces the underlying name of named types,
	// we check against this string, to detect time.Time
	const timeString = "struct{wall uint64; ext int64; loc *time.Location}"

	if typ.Underlying().String() != timeString {
		return nil, false
	}
	name, isNamed := typ.(*types.Named)
	isDate := isNamed && strings.Contains(strings.ToLower(name.Obj().Name()), "date")
	isCustomNamed := name.Obj().Pkg().Path() != "time"
	out := timeT
	if isDate {
		out = dateT
	}
	if isCustomNamed {
		return &Named{name: name, Underlying: out}, true
	}
	return out, true
}

// Extern is a special type used when
// no declaration should be written for a type,
// but the type should be imported from an other file.
type Extern struct {
	name *types.Named

	// the files where to find the definition,
	// depending on the generation mode
	ExternalFiles map[string]string
}

func (e *Extern) Name() *types.Named { return e.name }

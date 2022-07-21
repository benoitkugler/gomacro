package analysis

import (
	"go/types"
	"strings"
)

// Basic represents all simple types.
// Enums are special cased, meaning they are not of type `Basic`.
type Basic struct {
	typ types.Type // named or not
}

func (b *Basic) Name() *types.Named {
	named, _ := b.typ.(*types.Named)
	return named
}

// func (b *Basic) Type() types.Type {}

type Time struct {
	name *types.Named // optional

	// IsDate is true if only year/month/day are actually
	// of interest for this type.
	// The following heuristic is used to compute it :
	//	- a type containing Date in its name (case insensitive match) will have IsDate = true
	IsDate bool
}

// newTime returns `true` is the underlying type
// of `typ` is time.Time
func newTime(typ types.Type) (*Time, bool) {
	// since we can't acces the underlying name of named types,
	// we check against this string, to detect time.Time
	const timeString = "struct{wall uint64; ext int64; loc *time.Location}"

	if typ.Underlying().String() != timeString {
		return nil, false
	}
	name, isNamed := typ.(*types.Named)
	isDate := isNamed && strings.Contains(strings.ToLower(name.Obj().Name()), "date")
	return &Time{name: name, IsDate: isDate}, true
}

func (ti *Time) Name() *types.Named { return ti.name }

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

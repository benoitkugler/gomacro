package analysis

import "go/types"

// Basic represents all simple types.
// Enums are special cased
type Basic struct {
	goType types.Type
}

func (b *Basic) Name() *types.Named {
	named, _ := b.goType.(*types.Named)
	return named
}
func (b *Basic) Underlying() types.Type { return b.goType.Underlying() }

// TODO:
type Time struct {
	named *types.Named // optional
}

// since we can't acces the underlying name of named types,
// we check against this string, to detect time.Time
const timeString = "struct{wall uint64; ext int64; loc *time.Location}"

// isUnderlyingTime returns `true` is the underlying type
// of `typ` is time.Time
func isUnderlyingTime(typ types.Type) bool {
	return typ.Underlying().String() == timeString
}

// Extern is a special type used when
// no declaration should be written for a type,
// but the type should be imported from an other file.
type Extern struct {
	name *types.Named
	// the file where to find the definition
	ImportFile string
}

func (e *Extern) Name() *types.Named     { return e.name }
func (e *Extern) Underlying() types.Type { return e.name.Underlying() }

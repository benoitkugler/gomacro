package test

import "github.com/benoitkugler/gomacro/testutils/testsource"

type optionalID struct {
	Valid bool
	ID    RepasID
}

type Composite struct {
	A int
	B uint8
	C testsource.EnumUInt
}

// optional with generic

type optGeneric[T ~int64] struct {
	Valid bool
	ID    T
}

type IdQuestion int64

type defined optGeneric[IdQuestion]

type OptAlias = defined

type Advance [10]testsource.EnumUInt

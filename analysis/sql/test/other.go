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

package testsource

import "github.com/benoitkugler/gomacro/testutils/testsource/subpackage"

type Enum int // test it does not collide with subpackage.Enum

const (
	A2 Enum = iota
)

// this is represented by int values,
// but actually has only positive values
type EnumInt int

const (
	Ai EnumInt = iota // sdsd
	Bi                // sdsdB
	Ci                // sdsdC
	_                 // non exported
	Di                // sdsdD
)

// this is represented by int values,
// and has negative values
type enumOptionalBool int

const (
	Yes enumOptionalBool = (-1 + iota)
	No
	Maybe
)

type EnumUInt uint

const (
	A EnumUInt = iota // sdsd
	B                 // sdsdB
	C                 // sdsdC
	D                 // sdsdD
	e                 // not added
	_                 // not exported
)

type enumString string

const (
	SA enumString = "va" // sddA
	SB enumString = "vb" // sddB
	SC enumString = "vc" // sddC
	SD enumString = "vd" // sddD
)

const (
	Value = 1 // not named const are ignored
)

type a struct {
	value subpackage.Enum
}

type notAnEnum string

const SpecialValue notAnEnum = "dummy" // gomacro:no-enum

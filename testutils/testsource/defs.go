package testsource

import (
	"time"

	"github.com/benoitkugler/gomacro/testutils/testsource/subpackage"
)

type ConcretType1 struct {
	List2 []int
	V     int
}

// A regular comment
// gomacro:SQL special sql comment
// gomacro:SQL another special sql comment
type ConcretType2 struct {
	D float64
}

type ItfType interface {
	isI()
}

type ItfType2 interface {
	isI2()
}

func (ConcretType1) isI() {}
func (ConcretType2) isI() {}

func (ConcretType1) isI2() {}

var (
	_ ItfType = ConcretType1{}
	_ ItfType = ConcretType2{}

	_ ItfType2 = ConcretType1{}
)

type MyDate time.Time

type ComplexStruct struct {
	DictWithTag map[int]int `json:"with_tag"`
	NoJSON      EnumInt     `json:"-"`
	u           *int
	Time        time.Time
	B           string
	Value       ItfType
	L           ItfList
	A           int
	E           EnumInt
	E2          EnumUInt
	e3          enumString
	Date        MyDate
	F           [5][5]bool
	Imported    subpackage.StructWithComment
	EnumMap     map[EnumInt]bool
	OptID1      Generic[IdCamp]
	OptID2      Generic[IdFile]
}

type ItfList []ItfType

type StructWithExternalRef struct {
	Field1 subpackage.NamedSlice
	Field2 subpackage.NamedSlice
	Field3 int
}

type RecursiveType struct {
	Children []RecursiveType
}

type (
	Basic1 int
	Basic2 bool
	Basic3 float64
	Basic4 string
)

type WithOpaque struct {
	F1 StructWithExternalRef `gomacro-opaque:"dart"`
	F2 RecursiveType         `gomacro-opaque:"dart, typescript"`
	F3 StructWithExternalRef `gomacro-opaque:" typescript"`
}

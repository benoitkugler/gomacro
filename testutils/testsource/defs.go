package testsource

import (
	"context"
	"math/big"
	"time"

	"github.com/benoitkugler/gomacro/testutils/testsource/subpackage"
)

type concretType1 struct {
	List2 []int
	V     int
}

// A regular comment
// gomacro:SQL special sql comment
// gomacro:SQL another special sql comment
type concretType2 struct {
	D float64
}

type itfType interface {
	isI()
}

type itfType2 interface {
	isI2()
}

func (concretType1) isI() {}
func (concretType2) isI() {}

func (concretType1) isI2() {}

var (
	_ itfType = concretType1{}
	_ itfType = concretType2{}

	_ itfType2 = concretType1{}
)

type MyDate time.Time

type complexStruct struct {
	Dict     map[int]int
	U        *int
	Time     *time.Time
	B        string
	Value    itfType
	L        itfList
	A        int
	E        enumInt
	Date     MyDate
	F        [5]int
	Imported subpackage.StructWithComment
}

type itfList []itfType

type structWithExternalRef struct {
	Field1 context.Context    `gomacro-extern:"context:dart:extern.dart"`
	Field2 context.CancelFunc `gomacro-extern:"context:dart:extern2.dart:ts:extern.ts"`
	Field3 map[int]big.Rat    `gomacro-extern:"big:dart:extern3.dart"`
}

type recursiveType struct {
	Children []recursiveType
}

type notAnEnum string

const SpecialValue notAnEnum = "dummy" // gomacro:no-enum

type withEmbeded struct {
	notAnEnum

	complexStruct
}

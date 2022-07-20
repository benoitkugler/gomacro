package testsource

import (
	"context"
	"math/big"
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

type complexStruct struct {
	Dict  map[int]int
	B     string
	Value itfType
	L     itfList
	A     int
}

type itfList []itfType

type structWithExternalRef struct {
	Field1 context.Context `dart-extern:"context:extern.dart"`
	Field2 context.Context `dart-extern:"context:extern.dart"`
	Field3 map[int]big.Rat `dart-extern:"big:extern.dart"`
}

type recursiveType struct {
	Children []recursiveType
}

type notAnEnum string

const SpecialValue notAnEnum = "dummy" // gomacro:no-enum

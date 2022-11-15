package test

import (
	"math/rand"
	"time"

	"github.com/benoitkugler/gomacro/testutils/testsource"
	"github.com/benoitkugler/gomacro/testutils/testsource/subpackage"
)

// Code generated by gomacro/generator/go/randdata. DO NOT EDIT.

func randArray5int() [5]int {
	var out [5]int
	for i := range out {
		out[i] = randint()
	}
	return out
}

func randMapintint() map[int]int {
	l := 40 + rand.Intn(10)
	out := make(map[int]int, l)
	for i := 0; i < l; i++ {
		out[randint()] = randint()
	}
	return out
}

func randSliceint() []int {
	l := 40 + rand.Intn(10)
	out := make([]int, l)
	for i := range out {
		out[i] = randint()
	}
	return out
}

func randSlicetes_ItfType() []testsource.ItfType {
	l := 40 + rand.Intn(10)
	out := make([]testsource.ItfType, l)
	for i := range out {
		out[i] = randtes_ItfType()
	}
	return out
}

func randSlicetes_RecursiveType() []testsource.RecursiveType {
	l := 40 + rand.Intn(10)
	out := make([]testsource.RecursiveType, l)
	for i := range out {
		out[i] = randtes_RecursiveType()
	}
	return out
}

func randbool() bool {
	i := rand.Int31n(2)
	return i == 1
}

func randfloat64() float64 {
	return rand.Float64() * float64(rand.Int31())
}

func randint() int {
	return int(rand.Intn(1000000))
}

var letterRunes2 = []rune("azertyuiopqsdfghjklmwxcvbn123456789é@!?&èïab ")

func randstring() string {
	b := make([]rune, 50)
	maxLength := len(letterRunes2)
	for i := range b {
		b[i] = letterRunes2[rand.Intn(maxLength)]
	}
	return string(b)
}

func randsub_StructWithComment() subpackage.StructWithComment {
	return subpackage.StructWithComment{
		A: randint(),
	}
}

func randtTime() time.Time {
	return time.Unix(int64(rand.Int31()), 5)
}

func randtes_Basic1() testsource.Basic1 {
	return testsource.Basic1(randint())
}

func randtes_Basic2() testsource.Basic2 {
	return testsource.Basic2(randbool())
}

func randtes_Basic3() testsource.Basic3 {
	return testsource.Basic3(randfloat64())
}

func randtes_Basic4() testsource.Basic4 {
	return testsource.Basic4(randstring())
}

func randtes_ComplexStruct() testsource.ComplexStruct {
	return testsource.ComplexStruct{
		DictWithTag: randMapintint(),
		NoJSON:      randtes_EnumInt(),
		Time:        randtTime(),
		B:           randstring(),
		Value:       randtes_ItfType(),
		L:           randtes_ItfList(),
		A:           randint(),
		E:           randtes_EnumInt(),
		E2:          randtes_EnumUInt(),
		Date:        randtes_MyDate(),
		F:           randArray5int(),
		Imported:    randsub_StructWithComment(),
	}
}

func randtes_ConcretType1() testsource.ConcretType1 {
	return testsource.ConcretType1{
		List2: randSliceint(),
		V:     randint(),
	}
}

func randtes_ConcretType2() testsource.ConcretType2 {
	return testsource.ConcretType2{
		D: randfloat64(),
	}
}

func randtes_EnumInt() testsource.EnumInt {
	choix := [...]testsource.EnumInt{testsource.Ai, testsource.Bi, testsource.Ci, testsource.Di}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randtes_EnumUInt() testsource.EnumUInt {
	choix := [...]testsource.EnumUInt{testsource.A, testsource.B, testsource.C, testsource.D}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randtes_ItfList() testsource.ItfList {
	return testsource.ItfList(randSlicetes_ItfType())
}

func randtes_ItfType() testsource.ItfType {
	choix := [...]testsource.ItfType{
		randtes_ConcretType1(),
		randtes_ConcretType2(),
	}
	i := rand.Intn(2)
	return choix[i]
}

func randtes_ItfType2() testsource.ItfType2 {
	choix := [...]testsource.ItfType2{
		randtes_ConcretType1(),
	}
	i := rand.Intn(1)
	return choix[i]
}

func randtes_MyDate() testsource.MyDate {
	return testsource.MyDate(randtTime())
}

func randtes_RecursiveType() testsource.RecursiveType {
	return testsource.RecursiveType{
		Children: randSlicetes_RecursiveType(),
	}
}

func randtes_StructWithExternalRef() testsource.StructWithExternalRef {
	return testsource.StructWithExternalRef{
		Field2: randsub_NamedSlice(),
	}
}

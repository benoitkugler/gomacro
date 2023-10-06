package analysis

import (
	"go/types"
	"testing"

	. "github.com/benoitkugler/gomacro/testutils"
)

func TestImplements(t *testing.T) {
	dict := fetchPkgUnions(testPkg)

	concretType1 := Lookup(testPkg, "ConcretType1")
	concretType2 := Lookup(testPkg, "ConcretType2")
	itfType := Lookup(testPkg, "ItfType")
	itfType2 := Lookup(testPkg, "ItfType2")

	an := NewAnalysisFromTypes(testPkg, []types.Type{itfType, itfType2})

	cl1 := Struct{Name: concretType1}
	cl2 := Struct{Name: concretType2}

	cl1.setImplements(dict, an.Types)
	cl2.setImplements(dict, an.Types)

	Assert(t, len(cl1.Implements) == 2)
	Assert(t, len(cl2.Implements) == 1)
}

func TestStructComments(t *testing.T) {
	concretType1 := Lookup(testPkg, "ConcretType1")
	concretType2 := Lookup(testPkg, "ConcretType2")

	c1 := fetchStructComments(testPkg, concretType1)
	c2 := fetchStructComments(testPkg, concretType2)

	Assert(t, len(c1) == 0)
	Assert(t, len(c2) == 2)
}

func TestJSONTag(t *testing.T) {
	structT := Lookup(testPkg, "ComplexStruct")
	an := NewAnalysisFromTypes(testPkg, []types.Type{structT})

	st := an.Types[structT].(*Struct)
	Assert(t, st.Fields[0].JSONName() == "with_tag")
	Assert(t, !st.Fields[1].Exported())
}

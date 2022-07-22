package analysis

import (
	"go/types"
	"testing"

	. "github.com/benoitkugler/gomacro/testutils"
)

func TestImplements(t *testing.T) {
	dict := fetchPkgUnions(testPkg)

	concretType1 := testPkg.Types.Scope().Lookup("concretType1").Type().(*types.Named)
	concretType2 := testPkg.Types.Scope().Lookup("concretType2").Type().(*types.Named)
	itfType := testPkg.Types.Scope().Lookup("itfType").Type().(*types.Named)
	itfType2 := testPkg.Types.Scope().Lookup("itfType2").Type().(*types.Named)

	an := NewAnalysisFromTypes(testPkg, []*types.Named{itfType, itfType2})

	cl1 := Struct{name: concretType1}
	cl2 := Struct{name: concretType2}

	cl1.setImplements(dict, an.Types)
	cl2.setImplements(dict, an.Types)

	Assert(t, len(cl1.Implements) == 2)
	Assert(t, len(cl2.Implements) == 1)
}

func TestStructComments(t *testing.T) {
	concretType1 := testPkg.Types.Scope().Lookup("concretType1").Type().(*types.Named)
	concretType2 := testPkg.Types.Scope().Lookup("concretType2").Type().(*types.Named)

	c1 := fetchStructComments(testPkg, concretType1)
	c2 := fetchStructComments(testPkg, concretType2)

	Assert(t, len(c1) == 0)
	Assert(t, len(c2) == 2)
}

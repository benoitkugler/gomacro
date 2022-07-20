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

	cl1 := Class{name: concretType1}
	cl2 := Class{name: concretType2}

	Assert(t, len(cl1.Implements(dict)) == 2)
	Assert(t, len(cl2.Implements(dict)) == 1)

	fetchStructComments(testPkg, concretType1)
}

func TestStructComments(t *testing.T) {
	concretType1 := testPkg.Types.Scope().Lookup("concretType1").Type().(*types.Named)
	concretType2 := testPkg.Types.Scope().Lookup("concretType2").Type().(*types.Named)

	c1 := fetchStructComments(testPkg, concretType1)
	c2 := fetchStructComments(testPkg, concretType2)

	Assert(t, len(c1) == 0)
	Assert(t, len(c2) == 2)
}

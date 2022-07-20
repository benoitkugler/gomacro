package analysis

import (
	"go/types"
	"testing"

	. "github.com/benoitkugler/gomacro/testutils"
)

func TestFetchUnion(t *testing.T) {
	dict := fetchPkgUnions(testPkg)
	if len(dict) != 2 {
		t.Fatal(dict)
	}

	itf := testPkg.Types.Scope().Lookup("itfType").Type().(*types.Named)
	union := dict[itf]

	Assert(t, len(union.Members) == 2)
}

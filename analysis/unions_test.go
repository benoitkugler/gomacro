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

	itf := testPkg.Types.Scope().Lookup("ItfType").Type().(*types.Named)
	members := dict[itf]

	Assert(t, len(members) == 2)
}

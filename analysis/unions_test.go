package analysis

import (
	"testing"

	. "github.com/benoitkugler/gomacro/testutils"
)

func TestFetchUnion(t *testing.T) {
	dict := fetchPkgUnions(testPkg)
	if len(dict) != 2 {
		t.Fatal(dict)
	}

	itf := Lookup(testPkg, "ItfType")
	members := dict[itf]

	Assert(t, len(members) == 2)
}

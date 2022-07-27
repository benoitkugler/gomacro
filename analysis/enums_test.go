package analysis

import (
	"testing"

	. "github.com/benoitkugler/gomacro/testutils"
)

func TestFetchEnums(t *testing.T) {
	dict := fetchPkgEnums(testPkg)
	if len(dict) != 5 {
		t.Fatal(dict)
	}

	enumInt := Lookup(testPkg, "EnumInt")
	enumUInt := Lookup(testPkg, "EnumUInt")
	enumString := Lookup(testPkg, "enumString")
	enumOptionalBool := Lookup(testPkg, "enumOptionalBool")

	Assert(t, dict[enumInt].IsInteger())
	Assert(t, !dict[enumInt].IsIota)
	Assert(t, dict[enumUInt].IsInteger())
	Assert(t, dict[enumUInt].IsIota)
	Assert(t, !dict[enumString].IsInteger())
	Assert(t, dict[enumOptionalBool].IsInteger())
	Assert(t, !dict[enumOptionalBool].IsIota)

	Assert(t, len(dict[enumInt].Members) == 4)
}

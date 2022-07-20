package analysis

import (
	"go/types"
	"testing"

	. "github.com/benoitkugler/gomacro/testutils"
)

func TestFetchEnums(t *testing.T) {
	dict := fetchPkgEnums(testPkg)
	if len(dict) != 5 {
		t.Fatal(dict)
	}

	enumInt := testPkg.Types.Scope().Lookup("enumInt").Type().(*types.Named)
	enumUInt := testPkg.Types.Scope().Lookup("enumUInt").Type().(*types.Named)
	enumString := testPkg.Types.Scope().Lookup("enumString").Type().(*types.Named)
	enumOptionalBool := testPkg.Types.Scope().Lookup("enumOptionalBool").Type().(*types.Named)

	Assert(t, dict[enumInt].IsInteger())
	Assert(t, dict[enumUInt].IsInteger())
	Assert(t, !dict[enumString].IsInteger())
	Assert(t, !dict[enumOptionalBool].IsInteger())

	Assert(t, len(dict[enumInt].Members) == 4)
}

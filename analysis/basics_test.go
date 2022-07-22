package analysis

import (
	"go/types"
	"testing"

	. "github.com/benoitkugler/gomacro/testutils"
)

func TestBasicKind(t *testing.T) {
	basic1 := testPkg.Types.Scope().Lookup("basic1").Type().(*types.Named)
	basic2 := testPkg.Types.Scope().Lookup("basic2").Type().(*types.Named)
	basic3 := testPkg.Types.Scope().Lookup("basic3").Type().(*types.Named)
	basic4 := testPkg.Types.Scope().Lookup("basic4").Type().(*types.Named)

	an := NewAnalysisFromTypes(testPkg, []*types.Named{basic1, basic2, basic3, basic4})

	b1 := an.Types[basic1].(*Named).Underlying.(*Basic)
	b2 := an.Types[basic2].(*Named).Underlying.(*Basic)
	b3 := an.Types[basic3].(*Named).Underlying.(*Basic)
	b4 := an.Types[basic4].(*Named).Underlying.(*Basic)

	Assert(t, b1.Kind() == BKInt)
	Assert(t, b2.Kind() == BKBool)
	Assert(t, b3.Kind() == BKFloat)
	Assert(t, b4.Kind() == BKString)
}

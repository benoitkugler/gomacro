package analysis

import (
	"fmt"
	"go/types"
	"path/filepath"
	"testing"
	"time"

	. "github.com/benoitkugler/gomacro/testutils"
	"golang.org/x/tools/go/packages"
)

var (
	testPkg    *packages.Package
	testSource string
)

func init() {
	fn := "../testutils/testsource/defs.go"
	ti := time.Now()
	pa, err := loadSource(fn)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Test source code loaded in %s\n", time.Since(ti))
	testPkg = pa

	testSource, err = filepath.Abs(fn)
	if err != nil {
		panic(err)
	}
}

func TestPanics(t *testing.T) {
	pkg := *testPkg
	pkg.Syntax = nil

	ShouldPanic(t, func() { fetchConstComment(&pkg, testPkg.Types.Scope().Lookup("Yes").(*types.Const)) })

	ShouldPanic(t, func() { isSpecialComment("// gomacro:XXX a") })
}

func TestMethodTags(t *testing.T) {
	for _, v := range []Type{
		&Basic{goType: &types.Basic{}},
		&Enum{name: &types.Named{}},
		&Class{},
		&Union{},
		&Map{},
		&Array{},
	} {
		v.Name()
		// v.Underlying()
	}
}

func TestLoadSource(t *testing.T) {
	_, err := loadSource("not existing")
	Assert(t, err != nil)

	_, err = loadSource("../testutils/testsource/not_go/dummy.txt")
	Assert(t, err != nil)
}

func TestFetch(t *testing.T) {
	enums, _ := fetchEnumsAndUnions(testPkg)
	if len(enums) != 6 {
		t.Fatal(enums)
	}
}

func TestAnalysis(t *testing.T) {
	an := newAnalysis(testPkg, testSource)
	fmt.Println(an.Outline)
}

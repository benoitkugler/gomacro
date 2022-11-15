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
	pa, err := LoadSource(fn)
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

	ShouldPanic(t, func() { newExternMap(`gomacro-extern:"context#dart"`) })

	ShouldPanic(t, func() { (&Basic{B: types.Typ[types.Complex128]}).Kind() })

	ShouldPanic(t, func() { (&Analysis{}).createType(nil, context{}) })
	ShouldPanic(t, func() { (&Analysis{}).createType(types.NewStruct(nil, nil), context{}) })
	ShouldPanic(t, func() { (&Analysis{}).createType(types.NewChan(types.RecvOnly, nil), context{}) })
}

func TestMethodTags(t *testing.T) {
	for _, v := range []Type{
		&Basic{B: &types.Basic{}},
		&Time{},
		&Array{&Time{}, 1},
		&Map{&Time{}, &Time{}},
		&Enum{name: &types.Named{}},
		&Struct{},
		&Union{},
		&Extern{Origin: &Named{}},
		&Pointer{&Time{}},
		&Named{},
	} {
		v.Type()
	}
}

func newAnalysisFromFile(sourceFile string) (*Analysis, error) {
	pa, err := LoadSource(sourceFile)
	if err != nil {
		return nil, err
	}

	return NewAnalysisFromFile(pa, sourceFile), nil
}

func TestLoadSource(t *testing.T) {
	_, err := LoadSource("not existing")
	Assert(t, err != nil)

	_, err = LoadSource("../testutils/testsource/not_go/dummy.txt")
	Assert(t, err != nil)

	_, err = newAnalysisFromFile("../testutils/testsource/not_go/dummy.txt")
	Assert(t, err != nil)

	_, err = LoadSources([]string{"analysis.go", "../testutils/utils.go", "basics.go"})
	Assert(t, err == nil)
}

func TestFetch(t *testing.T) {
	enums, _ := fetchEnumsAndUnions(testPkg)
	if len(enums) != 6 {
		t.Fatal(enums)
	}
}

func TestAnalysFromTypes(t *testing.T) {
	st := Lookup(testPkg, "StructWithExternalRef")

	an := NewAnalysisFromTypes(testPkg, []types.Type{st})
	Assert(t, len(an.Source) == 1)
	Assert(t, len(an.Types) > 0)
}

func TestAnalysisStruct(t *testing.T) {
	an := NewAnalysisFromFile(testPkg, testSource)

	st := Lookup(testPkg, "StructWithExternalRef")
	fields := an.Types[st].(*Struct).Fields
	Assert(t, len(fields) == 3)

	ext, ok := fields[0].Type.(*Extern)
	Assert(t, ok)
	Assert(t, len(ext.ExternalFiles) == 1)

	ext, ok = fields[1].Type.(*Extern)
	Assert(t, ok)
	Assert(t, len(ext.ExternalFiles) == 2, ext.ExternalFiles)

	ma, ok := fields[2].Type.(*Map)
	Assert(t, ok)
	ext, ok = ma.Elem.(*Extern)
	Assert(t, ok)
	Assert(t, len(ext.ExternalFiles) == 1)
}

func TestGetByName(t *testing.T) {
	an := NewAnalysisFromFile(testPkg, testSource)

	st := Lookup(testPkg, "StructWithExternalRef")
	fmt.Println(an.GetByName(st, "Basic2").Type().String())
}

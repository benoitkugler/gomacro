package randdata

import (
	"go/types"
	"os"
	"strings"
	"testing"

	"github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/generator"
)

func TestGenerate(t *testing.T) {
	source := "../../../testutils/testsource/defs.go"
	pkg, err := analysis.LoadSource(source)
	if err != nil {
		t.Fatal(err)
	}
	an := analysis.NewAnalysisFromFile(pkg, source)

	testPath := strings.ReplaceAll(an.Pkg.PkgPath, "testutils/testsource", "generator/go/randdata")
	decls := generateWithTarget(an, types.NewPackage(testPath, "test"))
	out := generator.WriteDeclarations(decls)

	fn := "test/data.go"
	err = os.WriteFile(fn, []byte(out), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	var fmts generator.Formatters
	if err := fmts.FormatFile(generator.Go, fn); err != nil {
		t.Fatal(err)
	}
}

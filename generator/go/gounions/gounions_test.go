package gounions

import (
	"os"
	"testing"

	"github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/generator"
)

func TestUnions(t *testing.T) {
	source := "test/test.go"
	pkg, err := analysis.LoadSource(source)
	if err != nil {
		t.Fatal(err)
	}
	an := analysis.NewAnalysisFromFile(pkg, source)

	decls := Generate(an)
	out := generator.WriteDeclarations(decls)

	fn := "test/gen.go"
	err = os.WriteFile(fn, []byte(out), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	var fmts generator.Formatters
	if err := fmts.FormatFile(generator.Go, fn); err != nil {
		t.Fatal(err)
	}
}

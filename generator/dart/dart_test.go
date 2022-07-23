package dart

import (
	"os"
	"testing"

	"github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/generator"
)

func TestGenerate(t *testing.T) {
	an, err := analysis.NewAnalysisFromFile("../../testutils/testsource/defs.go")
	if err != nil {
		t.Fatal(err)
	}

	decls := Generate(an)
	out := generator.WriteDeclarations(decls)

	fn := "test/gen.dart"
	err = os.WriteFile(fn, []byte(out), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	var fmts generator.Formatters
	if err := fmts.FormatFile(generator.Dart, fn); err != nil {
		t.Fatal(err)
	}
}

package sql

import (
	"fmt"
	"os"
	"testing"

	"github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/generator"
)

func TestCreate(t *testing.T) {
	source := "../../analysis/sql/test/models.go"
	pkg, err := analysis.LoadSource(source)
	if err != nil {
		t.Fatal(err)
	}
	an := analysis.NewAnalysisFromFile(pkg, source)

	decls := Generate(an)

	out := generator.WriteDeclarations(decls)
	generated := "test/create.sql"
	err = os.WriteFile(generated, []byte(out), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	var fmts generator.Formatters
	if err := fmts.FormatFile(generator.Psql, generated); err != nil {
		t.Fatal(err)
	}
}

func TestEnums(t *testing.T) {
	content := "ADD CHECK(Kind = #[Kind.value] OR )"
	content = reEnums.ReplaceAllStringFunc(content, func(s string) string {
		return "3"
	})
	fmt.Println(content)
}

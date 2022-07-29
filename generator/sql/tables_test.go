package sql

import (
	"os"
	"testing"

	"github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/analysis/sql"
	"github.com/benoitkugler/gomacro/generator"
)

func TestCreate(t *testing.T) {
	fn := "../../analysis/sql/test/models.go"
	an, err := analysis.NewAnalysisFromFile(fn)
	if err != nil {
		t.Fatal(err)
	}

	var decls []generator.Declaration

	for _, ty := range an.Source {
		st, ok := an.Types[ty].(*analysis.Struct)
		if !ok {
			continue
		}

		decls = append(decls, generateTable(sql.NewTable(st))...)
	}

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

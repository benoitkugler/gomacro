package sqlcrud

import (
	"os"
	"testing"

	"github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/analysis/sql"
	"github.com/benoitkugler/gomacro/generator"
	. "github.com/benoitkugler/gomacro/testutils"
)

var (
	ana     *analysis.Analysis
	fileOut = "../../../analysis/sql/test/crud_gen.go"
)

func init() {
	// automatically remove output file which may not typecheck during dev
	_ = os.Remove(fileOut)
	fn := "../../../analysis/sql/test/models.go"
	pkg, err := analysis.LoadSource(fn)
	if err != nil {
		panic(err)
	}
	ana = analysis.NewAnalysisFromFile(pkg, fn)

	pqImportPath = "github.com/benoitkugler/gomacro/analysis/sql/test/pq"
}

func TestPrintID(t *testing.T) {
	table1 := sql.NewTable(ana.Types[Lookup(ana.Root, "Table1")].(*analysis.Struct))
	repas := sql.NewTable(ana.Types[Lookup(ana.Root, "Repas")].(*analysis.Struct))

	table1ID := table1.Columns[table1.Primary()].Field.Type.Type()
	repasID := repas.Columns[repas.Primary()].Field.Type.Type()

	ctx := context{ana.Root.Types, nil}
	Assert(t, ctx.typeName(table1ID) == "int64")
	Assert(t, ctx.typeName(repasID) == "RepasID")
}

func TestGenerate(t *testing.T) {
	decls := Generate(ana)
	out := generator.WriteDeclarations(decls)

	err := os.WriteFile(fileOut, []byte(out), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	var fmts generator.Formatters
	if err := fmts.FormatFile(generator.Go, fileOut); err != nil {
		t.Fatal(err)
	}
}

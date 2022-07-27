package sql

import (
	"testing"

	"github.com/benoitkugler/gomacro/analysis"
	. "github.com/benoitkugler/gomacro/testutils"
)

func TestSQL(t *testing.T) {
	fn := "test/models.go"
	an, err := analysis.NewAnalysisFromFile(fn)
	if err != nil {
		t.Fatal(err)
	}

	Assert(t, isTableID(an.Types[Lookup(an.Root, "RepasID")]) == "Repas")
	Assert(t, isTableID(an.Types[Lookup(an.Root, "IDInvalid")]) == "")

	table1 := NewTable(an.Types[Lookup(an.Root, "Table1")].(*analysis.Struct))
	Assert(t, len(table1.ForeignKeys()) == 3)
	Assert(t, table1.Primary() == &table1.S.Fields[0])

	repas := NewTable(an.Types[Lookup(an.Root, "Repas")].(*analysis.Struct))
	Assert(t, len(repas.ForeignKeys()) == 0)
	Assert(t, repas.Primary() == &repas.S.Fields[1])

	link := NewTable(an.Types[Lookup(an.Root, "Link")].(*analysis.Struct))
	Assert(t, link.Primary() == nil)

	question := NewTable(an.Types[Lookup(an.Root, "Question")].(*analysis.Struct))
	Assert(t, len(question.ForeignKeys()) == 1)
	Assert(t, question.ForeignKeys()[0].IsNullable())
}

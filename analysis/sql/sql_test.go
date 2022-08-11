package sql

import (
	"testing"

	"github.com/benoitkugler/gomacro/analysis"
	. "github.com/benoitkugler/gomacro/testutils"
)

func TestSQL(t *testing.T) {
	fn := "test/models.go"
	pkg, err := analysis.LoadSource(fn)
	Assert(t, err == nil)

	an := analysis.NewAnalysisFromFile(pkg, fn)

	Assert(t, isTableID(an.Types[Lookup(an.Root, "RepasID")]) == "Repas")
	Assert(t, isTableID(an.Types[Lookup(an.Root, "IDInvalid")]) == "")

	table1 := NewTable(an.Types[Lookup(an.Root, "Table1")].(*analysis.Struct))
	Assert(t, len(table1.ForeignKeys()) == 3)
	Assert(t, table1.Primary() == 0)

	repas := NewTable(an.Types[Lookup(an.Root, "Repas")].(*analysis.Struct))
	Assert(t, len(repas.ForeignKeys()) == 0)
	Assert(t, repas.Primary() == 1)

	link := NewTable(an.Types[Lookup(an.Root, "Link")].(*analysis.Struct))
	Assert(t, link.Primary() == -1)

	question := NewTable(an.Types[Lookup(an.Root, "Question")].(*analysis.Struct))
	Assert(t, len(question.ForeignKeys()) == 1)
	Assert(t, question.ForeignKeys()[0].IsNullable())

	exercicesQuestion := NewTable(an.Types[Lookup(an.Root, "ExerciceQuestion")].(*analysis.Struct))
	Assert(t, len(exercicesQuestion.AdditionalUniqueCols()) == 1)
}

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
	Assert(t, len(table1.ForeignKeys()) == 4)
	Assert(t, table1.Primary() == 0)
	Assert(t, table1.ForeignKeys()[0].TargetIDType() == Lookup(an.Root, "RepasID"))
	Assert(t, table1.ForeignKeys()[2].TargetIDType().String() == "int64")
	Assert(t, table1.ForeignKeys()[3].TargetIDType() == Lookup(an.Root, "RepasID"))

	_, ok := table1.Columns[6].SQLType.(Array)
	Assert(t, ok)

	_, ok = table1.Columns[8].SQLType.(Composite)
	Assert(t, ok)

	_, ok = table1.Columns[9].SQLType.(Array)
	Assert(t, ok)

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

	withOptTime := NewTable(an.Types[Lookup(an.Root, "WithOptionalTime")].(*analysis.Struct))
	Assert(t, len(withOptTime.Columns) == 3)
	bt, ok := withOptTime.Columns[1].SQLType.(Builtin)
	Assert(t, ok)
	Assert(t, !bt.IsNullable())
	bt, ok = withOptTime.Columns[2].SQLType.(Builtin)
	Assert(t, ok)
	Assert(t, bt.IsNullable())

	composite := an.Types[Lookup(an.Root, "Composite")].(*analysis.Struct)
	Assert(t, isComposite(composite))
}

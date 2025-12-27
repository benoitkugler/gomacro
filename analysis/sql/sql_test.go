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

	Assert(t, isTableID(an.Types[Lookup(an.Pkg, "RepasID")]) == "Repas")
	Assert(t, isTableID(an.Types[Lookup(an.Pkg, "IDInvalid")]) == "")

	table1 := NewTable(an.Types[Lookup(an.Pkg, "Table1")].(*analysis.Struct))
	Assert(t, len(table1.ForeignKeys()) == 5)
	Assert(t, table1.Primary() == 0)
	Assert(t, table1.ForeignKeys()[0].TargetIDType() == Lookup(an.Pkg, "RepasID"))
	Assert(t, table1.ForeignKeys()[2].TargetIDType().String() == "int64")
	Assert(t, table1.ForeignKeys()[3].TargetIDType() == Lookup(an.Pkg, "RepasID"))
	Assert(t, table1.ForeignKeys()[4].TargetIDType() == Lookup(an.Pkg, "IdQuestion"))

	_, ok := table1.Columns[6].SQLType.(Array)
	Assert(t, ok)

	_, ok = table1.Columns[8].SQLType.(Composite)
	Assert(t, ok)

	_, ok = table1.Columns[9].SQLType.(Array)
	Assert(t, ok)

	_, ok = table1.Columns[10].Field.IsSQLGuard()
	Assert(t, ok)

	Assert(t, len(table1.CustomQueries) == 2)
	Assert(t, len(table1.CustomQueries[1].Inputs) == 2)

	repas := NewTable(an.Types[Lookup(an.Pkg, "Repas")].(*analysis.Struct))
	Assert(t, len(repas.ForeignKeys()) == 0)
	Assert(t, repas.Primary() == 1)

	link := NewTable(an.Types[Lookup(an.Pkg, "Link")].(*analysis.Struct))
	Assert(t, link.Primary() == -1)

	question := NewTable(an.Types[Lookup(an.Pkg, "Question")].(*analysis.Struct))
	Assert(t, len(question.ForeignKeys()) == 1)
	isNullable, name := question.ForeignKeys()[0].IsNullable()
	Assert(t, isNullable && name == "Int64")

	exercicesQuestion := NewTable(an.Types[Lookup(an.Pkg, "ExerciceQuestion")].(*analysis.Struct))
	Assert(t, len(exercicesQuestion.AdditionalUniqueCols()) == 1)

	withOptTime := NewTable(an.Types[Lookup(an.Pkg, "WithOptionalTime")].(*analysis.Struct))
	Assert(t, len(withOptTime.Columns) == 3)
	bt, ok := withOptTime.Columns[1].SQLType.(Builtin)
	Assert(t, ok)
	Assert(t, !bt.IsNullable())
	bt, ok = withOptTime.Columns[2].SQLType.(Builtin)
	Assert(t, ok)
	Assert(t, bt.IsNullable())

	composite := an.Types[Lookup(an.Pkg, "Composite")].(*analysis.Struct)
	Assert(t, isComposite(composite))
}

func TestCustomQueries(t *testing.T) {
	matches := reCustomQueryFields.FindAllStringSubmatch("UPDATE Participant SET IdPersonne = $v1$ WHERE IdPersonne = $v2$", -1)
	Assert(t, len(matches) == 2)
	Assert(t, matches[0][1] == "IdPersonne")

	matches = reCustomQueryArrayFields.FindAllStringSubmatch("DELETE FROM LettreImage WHERE NOT (Id = ANY($others$));", -1)
	Assert(t, len(matches) == 1)
	Assert(t, matches[0][1] == "Id")
	Assert(t, matches[0][2] == "others")
}

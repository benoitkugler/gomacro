package sql

import (
	"fmt"
	"go/types"
	"os"
	"testing"

	"github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/analysis/sql"
	"github.com/benoitkugler/gomacro/generator"
	"github.com/benoitkugler/gomacro/testutils"
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

func TestReplacer(t *testing.T) {
	rp := tableNameReplacer([]sql.Table{
		{Name: types.NewNamed(types.NewTypeName(0, nil, "Exercice", nil), nil, nil)},
		{Name: types.NewNamed(types.NewTypeName(0, nil, "Question", nil), nil, nil)},
		{Name: types.NewNamed(types.NewTypeName(0, nil, "ExerciceQuestion", nil), nil, nil)},
	})
	testutils.Assert(t, rp.Replace("Exercice") == "exercices")
	testutils.Assert(t, rp.Replace("Question") == "questions")
	testutils.Assert(t, rp.Replace("ExerciceQuestion") == "exercice_questions")
	testutils.Assert(t, rp.Replace("IdExercice") == "IdExercice")
}

func TestEnums(t *testing.T) {
	content := "ADD CHECK(Kind = #[Kind.value] OR )"
	content = reEnums.ReplaceAllStringFunc(content, func(s string) string {
		return "3"
	})
	fmt.Println(content)
}

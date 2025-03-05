package generator

import (
	"fmt"
	"go/types"
	"testing"

	"github.com/benoitkugler/gomacro/analysis/sql"
	"github.com/benoitkugler/gomacro/testutils"
)

func TestReplacer(t *testing.T) {
	rp := NewTableNameReplacer([]sql.Table{
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

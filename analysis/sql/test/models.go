package test

import (
	"database/sql"

	"github.com/benoitkugler/gomacro/testutils/testsource"
)

type RepasID int64

type IDInvalid string

type Table1 struct {
	Id  int64
	Ex1 RepasID
	Ex2 RepasID
	L   sql.NullInt64 `gomacro-sql-foreign:"Link"`
}

type Repas struct {
	Order string
	Id    RepasID
}

type Link struct{}

// Question is a standalone question, used for instance in games.
type Question struct {
	Id          int64                    `json:"id"`
	Page        testsource.ComplexStruct `json:"page"`
	Public      bool                     `json:"public"` // in practice only true for admins
	IdTeacher   int64                    `json:"id_teacher"`
	Description string                   `json:"description"`
	// NeedExercice is not null if the question cannot be instantiated (or edited)
	// on its own
	NeedExercice sql.NullInt64 `json:"need_exercice" gomacro-sql-foreign:"Exercice"`
}

// sql: ADD UNIQUE(id_question, tag)
type QuestionTag struct {
	Tag        string `json:"tag"`
	IdQuestion int64  `sql_on_delete:"CASCADE" json:"id_question"`
}

// DifficultyTag are special question tags used to indicate the
// difficulty of one question.
// It is used to select question among implicit groups
type DifficultyTag string

// Exercice is the data structure for a full exercice, composed of a list of questions.
// There are two kinds of exercice :
//	- parallel : all the questions are independant
//	- progression : the questions are linked together by a shared Parameters set
type Exercice struct {
	Id          int64
	Title       string // displayed to the students
	Description string // used internally by the teachers
	// Parameters are parameters shared by all the questions,
	// which are added to the individual ones.
	// It will be empty for parallel exercices
	Parameters map[string]bool
	Flow       testsource.EnumInt
	// IdTeacher is the owner of the exercice
	IdTeacher int64 `json:"id_teacher"`
	Public    bool
}

// ExerciceQuestion models an ordered list of questions.
// All link items should be updated at once to preserve `Index` invariants
// sql: ADD PRIMARY KEY (id_exercice, index)
type ExerciceQuestion struct {
	IdExercice int64 `json:"id_exercice" sql_on_delete:"CASCADE"`
	IdQuestion int64 `json:"id_question"`
	Bareme     int   `json:"bareme"`
	Index      int   `json:"-" sql:"index"`
}

type (
	IdProgression int64
	IdExercice    int64
)

// Progression is the table storing the student progression
// for one exercice.
// Note that this data structure may also be used in memory,
// for instance for the editor loopback.
// sql: ADD UNIQUE(id, id_exercice)
type Progression struct {
	Id         IdProgression
	IdExercice int64 `json:"id_exercice" sql_on_delete:"CASCADE"`
}

// We enforce consistency with the additional `id_exercice` field
// sql: ADD FOREIGN KEY (id_exercice, index) REFERENCES exercice_questions ON DELETE CASCADE
// sql: ADD FOREIGN KEY (id_progression, id_exercice) REFERENCES progressions (id, id_exercice) ON DELETE CASCADE
type ProgressionQuestion struct {
	IdProgression IdProgression         `json:"id_progression" sql_on_delete:"CASCADE"`
	IdExercice    IdExercice            `json:"id_exercice" sql_on_delete:"CASCADE"`
	Index         int                   `json:"index"` // in the question list
	History       []testsource.EnumUInt `json:"history"`
}

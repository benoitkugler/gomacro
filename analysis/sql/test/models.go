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

type Link struct {
	Repas    RepasID
	IdTable1 int64
}

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

// gomacro:SQL ADD UNIQUE(IdQuestion, Tag)
type QuestionTag struct {
	Tag        string `json:"tag"`
	IdQuestion int64  ` gomacro-sql-foreign:"Question" gomacro-sql-on-delete:"CASCADE" json:"id_question"`
}

// DifficultyTag are special question tags used to indicate the
// difficulty of one question.
// It is used to select question among implicit groups
type DifficultyTag string

type Map map[string]bool

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
	Parameters Map
	Flow       testsource.EnumInt
	// IdTeacher is the owner of the exercice
	IdTeacher int64 `json:"id_teacher"`
	Public    bool
}

// ExerciceQuestion models an ordered list of questions.
// All link items should be updated at once to preserve `Index` invariants
// gomacro:SQL ADD PRIMARY KEY (IdExercice, Index)
type ExerciceQuestion struct {
	IdExercice int64 `json:"id_exercice" gomacro-sql-foreign:"Exercice" gomacro-sql-on-delete:"CASCADE"`
	IdQuestion int64 `json:"id_question" gomacro-sql-foreign:"Question"`
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
// gomacro:SQL ADD UNIQUE(Id, IdExercice)
type Progression struct {
	Id         IdProgression
	IdExercice int64 `json:"id_exercice" gomacro-sql-on-delete:"CASCADE"`
}

type EnumArray []testsource.EnumUInt

// We enforce consistency with the additional `id_exercice` field
// gomacro:SQL ADD UNIQUE(IdProgression, Index)
// gomacro:SQL ADD FOREIGN KEY (IdExercice, Index) REFERENCES exercice_questions ON DELETE CASCADE
// gomacro:SQL ADD FOREIGN KEY (IdProgression, IdExercice) REFERENCES progressions (Id, IdExercice) ON DELETE CASCADE
type ProgressionQuestion struct {
	IdProgression IdProgression `json:"id_progression" gomacro-sql-on-delete:"CASCADE"`
	IdExercice    IdExercice    `json:"id_exercice" gomacro-sql-on-delete:"CASCADE"`
	Index         int           `json:"index"` // in the question list
	History       EnumArray     `json:"history"`
}

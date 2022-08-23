package test

// Code generated by gomacro/generator/go/sqlcrud. DO NOT EDIT.

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/benoitkugler/gomacro/analysis/sql/test/pq"
)

type scanner interface {
	Scan(...interface{}) error
}

// DB groups transaction like objects, and
// is implemented by *sql.DB and *sql.Tx
type DB interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Prepare(query string) (*sql.Stmt, error)
}

func scanOneExercice(row scanner) (Exercice, error) {
	var item Exercice
	err := row.Scan(
		&item.Id,
		&item.Title,
		&item.Description,
		&item.Parameters,
		&item.Flow,
		&item.IdTeacher,
		&item.Public,
	)
	return item, err
}

func ScanExercice(row *sql.Row) (Exercice, error) { return scanOneExercice(row) }

// SelectAll returns all the items in the exercices table.
func SelectAllExercices(db DB) (Exercices, error) {
	rows, err := db.Query("SELECT * FROM exercices")
	if err != nil {
		return nil, err
	}
	return ScanExercices(rows)
}

// SelectExercice returns the entry matching 'id'.
func SelectExercice(tx DB, id int64) (Exercice, error) {
	row := tx.QueryRow("SELECT * FROM exercices WHERE id = $1", id)
	return ScanExercice(row)
}

// SelectExercices returns the entry matching the given 'ids'.
func SelectExercices(tx DB, ids ...int64) (Exercices, error) {
	rows, err := tx.Query("SELECT * FROM exercices WHERE id = ANY($1)", int64ArrayToPQ(ids))
	if err != nil {
		return nil, err
	}
	return ScanExercices(rows)
}

type Exercices map[int64]Exercice

func (m Exercices) IDs() []int64 {
	out := make([]int64, 0, len(m))
	for i := range m {
		out = append(out, i)
	}
	return out
}

func ScanExercices(rs *sql.Rows) (Exercices, error) {
	var (
		s   Exercice
		err error
	)
	defer func() {
		errClose := rs.Close()
		if err == nil {
			err = errClose
		}
	}()
	structs := make(Exercices, 16)
	for rs.Next() {
		s, err = scanOneExercice(rs)
		if err != nil {
			return nil, err
		}
		structs[s.Id] = s
	}
	if err = rs.Err(); err != nil {
		return nil, err
	}
	return structs, nil
}

// Insert one Exercice in the database and returns the item with id filled.
func (item Exercice) Insert(tx DB) (out Exercice, err error) {
	row := tx.QueryRow(`INSERT INTO exercices (
		title, description, parameters, flow, idteacher, public
		) VALUES (
		$1, $2, $3, $4, $5, $6
		) RETURNING *;
		`, item.Title, item.Description, item.Parameters, item.Flow, item.IdTeacher, item.Public)
	return ScanExercice(row)
}

// Update Exercice in the database and returns the new version.
func (item Exercice) Update(tx DB) (out Exercice, err error) {
	row := tx.QueryRow(`UPDATE exercices SET (
		title, description, parameters, flow, idteacher, public
		) = (
		$1, $2, $3, $4, $5, $6
		) WHERE id = $7 RETURNING *;
		`, item.Title, item.Description, item.Parameters, item.Flow, item.IdTeacher, item.Public, item.Id)
	return ScanExercice(row)
}

// Deletes the Exercice and returns the item
func DeleteExerciceById(tx DB, id int64) (Exercice, error) {
	row := tx.QueryRow("DELETE FROM exercices WHERE id = $1 RETURNING *;", id)
	return ScanExercice(row)
}

// Deletes the Exercice in the database and returns the ids.
func DeleteExercicesByIDs(tx DB, ids ...int64) ([]int64, error) {
	rows, err := tx.Query("DELETE FROM exercices WHERE id = ANY($1) RETURNING id", int64ArrayToPQ(ids))
	if err != nil {
		return nil, err
	}
	return Scanint64Array(rows)
}

func scanOneExerciceQuestion(row scanner) (ExerciceQuestion, error) {
	var item ExerciceQuestion
	err := row.Scan(
		&item.IdExercice,
		&item.IdQuestion,
		&item.Bareme,
		&item.Index,
	)
	return item, err
}

func ScanExerciceQuestion(row *sql.Row) (ExerciceQuestion, error) {
	return scanOneExerciceQuestion(row)
}

// SelectAll returns all the items in the exercice_questions table.
func SelectAllExerciceQuestions(db DB) (ExerciceQuestions, error) {
	rows, err := db.Query("SELECT * FROM exercice_questions")
	if err != nil {
		return nil, err
	}
	return ScanExerciceQuestions(rows)
}

type ExerciceQuestions []ExerciceQuestion

func ScanExerciceQuestions(rs *sql.Rows) (ExerciceQuestions, error) {
	var (
		item ExerciceQuestion
		err  error
	)
	defer func() {
		errClose := rs.Close()
		if err == nil {
			err = errClose
		}
	}()
	structs := make(ExerciceQuestions, 0, 16)
	for rs.Next() {
		item, err = scanOneExerciceQuestion(rs)
		if err != nil {
			return nil, err
		}
		structs = append(structs, item)
	}
	if err = rs.Err(); err != nil {
		return nil, err
	}
	return structs, nil
}

// Insert the links ExerciceQuestion in the database.
// It is a no-op if 'items' is empty.
func InsertManyExerciceQuestions(tx *sql.Tx, items ...ExerciceQuestion) error {
	if len(items) == 0 {
		return nil
	}

	stmt, err := tx.Prepare(pq.CopyIn("exercice_questions",
		"idexercice",
		"idquestion",
		"bareme",
		"index",
	))
	if err != nil {
		return err
	}

	for _, item := range items {
		_, err = stmt.Exec(item.IdExercice, item.IdQuestion, item.Bareme, item.Index)
		if err != nil {
			return err
		}
	}

	if _, err = stmt.Exec(); err != nil {
		return err
	}

	if err = stmt.Close(); err != nil {
		return err
	}
	return nil
}

// Delete the link ExerciceQuestion from the database.
// Only the foreign keys IdExercice, IdQuestion fields are used in 'item'.
func (item ExerciceQuestion) Delete(tx DB) error {
	_, err := tx.Exec(`DELETE FROM exercice_questions WHERE IdExercice = $1 AND IdQuestion = $2;`, item.IdExercice, item.IdQuestion)
	return err
}

// ByIdExercice returns a map with 'IdExercice' as keys.
func (items ExerciceQuestions) ByIdExercice() map[int64]ExerciceQuestions {
	out := make(map[int64]ExerciceQuestions)
	for _, target := range items {
		out[target.IdExercice] = append(out[target.IdExercice], target)
	}
	return out
}

// IdExercices returns the list of ids of IdExercice
// contained in this link table.
// They are not garanteed to be distinct.
func (items ExerciceQuestions) IdExercices() []int64 {
	out := make([]int64, len(items))
	for index, target := range items {
		out[index] = target.IdExercice
	}
	return out
}

func SelectExerciceQuestionsByIdExercices(tx DB, idExercices ...int64) (ExerciceQuestions, error) {
	rows, err := tx.Query("SELECT * FROM exercice_questions WHERE idexercice = ANY($1)", int64ArrayToPQ(idExercices))
	if err != nil {
		return nil, err
	}
	return ScanExerciceQuestions(rows)
}

func DeleteExerciceQuestionsByIdExercices(tx DB, idExercices ...int64) (ExerciceQuestions, error) {
	rows, err := tx.Query("DELETE FROM exercice_questions WHERE idexercice = ANY($1) RETURNING *", int64ArrayToPQ(idExercices))
	if err != nil {
		return nil, err
	}
	return ScanExerciceQuestions(rows)
}

// ByIdQuestion returns a map with 'IdQuestion' as keys.
func (items ExerciceQuestions) ByIdQuestion() map[int64]ExerciceQuestions {
	out := make(map[int64]ExerciceQuestions)
	for _, target := range items {
		out[target.IdQuestion] = append(out[target.IdQuestion], target)
	}
	return out
}

// IdQuestions returns the list of ids of IdQuestion
// contained in this link table.
// They are not garanteed to be distinct.
func (items ExerciceQuestions) IdQuestions() []int64 {
	out := make([]int64, len(items))
	for index, target := range items {
		out[index] = target.IdQuestion
	}
	return out
}

func SelectExerciceQuestionsByIdQuestions(tx DB, idQuestions ...int64) (ExerciceQuestions, error) {
	rows, err := tx.Query("SELECT * FROM exercice_questions WHERE idquestion = ANY($1)", int64ArrayToPQ(idQuestions))
	if err != nil {
		return nil, err
	}
	return ScanExerciceQuestions(rows)
}

func DeleteExerciceQuestionsByIdQuestions(tx DB, idQuestions ...int64) (ExerciceQuestions, error) {
	rows, err := tx.Query("DELETE FROM exercice_questions WHERE idquestion = ANY($1) RETURNING *", int64ArrayToPQ(idQuestions))
	if err != nil {
		return nil, err
	}
	return ScanExerciceQuestions(rows)
}

// SelectExerciceQuestionByIdExerciceAndIndex return zero or one item, thanks to a UNIQUE SQL constraint.
func SelectExerciceQuestionByIdExerciceAndIndex(tx DB, idExercice int64, index int) (item ExerciceQuestion, found bool, err error) {
	row := tx.QueryRow("SELECT * FROM exercice_questions WHERE IdExercice = $1 AND Index = $2", idExercice, index)
	item, err = ScanExerciceQuestion(row)
	if err == sql.ErrNoRows {
		return item, false, nil
	}
	return item, true, err
}

func scanOneLink(row scanner) (Link, error) {
	var item Link
	err := row.Scan(
		&item.Repas,
		&item.IdTable1,
	)
	return item, err
}

func ScanLink(row *sql.Row) (Link, error) { return scanOneLink(row) }

// SelectAll returns all the items in the links table.
func SelectAllLinks(db DB) (Links, error) {
	rows, err := db.Query("SELECT * FROM links")
	if err != nil {
		return nil, err
	}
	return ScanLinks(rows)
}

type Links []Link

func ScanLinks(rs *sql.Rows) (Links, error) {
	var (
		item Link
		err  error
	)
	defer func() {
		errClose := rs.Close()
		if err == nil {
			err = errClose
		}
	}()
	structs := make(Links, 0, 16)
	for rs.Next() {
		item, err = scanOneLink(rs)
		if err != nil {
			return nil, err
		}
		structs = append(structs, item)
	}
	if err = rs.Err(); err != nil {
		return nil, err
	}
	return structs, nil
}

// Insert the links Link in the database.
// It is a no-op if 'items' is empty.
func InsertManyLinks(tx *sql.Tx, items ...Link) error {
	if len(items) == 0 {
		return nil
	}

	stmt, err := tx.Prepare(pq.CopyIn("links",
		"repas",
		"idtable1",
	))
	if err != nil {
		return err
	}

	for _, item := range items {
		_, err = stmt.Exec(item.Repas, item.IdTable1)
		if err != nil {
			return err
		}
	}

	if _, err = stmt.Exec(); err != nil {
		return err
	}

	if err = stmt.Close(); err != nil {
		return err
	}
	return nil
}

// Delete the link Link from the database.
// Only the foreign keys Repas fields are used in 'item'.
func (item Link) Delete(tx DB) error {
	_, err := tx.Exec(`DELETE FROM links WHERE Repas = $1;`, item.Repas)
	return err
}

// ByRepas returns a map with 'Repas' as keys.
func (items Links) ByRepas() map[RepasID]Links {
	out := make(map[RepasID]Links)
	for _, target := range items {
		out[target.Repas] = append(out[target.Repas], target)
	}
	return out
}

// Repass returns the list of ids of Repas
// contained in this link table.
// They are not garanteed to be distinct.
func (items Links) Repass() []RepasID {
	out := make([]RepasID, len(items))
	for index, target := range items {
		out[index] = target.Repas
	}
	return out
}

func SelectLinksByRepass(tx DB, repass ...RepasID) (Links, error) {
	rows, err := tx.Query("SELECT * FROM links WHERE repas = ANY($1)", RepasIDArrayToPQ(repass))
	if err != nil {
		return nil, err
	}
	return ScanLinks(rows)
}

func DeleteLinksByRepass(tx DB, repass ...RepasID) (Links, error) {
	rows, err := tx.Query("DELETE FROM links WHERE repas = ANY($1) RETURNING *", RepasIDArrayToPQ(repass))
	if err != nil {
		return nil, err
	}
	return ScanLinks(rows)
}

func scanOneProgression(row scanner) (Progression, error) {
	var item Progression
	err := row.Scan(
		&item.Id,
		&item.IdExercice,
	)
	return item, err
}

func ScanProgression(row *sql.Row) (Progression, error) { return scanOneProgression(row) }

// SelectAll returns all the items in the progressions table.
func SelectAllProgressions(db DB) (Progressions, error) {
	rows, err := db.Query("SELECT * FROM progressions")
	if err != nil {
		return nil, err
	}
	return ScanProgressions(rows)
}

// SelectProgression returns the entry matching 'id'.
func SelectProgression(tx DB, id IdProgression) (Progression, error) {
	row := tx.QueryRow("SELECT * FROM progressions WHERE id = $1", id)
	return ScanProgression(row)
}

// SelectProgressions returns the entry matching the given 'ids'.
func SelectProgressions(tx DB, ids ...IdProgression) (Progressions, error) {
	rows, err := tx.Query("SELECT * FROM progressions WHERE id = ANY($1)", IdProgressionArrayToPQ(ids))
	if err != nil {
		return nil, err
	}
	return ScanProgressions(rows)
}

type Progressions map[IdProgression]Progression

func (m Progressions) IDs() []IdProgression {
	out := make([]IdProgression, 0, len(m))
	for i := range m {
		out = append(out, i)
	}
	return out
}

func ScanProgressions(rs *sql.Rows) (Progressions, error) {
	var (
		s   Progression
		err error
	)
	defer func() {
		errClose := rs.Close()
		if err == nil {
			err = errClose
		}
	}()
	structs := make(Progressions, 16)
	for rs.Next() {
		s, err = scanOneProgression(rs)
		if err != nil {
			return nil, err
		}
		structs[s.Id] = s
	}
	if err = rs.Err(); err != nil {
		return nil, err
	}
	return structs, nil
}

// Insert one Progression in the database and returns the item with id filled.
func (item Progression) Insert(tx DB) (out Progression, err error) {
	row := tx.QueryRow(`INSERT INTO progressions (
		idexercice
		) VALUES (
		$1
		) RETURNING *;
		`, item.IdExercice)
	return ScanProgression(row)
}

// Update Progression in the database and returns the new version.
func (item Progression) Update(tx DB) (out Progression, err error) {
	row := tx.QueryRow(`UPDATE progressions SET (
		idexercice
		) = (
		$1
		) WHERE id = $2 RETURNING *;
		`, item.IdExercice, item.Id)
	return ScanProgression(row)
}

// Deletes the Progression and returns the item
func DeleteProgressionById(tx DB, id IdProgression) (Progression, error) {
	row := tx.QueryRow("DELETE FROM progressions WHERE id = $1 RETURNING *;", id)
	return ScanProgression(row)
}

// Deletes the Progression in the database and returns the ids.
func DeleteProgressionsByIDs(tx DB, ids ...IdProgression) ([]IdProgression, error) {
	rows, err := tx.Query("DELETE FROM progressions WHERE id = ANY($1) RETURNING id", IdProgressionArrayToPQ(ids))
	if err != nil {
		return nil, err
	}
	return ScanIdProgressionArray(rows)
}

func scanOneProgressionQuestion(row scanner) (ProgressionQuestion, error) {
	var item ProgressionQuestion
	err := row.Scan(
		&item.IdProgression,
		&item.IdExercice,
		&item.Index,
		&item.History,
	)
	return item, err
}

func ScanProgressionQuestion(row *sql.Row) (ProgressionQuestion, error) {
	return scanOneProgressionQuestion(row)
}

// SelectAll returns all the items in the progression_questions table.
func SelectAllProgressionQuestions(db DB) (ProgressionQuestions, error) {
	rows, err := db.Query("SELECT * FROM progression_questions")
	if err != nil {
		return nil, err
	}
	return ScanProgressionQuestions(rows)
}

type ProgressionQuestions []ProgressionQuestion

func ScanProgressionQuestions(rs *sql.Rows) (ProgressionQuestions, error) {
	var (
		item ProgressionQuestion
		err  error
	)
	defer func() {
		errClose := rs.Close()
		if err == nil {
			err = errClose
		}
	}()
	structs := make(ProgressionQuestions, 0, 16)
	for rs.Next() {
		item, err = scanOneProgressionQuestion(rs)
		if err != nil {
			return nil, err
		}
		structs = append(structs, item)
	}
	if err = rs.Err(); err != nil {
		return nil, err
	}
	return structs, nil
}

// Insert the links ProgressionQuestion in the database.
// It is a no-op if 'items' is empty.
func InsertManyProgressionQuestions(tx *sql.Tx, items ...ProgressionQuestion) error {
	if len(items) == 0 {
		return nil
	}

	stmt, err := tx.Prepare(pq.CopyIn("progression_questions",
		"idprogression",
		"idexercice",
		"index",
		"history",
	))
	if err != nil {
		return err
	}

	for _, item := range items {
		_, err = stmt.Exec(item.IdProgression, item.IdExercice, item.Index, item.History)
		if err != nil {
			return err
		}
	}

	if _, err = stmt.Exec(); err != nil {
		return err
	}

	if err = stmt.Close(); err != nil {
		return err
	}
	return nil
}

// Delete the link ProgressionQuestion from the database.
// Only the foreign keys IdProgression, IdExercice fields are used in 'item'.
func (item ProgressionQuestion) Delete(tx DB) error {
	_, err := tx.Exec(`DELETE FROM progression_questions WHERE IdProgression = $1 AND IdExercice = $2;`, item.IdProgression, item.IdExercice)
	return err
}

// ByIdProgression returns a map with 'IdProgression' as keys.
func (items ProgressionQuestions) ByIdProgression() map[IdProgression]ProgressionQuestions {
	out := make(map[IdProgression]ProgressionQuestions)
	for _, target := range items {
		out[target.IdProgression] = append(out[target.IdProgression], target)
	}
	return out
}

// IdProgressions returns the list of ids of IdProgression
// contained in this link table.
// They are not garanteed to be distinct.
func (items ProgressionQuestions) IdProgressions() []IdProgression {
	out := make([]IdProgression, len(items))
	for index, target := range items {
		out[index] = target.IdProgression
	}
	return out
}

func SelectProgressionQuestionsByIdProgressions(tx DB, idProgressions ...IdProgression) (ProgressionQuestions, error) {
	rows, err := tx.Query("SELECT * FROM progression_questions WHERE idprogression = ANY($1)", IdProgressionArrayToPQ(idProgressions))
	if err != nil {
		return nil, err
	}
	return ScanProgressionQuestions(rows)
}

func DeleteProgressionQuestionsByIdProgressions(tx DB, idProgressions ...IdProgression) (ProgressionQuestions, error) {
	rows, err := tx.Query("DELETE FROM progression_questions WHERE idprogression = ANY($1) RETURNING *", IdProgressionArrayToPQ(idProgressions))
	if err != nil {
		return nil, err
	}
	return ScanProgressionQuestions(rows)
}

// ByIdExercice returns a map with 'IdExercice' as keys.
func (items ProgressionQuestions) ByIdExercice() map[IdExercice]ProgressionQuestions {
	out := make(map[IdExercice]ProgressionQuestions)
	for _, target := range items {
		out[target.IdExercice] = append(out[target.IdExercice], target)
	}
	return out
}

// IdExercices returns the list of ids of IdExercice
// contained in this link table.
// They are not garanteed to be distinct.
func (items ProgressionQuestions) IdExercices() []IdExercice {
	out := make([]IdExercice, len(items))
	for index, target := range items {
		out[index] = target.IdExercice
	}
	return out
}

func SelectProgressionQuestionsByIdExercices(tx DB, idExercices ...IdExercice) (ProgressionQuestions, error) {
	rows, err := tx.Query("SELECT * FROM progression_questions WHERE idexercice = ANY($1)", IdExerciceArrayToPQ(idExercices))
	if err != nil {
		return nil, err
	}
	return ScanProgressionQuestions(rows)
}

func DeleteProgressionQuestionsByIdExercices(tx DB, idExercices ...IdExercice) (ProgressionQuestions, error) {
	rows, err := tx.Query("DELETE FROM progression_questions WHERE idexercice = ANY($1) RETURNING *", IdExerciceArrayToPQ(idExercices))
	if err != nil {
		return nil, err
	}
	return ScanProgressionQuestions(rows)
}

// SelectProgressionQuestionByIdProgressionAndIndex return zero or one item, thanks to a UNIQUE SQL constraint.
func SelectProgressionQuestionByIdProgressionAndIndex(tx DB, idProgression IdProgression, index int) (item ProgressionQuestion, found bool, err error) {
	row := tx.QueryRow("SELECT * FROM progression_questions WHERE IdProgression = $1 AND Index = $2", idProgression, index)
	item, err = ScanProgressionQuestion(row)
	if err == sql.ErrNoRows {
		return item, false, nil
	}
	return item, true, err
}

// SelectProgressionByIdAndIdExercice return zero or one item, thanks to a UNIQUE SQL constraint.
func SelectProgressionByIdAndIdExercice(tx DB, id IdProgression, idExercice int64) (item Progression, found bool, err error) {
	row := tx.QueryRow("SELECT * FROM progressions WHERE Id = $1 AND IdExercice = $2", id, idExercice)
	item, err = ScanProgression(row)
	if err == sql.ErrNoRows {
		return item, false, nil
	}
	return item, true, err
}

func scanOneQuestion(row scanner) (Question, error) {
	var item Question
	err := row.Scan(
		&item.Id,
		&item.Page,
		&item.Public,
		&item.IdTeacher,
		&item.Description,
		&item.NeedExercice,
	)
	return item, err
}

func ScanQuestion(row *sql.Row) (Question, error) { return scanOneQuestion(row) }

// SelectAll returns all the items in the questions table.
func SelectAllQuestions(db DB) (Questions, error) {
	rows, err := db.Query("SELECT * FROM questions")
	if err != nil {
		return nil, err
	}
	return ScanQuestions(rows)
}

// SelectQuestion returns the entry matching 'id'.
func SelectQuestion(tx DB, id int64) (Question, error) {
	row := tx.QueryRow("SELECT * FROM questions WHERE id = $1", id)
	return ScanQuestion(row)
}

// SelectQuestions returns the entry matching the given 'ids'.
func SelectQuestions(tx DB, ids ...int64) (Questions, error) {
	rows, err := tx.Query("SELECT * FROM questions WHERE id = ANY($1)", int64ArrayToPQ(ids))
	if err != nil {
		return nil, err
	}
	return ScanQuestions(rows)
}

type Questions map[int64]Question

func (m Questions) IDs() []int64 {
	out := make([]int64, 0, len(m))
	for i := range m {
		out = append(out, i)
	}
	return out
}

func ScanQuestions(rs *sql.Rows) (Questions, error) {
	var (
		s   Question
		err error
	)
	defer func() {
		errClose := rs.Close()
		if err == nil {
			err = errClose
		}
	}()
	structs := make(Questions, 16)
	for rs.Next() {
		s, err = scanOneQuestion(rs)
		if err != nil {
			return nil, err
		}
		structs[s.Id] = s
	}
	if err = rs.Err(); err != nil {
		return nil, err
	}
	return structs, nil
}

// Insert one Question in the database and returns the item with id filled.
func (item Question) Insert(tx DB) (out Question, err error) {
	row := tx.QueryRow(`INSERT INTO questions (
		page, public, idteacher, description, needexercice
		) VALUES (
		$1, $2, $3, $4, $5
		) RETURNING *;
		`, item.Page, item.Public, item.IdTeacher, item.Description, item.NeedExercice)
	return ScanQuestion(row)
}

// Update Question in the database and returns the new version.
func (item Question) Update(tx DB) (out Question, err error) {
	row := tx.QueryRow(`UPDATE questions SET (
		page, public, idteacher, description, needexercice
		) = (
		$1, $2, $3, $4, $5
		) WHERE id = $6 RETURNING *;
		`, item.Page, item.Public, item.IdTeacher, item.Description, item.NeedExercice, item.Id)
	return ScanQuestion(row)
}

// Deletes the Question and returns the item
func DeleteQuestionById(tx DB, id int64) (Question, error) {
	row := tx.QueryRow("DELETE FROM questions WHERE id = $1 RETURNING *;", id)
	return ScanQuestion(row)
}

// Deletes the Question in the database and returns the ids.
func DeleteQuestionsByIDs(tx DB, ids ...int64) ([]int64, error) {
	rows, err := tx.Query("DELETE FROM questions WHERE id = ANY($1) RETURNING id", int64ArrayToPQ(ids))
	if err != nil {
		return nil, err
	}
	return Scanint64Array(rows)
}

func SelectQuestionsByNeedExercices(tx DB, needExercices ...int64) (Questions, error) {
	rows, err := tx.Query("SELECT * FROM questions WHERE needexercice = ANY($1)", int64ArrayToPQ(needExercices))
	if err != nil {
		return nil, err
	}
	return ScanQuestions(rows)
}

func DeleteQuestionsByNeedExercices(tx DB, needExercices ...int64) ([]int64, error) {
	rows, err := tx.Query("DELETE FROM questions WHERE needexercice = ANY($1) RETURNING id", int64ArrayToPQ(needExercices))
	if err != nil {
		return nil, err
	}
	return Scanint64Array(rows)
}

func scanOneQuestionTag(row scanner) (QuestionTag, error) {
	var item QuestionTag
	err := row.Scan(
		&item.Tag,
		&item.IdQuestion,
	)
	return item, err
}

func ScanQuestionTag(row *sql.Row) (QuestionTag, error) { return scanOneQuestionTag(row) }

// SelectAll returns all the items in the question_tags table.
func SelectAllQuestionTags(db DB) (QuestionTags, error) {
	rows, err := db.Query("SELECT * FROM question_tags")
	if err != nil {
		return nil, err
	}
	return ScanQuestionTags(rows)
}

type QuestionTags []QuestionTag

func ScanQuestionTags(rs *sql.Rows) (QuestionTags, error) {
	var (
		item QuestionTag
		err  error
	)
	defer func() {
		errClose := rs.Close()
		if err == nil {
			err = errClose
		}
	}()
	structs := make(QuestionTags, 0, 16)
	for rs.Next() {
		item, err = scanOneQuestionTag(rs)
		if err != nil {
			return nil, err
		}
		structs = append(structs, item)
	}
	if err = rs.Err(); err != nil {
		return nil, err
	}
	return structs, nil
}

// Insert the links QuestionTag in the database.
// It is a no-op if 'items' is empty.
func InsertManyQuestionTags(tx *sql.Tx, items ...QuestionTag) error {
	if len(items) == 0 {
		return nil
	}

	stmt, err := tx.Prepare(pq.CopyIn("question_tags",
		"tag",
		"idquestion",
	))
	if err != nil {
		return err
	}

	for _, item := range items {
		_, err = stmt.Exec(item.Tag, item.IdQuestion)
		if err != nil {
			return err
		}
	}

	if _, err = stmt.Exec(); err != nil {
		return err
	}

	if err = stmt.Close(); err != nil {
		return err
	}
	return nil
}

// Delete the link QuestionTag from the database.
// Only the foreign keys IdQuestion fields are used in 'item'.
func (item QuestionTag) Delete(tx DB) error {
	_, err := tx.Exec(`DELETE FROM question_tags WHERE IdQuestion = $1;`, item.IdQuestion)
	return err
}

// ByIdQuestion returns a map with 'IdQuestion' as keys.
func (items QuestionTags) ByIdQuestion() map[int64]QuestionTags {
	out := make(map[int64]QuestionTags)
	for _, target := range items {
		out[target.IdQuestion] = append(out[target.IdQuestion], target)
	}
	return out
}

// IdQuestions returns the list of ids of IdQuestion
// contained in this link table.
// They are not garanteed to be distinct.
func (items QuestionTags) IdQuestions() []int64 {
	out := make([]int64, len(items))
	for index, target := range items {
		out[index] = target.IdQuestion
	}
	return out
}

func SelectQuestionTagsByIdQuestions(tx DB, idQuestions ...int64) (QuestionTags, error) {
	rows, err := tx.Query("SELECT * FROM question_tags WHERE idquestion = ANY($1)", int64ArrayToPQ(idQuestions))
	if err != nil {
		return nil, err
	}
	return ScanQuestionTags(rows)
}

func DeleteQuestionTagsByIdQuestions(tx DB, idQuestions ...int64) (QuestionTags, error) {
	rows, err := tx.Query("DELETE FROM question_tags WHERE idquestion = ANY($1) RETURNING *", int64ArrayToPQ(idQuestions))
	if err != nil {
		return nil, err
	}
	return ScanQuestionTags(rows)
}

// SelectQuestionTagByIdQuestionAndTag return zero or one item, thanks to a UNIQUE SQL constraint.
func SelectQuestionTagByIdQuestionAndTag(tx DB, idQuestion int64, tag string) (item QuestionTag, found bool, err error) {
	row := tx.QueryRow("SELECT * FROM question_tags WHERE IdQuestion = $1 AND Tag = $2", idQuestion, tag)
	item, err = ScanQuestionTag(row)
	if err == sql.ErrNoRows {
		return item, false, nil
	}
	return item, true, err
}

func scanOneRepas(row scanner) (Repas, error) {
	var item Repas
	err := row.Scan(
		&item.Order,
		&item.Id,
	)
	return item, err
}

func ScanRepas(row *sql.Row) (Repas, error) { return scanOneRepas(row) }

// SelectAll returns all the items in the repass table.
func SelectAllRepass(db DB) (Repass, error) {
	rows, err := db.Query("SELECT * FROM repass")
	if err != nil {
		return nil, err
	}
	return ScanRepass(rows)
}

// SelectRepas returns the entry matching 'id'.
func SelectRepas(tx DB, id RepasID) (Repas, error) {
	row := tx.QueryRow("SELECT * FROM repass WHERE id = $1", id)
	return ScanRepas(row)
}

// SelectRepass returns the entry matching the given 'ids'.
func SelectRepass(tx DB, ids ...RepasID) (Repass, error) {
	rows, err := tx.Query("SELECT * FROM repass WHERE id = ANY($1)", RepasIDArrayToPQ(ids))
	if err != nil {
		return nil, err
	}
	return ScanRepass(rows)
}

type Repass map[RepasID]Repas

func (m Repass) IDs() []RepasID {
	out := make([]RepasID, 0, len(m))
	for i := range m {
		out = append(out, i)
	}
	return out
}

func ScanRepass(rs *sql.Rows) (Repass, error) {
	var (
		s   Repas
		err error
	)
	defer func() {
		errClose := rs.Close()
		if err == nil {
			err = errClose
		}
	}()
	structs := make(Repass, 16)
	for rs.Next() {
		s, err = scanOneRepas(rs)
		if err != nil {
			return nil, err
		}
		structs[s.Id] = s
	}
	if err = rs.Err(); err != nil {
		return nil, err
	}
	return structs, nil
}

// Insert one Repas in the database and returns the item with id filled.
func (item Repas) Insert(tx DB) (out Repas, err error) {
	row := tx.QueryRow(`INSERT INTO repass (
		order
		) VALUES (
		$1
		) RETURNING *;
		`, item.Order)
	return ScanRepas(row)
}

// Update Repas in the database and returns the new version.
func (item Repas) Update(tx DB) (out Repas, err error) {
	row := tx.QueryRow(`UPDATE repass SET (
		order
		) = (
		$1
		) WHERE id = $2 RETURNING *;
		`, item.Order, item.Id)
	return ScanRepas(row)
}

// Deletes the Repas and returns the item
func DeleteRepasById(tx DB, id RepasID) (Repas, error) {
	row := tx.QueryRow("DELETE FROM repass WHERE id = $1 RETURNING *;", id)
	return ScanRepas(row)
}

// Deletes the Repas in the database and returns the ids.
func DeleteRepassByIDs(tx DB, ids ...RepasID) ([]RepasID, error) {
	rows, err := tx.Query("DELETE FROM repass WHERE id = ANY($1) RETURNING id", RepasIDArrayToPQ(ids))
	if err != nil {
		return nil, err
	}
	return ScanRepasIDArray(rows)
}

func scanOneTable1(row scanner) (Table1, error) {
	var item Table1
	err := row.Scan(
		&item.Id,
		&item.Ex1,
		&item.Ex2,
		&item.L,
		&item.Other,
	)
	return item, err
}

func ScanTable1(row *sql.Row) (Table1, error) { return scanOneTable1(row) }

// SelectAll returns all the items in the table1s table.
func SelectAllTable1s(db DB) (Table1s, error) {
	rows, err := db.Query("SELECT * FROM table1s")
	if err != nil {
		return nil, err
	}
	return ScanTable1s(rows)
}

// SelectTable1 returns the entry matching 'id'.
func SelectTable1(tx DB, id int64) (Table1, error) {
	row := tx.QueryRow("SELECT * FROM table1s WHERE id = $1", id)
	return ScanTable1(row)
}

// SelectTable1s returns the entry matching the given 'ids'.
func SelectTable1s(tx DB, ids ...int64) (Table1s, error) {
	rows, err := tx.Query("SELECT * FROM table1s WHERE id = ANY($1)", int64ArrayToPQ(ids))
	if err != nil {
		return nil, err
	}
	return ScanTable1s(rows)
}

type Table1s map[int64]Table1

func (m Table1s) IDs() []int64 {
	out := make([]int64, 0, len(m))
	for i := range m {
		out = append(out, i)
	}
	return out
}

func ScanTable1s(rs *sql.Rows) (Table1s, error) {
	var (
		s   Table1
		err error
	)
	defer func() {
		errClose := rs.Close()
		if err == nil {
			err = errClose
		}
	}()
	structs := make(Table1s, 16)
	for rs.Next() {
		s, err = scanOneTable1(rs)
		if err != nil {
			return nil, err
		}
		structs[s.Id] = s
	}
	if err = rs.Err(); err != nil {
		return nil, err
	}
	return structs, nil
}

// Insert one Table1 in the database and returns the item with id filled.
func (item Table1) Insert(tx DB) (out Table1, err error) {
	row := tx.QueryRow(`INSERT INTO table1s (
		ex1, ex2, l, other
		) VALUES (
		$1, $2, $3, $4
		) RETURNING *;
		`, item.Ex1, item.Ex2, item.L, item.Other)
	return ScanTable1(row)
}

// Update Table1 in the database and returns the new version.
func (item Table1) Update(tx DB) (out Table1, err error) {
	row := tx.QueryRow(`UPDATE table1s SET (
		ex1, ex2, l, other
		) = (
		$1, $2, $3, $4
		) WHERE id = $5 RETURNING *;
		`, item.Ex1, item.Ex2, item.L, item.Other, item.Id)
	return ScanTable1(row)
}

// Deletes the Table1 and returns the item
func DeleteTable1ById(tx DB, id int64) (Table1, error) {
	row := tx.QueryRow("DELETE FROM table1s WHERE id = $1 RETURNING *;", id)
	return ScanTable1(row)
}

// Deletes the Table1 in the database and returns the ids.
func DeleteTable1sByIDs(tx DB, ids ...int64) ([]int64, error) {
	rows, err := tx.Query("DELETE FROM table1s WHERE id = ANY($1) RETURNING id", int64ArrayToPQ(ids))
	if err != nil {
		return nil, err
	}
	return Scanint64Array(rows)
}

// ByEx1 returns a map with 'Ex1' as keys.
func (items Table1s) ByEx1() map[RepasID]Table1s {
	out := make(map[RepasID]Table1s)
	for _, target := range items {
		dict := out[target.Ex1]
		if dict == nil {
			dict = make(Table1s)
		}
		dict[target.Id] = target
		out[target.Ex1] = dict
	}
	return out
}

// Ex1s returns the list of ids of Ex1
// contained in this table.
// They are not garanteed to be distinct.
func (items Table1s) Ex1s() []RepasID {
	out := make([]RepasID, 0, len(items))
	for _, target := range items {
		out = append(out, target.Ex1)
	}
	return out
}

func SelectTable1sByEx1s(tx DB, ex1s ...RepasID) (Table1s, error) {
	rows, err := tx.Query("SELECT * FROM table1s WHERE ex1 = ANY($1)", RepasIDArrayToPQ(ex1s))
	if err != nil {
		return nil, err
	}
	return ScanTable1s(rows)
}

func DeleteTable1sByEx1s(tx DB, ex1s ...RepasID) ([]int64, error) {
	rows, err := tx.Query("DELETE FROM table1s WHERE ex1 = ANY($1) RETURNING id", RepasIDArrayToPQ(ex1s))
	if err != nil {
		return nil, err
	}
	return Scanint64Array(rows)
}

// ByEx2 returns a map with 'Ex2' as keys.
func (items Table1s) ByEx2() map[RepasID]Table1s {
	out := make(map[RepasID]Table1s)
	for _, target := range items {
		dict := out[target.Ex2]
		if dict == nil {
			dict = make(Table1s)
		}
		dict[target.Id] = target
		out[target.Ex2] = dict
	}
	return out
}

// Ex2s returns the list of ids of Ex2
// contained in this table.
// They are not garanteed to be distinct.
func (items Table1s) Ex2s() []RepasID {
	out := make([]RepasID, 0, len(items))
	for _, target := range items {
		out = append(out, target.Ex2)
	}
	return out
}

func SelectTable1sByEx2s(tx DB, ex2s ...RepasID) (Table1s, error) {
	rows, err := tx.Query("SELECT * FROM table1s WHERE ex2 = ANY($1)", RepasIDArrayToPQ(ex2s))
	if err != nil {
		return nil, err
	}
	return ScanTable1s(rows)
}

func DeleteTable1sByEx2s(tx DB, ex2s ...RepasID) ([]int64, error) {
	rows, err := tx.Query("DELETE FROM table1s WHERE ex2 = ANY($1) RETURNING id", RepasIDArrayToPQ(ex2s))
	if err != nil {
		return nil, err
	}
	return Scanint64Array(rows)
}

func SelectTable1sByLs(tx DB, ls ...int64) (Table1s, error) {
	rows, err := tx.Query("SELECT * FROM table1s WHERE l = ANY($1)", int64ArrayToPQ(ls))
	if err != nil {
		return nil, err
	}
	return ScanTable1s(rows)
}

func DeleteTable1sByLs(tx DB, ls ...int64) ([]int64, error) {
	rows, err := tx.Query("DELETE FROM table1s WHERE l = ANY($1) RETURNING id", int64ArrayToPQ(ls))
	if err != nil {
		return nil, err
	}
	return Scanint64Array(rows)
}

func SelectTable1sByOthers(tx DB, others ...RepasID) (Table1s, error) {
	rows, err := tx.Query("SELECT * FROM table1s WHERE other = ANY($1)", RepasIDArrayToPQ(others))
	if err != nil {
		return nil, err
	}
	return ScanTable1s(rows)
}

func DeleteTable1sByOthers(tx DB, others ...RepasID) ([]int64, error) {
	rows, err := tx.Query("DELETE FROM table1s WHERE other = ANY($1) RETURNING id", RepasIDArrayToPQ(others))
	if err != nil {
		return nil, err
	}
	return Scanint64Array(rows)
}

func loadJSON(out interface{}, src interface{}) error {
	if src == nil {
		return nil //zero value out
	}
	bs, ok := src.([]byte)
	if !ok {
		return errors.New("not a []byte")
	}
	return json.Unmarshal(bs, out)
}

func dumpJSON(s interface{}) (driver.Value, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return driver.Value(string(b)), nil
}

func IdExerciceArrayToPQ(ids []IdExercice) pq.Int64Array {
	out := make(pq.Int64Array, len(ids))
	for i, v := range ids {
		out[i] = int64(v)
	}
	return out
}

// ScanIdExerciceArray scans the result of a query returning a
// list of ID's.
func ScanIdExerciceArray(rs *sql.Rows) ([]IdExercice, error) {
	defer rs.Close()
	ints := make([]IdExercice, 0, 16)
	var err error
	for rs.Next() {
		var s IdExercice
		if err = rs.Scan(&s); err != nil {
			return nil, err
		}
		ints = append(ints, s)
	}
	if err = rs.Err(); err != nil {
		return nil, err
	}
	return ints, nil
}

type IdExerciceSet map[IdExercice]bool

func NewIdExerciceSetFrom(ids []IdExercice) IdExerciceSet {
	out := make(IdExerciceSet, len(ids))
	for _, key := range ids {
		out[key] = true
	}
	return out
}

func (s IdExerciceSet) Add(id IdExercice) { s[id] = true }

func (s IdExerciceSet) Has(id IdExercice) bool { return s[id] }

func (s IdExerciceSet) Keys() []IdExercice {
	out := make([]IdExercice, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	return out
}

func IdProgressionArrayToPQ(ids []IdProgression) pq.Int64Array {
	out := make(pq.Int64Array, len(ids))
	for i, v := range ids {
		out[i] = int64(v)
	}
	return out
}

// ScanIdProgressionArray scans the result of a query returning a
// list of ID's.
func ScanIdProgressionArray(rs *sql.Rows) ([]IdProgression, error) {
	defer rs.Close()
	ints := make([]IdProgression, 0, 16)
	var err error
	for rs.Next() {
		var s IdProgression
		if err = rs.Scan(&s); err != nil {
			return nil, err
		}
		ints = append(ints, s)
	}
	if err = rs.Err(); err != nil {
		return nil, err
	}
	return ints, nil
}

type IdProgressionSet map[IdProgression]bool

func NewIdProgressionSetFrom(ids []IdProgression) IdProgressionSet {
	out := make(IdProgressionSet, len(ids))
	for _, key := range ids {
		out[key] = true
	}
	return out
}

func (s IdProgressionSet) Add(id IdProgression) { s[id] = true }

func (s IdProgressionSet) Has(id IdProgression) bool { return s[id] }

func (s IdProgressionSet) Keys() []IdProgression {
	out := make([]IdProgression, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	return out
}

func RepasIDArrayToPQ(ids []RepasID) pq.Int64Array {
	out := make(pq.Int64Array, len(ids))
	for i, v := range ids {
		out[i] = int64(v)
	}
	return out
}

// ScanRepasIDArray scans the result of a query returning a
// list of ID's.
func ScanRepasIDArray(rs *sql.Rows) ([]RepasID, error) {
	defer rs.Close()
	ints := make([]RepasID, 0, 16)
	var err error
	for rs.Next() {
		var s RepasID
		if err = rs.Scan(&s); err != nil {
			return nil, err
		}
		ints = append(ints, s)
	}
	if err = rs.Err(); err != nil {
		return nil, err
	}
	return ints, nil
}

type RepasIDSet map[RepasID]bool

func NewRepasIDSetFrom(ids []RepasID) RepasIDSet {
	out := make(RepasIDSet, len(ids))
	for _, key := range ids {
		out[key] = true
	}
	return out
}

func (s RepasIDSet) Add(id RepasID) { s[id] = true }

func (s RepasIDSet) Has(id RepasID) bool { return s[id] }

func (s RepasIDSet) Keys() []RepasID {
	out := make([]RepasID, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	return out
}

func int64ArrayToPQ(ids []int64) pq.Int64Array { return ids }

// Scanint64Array scans the result of a query returning a
// list of ID's.
func Scanint64Array(rs *sql.Rows) ([]int64, error) {
	defer rs.Close()
	ints := make([]int64, 0, 16)
	var err error
	for rs.Next() {
		var s int64
		if err = rs.Scan(&s); err != nil {
			return nil, err
		}
		ints = append(ints, s)
	}
	if err = rs.Err(); err != nil {
		return nil, err
	}
	return ints, nil
}

type int64Set map[int64]bool

func Newint64SetFrom(ids []int64) int64Set {
	out := make(int64Set, len(ids))
	for _, key := range ids {
		out[key] = true
	}
	return out
}

func (s int64Set) Add(id int64) { s[id] = true }

func (s int64Set) Has(id int64) bool { return s[id] }

func (s int64Set) Keys() []int64 {
	out := make([]int64, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	return out
}

func (s *EnumArray) Scan(src interface{}) error  { return loadJSON(s, src) }
func (s EnumArray) Value() (driver.Value, error) { return dumpJSON(s) }

func (s *Map) Scan(src interface{}) error  { return loadJSON(s, src) }
func (s Map) Value() (driver.Value, error) { return dumpJSON(s) }

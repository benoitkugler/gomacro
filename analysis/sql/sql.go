// Package sql implements the logic required to
// analyze the links between SQL tables (represented as structs in the Go source code).
package sql

import (
	"go/types"
	"reflect"
	"strings"

	an "github.com/benoitkugler/gomacro/analysis"
)

// SelectTables returns the SQL tables found in the given analysis.
func SelectTables(ana *an.Analysis) (out []Table) {
	for _, ty := range ana.Source {
		st, ok := ana.Types[ty].(*an.Struct)
		if !ok {
			continue
		}
		out = append(out, NewTable(st))
	}
	return out
}

// TableName is a singular name of a table entity,
// as found in Go source.
type TableName string

func isInt64(ty types.Type) bool {
	basic, ok := ty.Underlying().(*types.Basic)
	if !ok {
		return false
	}
	return basic.Kind() == types.Int64
}

func isNullInt64(ty an.Type) bool {
	named, ok := ty.Type().(*types.Named)
	if !ok {
		return false
	}
	nullable := isNullXXX(named)
	return nullable != nil && isInt64(nullable)
}

// return the type of a sql.NullXXX struct
// or nil
func isNullXXX(typ *types.Named) types.Type {
	isFieldValid := func(field *types.Var) bool {
		typ, ok := field.Type().Underlying().(*types.Basic)
		return ok && typ.Info() == types.IsBoolean && field.Name() == "Valid"
	}

	str, ok := typ.Underlying().(*types.Struct)
	if !ok || str.NumFields() != 2 { // not a possible struct
		return nil
	}
	if isFieldValid(str.Field(0)) {
		return str.Field(1).Type()
	} else if isFieldValid(str.Field(1)) {
		return str.Field(0).Type()
	}
	return nil
}

// return a non empty table name for int64 named types whose name starts or finish
// by ID, id, Id, iD
func isTableID(ty an.Type) TableName {
	named, ok := ty.(*an.Named)
	if !ok {
		return ""
	}
	if ok := isInt64(named.Underlying.Type()); !ok {
		return ""
	}
	name := an.LocalName(ty)
	if len(name) > 2 && strings.HasPrefix(strings.ToLower(name), "id") {
		return TableName(name[2:])
	} else if len(name) > 2 && strings.HasSuffix(strings.ToLower(name), "id") {
		return TableName(name[:len(name)-2])
	}
	return ""
}

// ForeignKey is a struct field used as (single) SQL foreign key.
type ForeignKey struct {
	// F has underlying type int64 or sql.NullInt64
	F an.StructField

	// Target is the foreign table being referenced
	Target TableName

	// IsUnique is true if an SQL UNIQUE constraint
	// was added as special comment
	IsUnique bool
}

func (ta *Table) newForeignKey(field an.StructField) (ForeignKey, bool) {
	out := ForeignKey{F: field, IsUnique: ta.uniqueColumns[field.Field.Name()]}
	// look for an ID type
	if table := isTableID(field.Type); table != "" && string(table) != ta.Name.Obj().Name() {
		out.Target = table
		return out, true
	}

	// look for a tag
	if table := reflect.StructTag(field.Tag).Get("gomacro-sql-foreign"); table != "" {
		if !(isInt64(field.Type.Type()) || isNullInt64(field.Type)) {
			panic("invalid type for foreign key " + field.Field.Name())
		}
		out.Target = TableName(table)
		return out, true
	}

	return ForeignKey{}, false
}

// IsNullable returns true if the key is optional.
func (fk ForeignKey) IsNullable() bool {
	return !isInt64(fk.F.Field.Type())
}

// OnDelete returns the action defined by the tag
// `gomacro-sql-on-delete:"<action>"`, or an empty string.
func (fk ForeignKey) OnDelete() string {
	return fk.F.Tag.Get("gomacro-sql-on-delete")
}

type Column struct {
	// SQLType is the resolved SQL type
	SQLType Type

	// Field is the Go struct field yielding this column
	Field an.StructField
}

// Table is a Struct used as SQL table.
type Table struct {
	Name *types.Named

	uniqueColumns map[string]bool

	// Columns only exposes exported struct fields.
	Columns []Column

	// CustomComments are the user provided constraints
	// defined with `// gomacro:SQL <constraint>` comments
	CustomConstraints []string

	uniquesCols [][]string // not filtered
}

func NewTable(s *an.Struct) Table {
	out := Table{
		Name: s.Name,
	}
	for _, fi := range s.Fields {
		if !fi.Field.Exported() {
			continue
		}

		out.Columns = append(out.Columns, Column{
			Field:   fi,
			SQLType: newType(fi.Type),
		})
	}

	out.processComments(s.Comments)

	return out
}

// TableName return the Go table name.
func (ta *Table) TableName() TableName {
	return TableName(ta.Name.Obj().Name())
}

func (ta *Table) processComments(comments []an.SpecialComment) {
	ta.uniqueColumns = make(map[string]bool)

	for _, comment := range comments {
		if comment.Kind != an.CommentSQL {
			continue
		}

		if column := isUniqueConstraint(comment.Content); column != "" {
			ta.uniqueColumns[column] = true
		}

		if columns := isUniquesConstraint(comment.Content); len(columns) != 0 {
			ta.uniquesCols = append(ta.uniquesCols, columns)
		}

		ta.CustomConstraints = append(ta.CustomConstraints, comment.Content)
	}
}

// Primary returns the index of the slice element
// which is the ID field, or -1 if not found
func (ta Table) Primary() int {
	for i, field := range ta.Columns {
		if strings.ToLower(field.Field.Field.Name()) == "id" {
			return i
		}
	}
	return -1
}

// ForeignKeys returns the columns which are references
// into other tables (sorted by name).
// They are identified by table ID types or with the gomacro-sql-foreign:"<table>" tag.
func (ta Table) ForeignKeys() (out []ForeignKey) {
	for _, field := range ta.Columns {
		if fk, ok := ta.newForeignKey(field.Field); ok {
			out = append(out, fk)
		}
	}
	return out
}

// `AdditionalUniqueCols` returns the columns which have a
// UNIQUE constraint.
// The columns returned by `ForeignKeys` are not included,
// since they usually require additional handling.
func (ta Table) AdditionalUniqueCols() [][]Column {
	foreignKeys := make(map[string]bool)
	for _, key := range ta.ForeignKeys() {
		foreignKeys[key.F.Field.Name()] = true
	}

	colsByName := make(map[string]Column)
	for _, col := range ta.Columns {
		colsByName[col.Field.Field.Name()] = col
	}

	var out [][]Column
	for _, colNames := range ta.uniquesCols {
		// ignore foreign keys
		if len(colNames) == 1 && foreignKeys[colNames[0]] {
			continue
		}
		cols := make([]Column, len(colNames))
		for i, name := range colNames {
			cols[i] = colsByName[name]
		}

		out = append(out, cols)
	}
	return out
}

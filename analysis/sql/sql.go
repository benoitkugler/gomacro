// Package sql implements the logic required to
// analyze the links between SQL tables (represented as structs in the Go source code).
package sql

import (
	"go/types"
	"reflect"
	"strings"

	an "github.com/benoitkugler/gomacro/analysis"
)

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
	nullable := isNullable(named)
	return nullable != nil && isInt64(nullable)
}

// return the type of a sql.NullXXX struct
// or nil
func isNullable(typ *types.Named) types.Type {
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
	// F has type int64 or sql.NullInt64
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
	if table := isTableID(field.Type); table != "" && string(table) != an.LocalName(ta.S) {
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

// Table is a Struct used as SQL table.
type Table struct {
	S *an.Struct

	uniqueColumns map[string]bool
}

func NewTable(s *an.Struct) Table {
	out := Table{S: s}
	out.processComments()
	return out
}

func (ta *Table) processComments() {
	ta.uniqueColumns = make(map[string]bool)

	for _, comment := range ta.S.Comments {
		if comment.Kind != an.CommentSQL {
			continue
		}

		if column := isUniqueConstraint(comment.Content); column != "" {
			ta.uniqueColumns[column] = true
		}
	}
}

// Primary returns a pointer to the slice element
// which is the ID field, or nil if not found
func (ta Table) Primary() *an.StructField {
	for i, field := range ta.S.Fields {
		if strings.ToLower(field.Field.Name()) == "id" {
			return &ta.S.Fields[i]
		}
	}
	return nil
}

// ForeignKeys returns the columns which are references
// into other tables (sorted by name).
// They are identified by table ID types or with the gomacro-sql-foreign:"<table>" tag.
func (ta Table) ForeignKeys() (out []ForeignKey) {
	for _, field := range ta.S.Fields {
		if fk, ok := ta.newForeignKey(field); ok {
			out = append(out, fk)
		}
	}
	return out
}

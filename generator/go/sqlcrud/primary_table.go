package sqlcrud

import (
	"fmt"
	"strings"

	"github.com/benoitkugler/gomacro/analysis/sql"
	gen "github.com/benoitkugler/gomacro/generator"
)

func (ctx context) generatePrimaryTable(ta sql.Table) []gen.Declaration {
	primaryIndex := ta.Primary()
	idTypeName := ctx.typeName(ta.Columns[primaryIndex].Field.Type.Type())
	goTypeName := ta.TableName()
	sqlTableName := gen.SQLTableName(goTypeName)

	var (
		scanFields                 = make([]string, len(ta.Columns))
		columnNamesWithoutPrimary  []string // required for create/update statements
		placeholdersWithoutPrimary []string // required for create/update statements
		goFieldsWithoutPrimary     []string // required for create/update statements
	)
	for i, col := range ta.Columns {
		fieldName := col.Field.Field.Name()
		columnName := sqlColumnName(col.Field)
		scanFields[i] = fmt.Sprintf("&item.%s,", fieldName)

		if i != primaryIndex {
			columnNamesWithoutPrimary = append(columnNamesWithoutPrimary, columnName)
			// placeholders like $1 $2 ...
			placeholdersWithoutPrimary = append(placeholdersWithoutPrimary, fmt.Sprintf("$%d", len(placeholdersWithoutPrimary)+1))
			goFieldsWithoutPrimary = append(goFieldsWithoutPrimary, fmt.Sprintf("item.%s", fieldName))
		}
	}

	content := fmt.Sprintf(`
func scanOne%[1]s(row scanner) (%[1]s, error) {
	var item %[1]s
	err := row.Scan(
		%[4]s
	)
	return item, err
}

func Scan%[1]s(row *sql.Row) (%[1]s, error) { return scanOne%[1]s(row) }

// SelectAll returns all the items in the %[3]s table.
func SelectAll%[1]ss(db DB) (%[1]ss, error) {
	rows, err := db.Query("SELECT * FROM %[3]s")
	if err != nil {
		return nil, err
	}
	return Scan%[1]ss(rows)
}

// Select%[1]s returns the entry matching 'id'.
func Select%[1]s(tx DB, id %[2]s) (%[1]s, error) {
	row := tx.QueryRow("SELECT * FROM %[3]s WHERE id = $1", id)
	return Scan%[1]s(row)
}

// Select%[1]ss returns the entry matching the given 'ids'.
func Select%[1]ss(tx DB, ids ...%[2]s) (%[1]ss, error) {
	rows, err := tx.Query("SELECT * FROM %[3]s WHERE id = ANY($1)", %[2]sArrayToPQ(ids))
	if err != nil {
		return nil, err
	}
	return Scan%[1]ss(rows)
}

type %[1]ss map[%[2]s]%[1]s

func (m %[1]ss) IDs() []%[2]s {
	out := make([]%[2]s, 0, len(m))
	for i := range m {
		out = append(out, i)
	}
	return out
}

func Scan%[1]ss(rs *sql.Rows) (%[1]ss, error) {
	var (
		s %[1]s
		err error
	)
	defer func() {
		errClose := rs.Close()
		if err == nil {
			err = errClose
		}
	}()
	structs := make(%[1]ss,  16)
	for rs.Next() {
		s, err = scanOne%[1]s(rs)
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

// Insert one %[1]s in the database and returns the item with id filled.
func (item %[1]s) Insert(tx DB) (out %[1]s, err error) {
	row := tx.QueryRow(`+"`"+`INSERT INTO %[3]s (
		%[5]s
		) VALUES (
		%[6]s
		) RETURNING *;
		`+"`,"+`%[7]s)
	return Scan%[1]s(row)
}

// Update %[1]s in the database and returns the new version.
func (item %[1]s) Update(tx DB) (out %[1]s, err error) {
	row := tx.QueryRow(`+"`"+`UPDATE %[3]s SET (
		%[5]s
		) = (
		%[6]s
		) WHERE id = $%[8]d RETURNING *;
		`+"`,"+`%[7]s, item.%[9]s)
	return Scan%[1]s(row)
}

// Deletes the %[1]s and returns the item
func Delete%[1]sById(tx DB, id %[2]s) (%[1]s, error) {
	row := tx.QueryRow("DELETE FROM %[3]s WHERE id = $1 RETURNING *;", id)
	return Scan%[1]s(row)
}

// Deletes the %[1]s in the database and returns the ids.
func Delete%[1]ssByIDs(tx DB, ids ...%[2]s) ([]%[2]s, error) {
	rows, err := tx.Query("DELETE FROM %[3]s WHERE id = ANY($1) RETURNING id", %[2]sArrayToPQ(ids))
	if err != nil {
		return nil, err
	}
	return Scan%[2]sArray(rows)
}	
`, goTypeName, idTypeName, sqlTableName,
		strings.Join(scanFields, "\n"), strings.Join(columnNamesWithoutPrimary, ", "), strings.Join(placeholdersWithoutPrimary, ", "), strings.Join(goFieldsWithoutPrimary, ", "),
		len(ta.Columns), ta.Columns[primaryIndex].Field.Field.Name(),
	)

	var out []gen.Declaration

	// generate "join like" queries
	for _, key := range ta.ForeignKeys() {
		fieldName := key.F.Field.Name()
		columnName := sqlColumnName(key.F)
		varName := gen.ToLowerFirst(fieldName)
		keyTypeName := "int64"
		if !key.IsNullable() {
			keyTypeName = ctx.typeName(key.F.Field.Type())
		}

		// add int array converter if required
		if ctx.generateArrayConverter(key) {
			out = append(out, ctx.idArrayConverters(keyTypeName))
		}

		if key.IsUnique {
			content += fmt.Sprintf(`
			// Select%[1]sBy%[2]s return zero or one item, thanks to a UNIQUE SQL constraint.
			func Select%[1]sBy%[2]s(tx DB, %[3]s %[5]s) (item %[1]s, found bool, err error) {
				row := tx.QueryRow("SELECT * FROM %[4]s WHERE %[6]s = $1", %[3]s)
				item, err = Scan%[1]s(row)
				if err == sql.ErrNoRows {
					return item, false, nil
				}
				return item, true, err
			}	
			`, goTypeName, fieldName, varName, sqlTableName, keyTypeName, columnName)
		}

		content += fmt.Sprintf(`
		func Select%[1]ssBy%[2]ss(tx DB, %[3]ss ...%[6]s) (%[1]ss, error) {
			rows, err := tx.Query("SELECT * FROM %[4]s WHERE %[7]s = ANY($1)", %[6]sArrayToPQ(%[3]ss))
			if err != nil {
				return nil, err
			}
			return Scan%[1]ss(rows)
		}	

		func Delete%[1]ssBy%[2]ss(tx DB, %[3]ss ...%[6]s) ([]%[5]s, error) {
			rows, err := tx.Query("DELETE FROM %[4]s WHERE %[7]s = ANY($1) RETURNING id", %[6]sArrayToPQ(%[3]ss))
			if err != nil {
				return nil, err
			}
			return Scan%[5]sArray(rows)
		}	
		`, goTypeName, fieldName, varName, sqlTableName, idTypeName, keyTypeName, columnName)
	}

	return append(out,
		ctx.idArrayConverters(idTypeName),
		gen.Declaration{
			ID:      string(goTypeName),
			Content: content,
		},
	)
}

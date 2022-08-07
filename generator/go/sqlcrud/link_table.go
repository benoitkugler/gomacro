package sqlcrud

import (
	"fmt"
	"strings"

	"github.com/benoitkugler/gomacro/analysis/sql"
	gen "github.com/benoitkugler/gomacro/generator"
)

func (ctx context) generateLinkTable(ta sql.Table) (out []gen.Declaration) {
	goTypeName := ta.TableName()
	sqlTableName := gen.SQLTableName(goTypeName)

	var (
		scanFields        = make([]string, len(ta.Columns))
		quotedColumnNames = make([]string, len(ta.Columns)) // required for insert statements

		goFields = make([]string, len(ta.Columns)) // required for create/update statements
	)
	for i, col := range ta.Columns {
		fieldName := col.Field.Field.Name()
		columnName := sqlColumnName(col.Field)
		scanFields[i] = fmt.Sprintf("&item.%s,", fieldName)

		quotedColumnNames[i] = fmt.Sprintf("%q,", columnName)
		goFields[i] = fmt.Sprintf("item.%s", fieldName)
	}

	var (
		foreignKeyFields []string
		foreignKeyComps  []string
		foreignKeyAccess []string
	)
	for i, key := range ta.ForeignKeys() {
		fieldName := key.F.Field.Name()
		foreignKeyFields = append(foreignKeyFields, fieldName)
		if key.IsNullable() { // add a guard against null values
			foreignKeyComps = append(foreignKeyComps, fmt.Sprintf("((%[1]s IS NULL AND $%[2]d IS NULL) OR %[1]s = $%[2]d)", fieldName, i+1))
		} else {
			foreignKeyComps = append(foreignKeyComps, fmt.Sprintf("%s = $%d", fieldName, i+1))
		}
		foreignKeyAccess = append(foreignKeyAccess, fmt.Sprintf("item.%s", fieldName))
	}

	content := fmt.Sprintf(`
	func scanOne%[1]s(row scanner) (%[1]s, error) {
		var item %[1]s
		err := row.Scan(
			%[3]s
		)
		return item, err
	}

	func Scan%[1]s(row *sql.Row) (%[1]s, error) { return scanOne%[1]s(row) }

	// SelectAll returns all the items in the %[2]s table.
	func SelectAll%[1]ss(db DB) (%[1]ss, error) {
		rows, err := db.Query("SELECT * FROM %[2]s")
		if err != nil {
			return nil, err
		}
		return Scan%[1]ss(rows)
	}

	type %[1]ss []%[1]s

	func Scan%[1]ss(rs *sql.Rows) (%[1]ss , error) {
		var (
			item %[1]s
			err error
		)
		defer func() {
			errClose := rs.Close()
			if err == nil {
				err = errClose
			}
		}()
		structs := make(%[1]ss , 0, 16)
		for rs.Next() {
			item, err = scanOne%[1]s(rs)
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

	// Insert the links %[1]s in the database.
	// It is a no-op if 'items' is empty.
	func InsertMany%[1]ss(tx *sql.Tx, items ...%[1]s) error {
		if len(items) == 0 {
			return nil
		}

		stmt, err := tx.Prepare(pq.CopyIn("%[2]s", 
			%[4]s
		))
		if err != nil {
			return err
		}

		for _, item := range items {
			_, err = stmt.Exec(%[5]s)
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

	// Delete the link %[1]s from the database.
	// Only the foreign keys %[6]s fields are used in 'item'.
	func (item %[1]s) Delete(tx DB) error {
		_, err := tx.Exec(`+"`"+`DELETE FROM %[2]s WHERE %[7]s;`+
		"`,"+` %[8]s)
		return err
	}
	`, goTypeName, sqlTableName,
		strings.Join(scanFields, "\n"), strings.Join(quotedColumnNames, "\n"), strings.Join(goFields, ", "),
		strings.Join(foreignKeyFields, ", "), strings.Join(foreignKeyComps, " AND "), strings.Join(foreignKeyAccess, ", "),
	)

	// generate "join like" queries
	for _, key := range ta.ForeignKeys() {
		fieldName := key.F.Field.Name()
		columnName := sqlColumnName(key.F)
		varName := gen.ToLowerFirst(fieldName)
		keyTypeName := "int64"

		// lookup methods
		if !key.IsNullable() {
			keyTypeName = ctx.typeName(key.F.Field.Type())
			out = append(out, ctx.idArrayConverters(keyTypeName)) // add the converter

			if key.IsUnique {
				content += fmt.Sprintf(`
				// By%[1]s returns a map with '%[1]s' as keys.
				func (items %[2]ss) By%[1]s() map[%[3]s]%[2]s {
					out := make(map[%[3]s]%[2]s, len(items))
					for _, target := range items {
						out[target.%[1]s] = target
					}
					return out
				}`, fieldName, goTypeName, keyTypeName)
			} else {
				content += fmt.Sprintf(`
				// By%[1]s returns a map with '%[1]s' as keys.
				func (items %[2]ss) By%[1]s() map[%[3]s]%[2]ss {
					out := make(map[%[3]s]%[2]ss)
					for _, target := range items {
						out[target.%[1]s] = append(out[target.%[1]s], target)
					}
					return out
				}	
				`, fieldName, goTypeName, keyTypeName)
			}

			content += fmt.Sprintf(`
			// %[1]ss returns the list of ids of %[1]s
			// contained in this link table.
			// They are not garanteed to be distinct.
			func (items %[2]ss) %[1]ss() []%[3]s {
				out := make([]%[3]s, len(items))
				for index, target := range items {
					out[index] = target.%[1]s
				}
				return out
			}
			`, fieldName, goTypeName, keyTypeName)
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
		func Select%[1]ssBy%[2]ss(tx DB, %[3]ss ...%[5]s) (%[1]ss, error) {
			rows, err := tx.Query("SELECT * FROM %[4]s WHERE %[6]s = ANY($1)", %[5]sArrayToPQ(%[3]ss))
			if err != nil {
				return nil, err
			}
			return Scan%[1]ss(rows)
		}

		func Delete%[1]ssBy%[2]ss(tx DB, %[3]ss ...%[5]s) (%[1]ss, error)  {
			rows, err := tx.Query("DELETE FROM %[4]s WHERE %[6]s = ANY($1) RETURNING *", %[5]sArrayToPQ(%[3]ss))
			if err != nil {
				return nil, err
			}
			return Scan%[1]ss(rows)
		}	
		`, goTypeName, fieldName, varName, sqlTableName, keyTypeName, columnName)
	}

	return append(out, gen.Declaration{
		ID:      string(goTypeName),
		Content: content,
	})
}

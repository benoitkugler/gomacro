package sqlcrud

import (
	"fmt"
	"strings"

	"github.com/benoitkugler/gomacro/analysis/sql"
	gen "github.com/benoitkugler/gomacro/generator"
)

func (ctx context) generateLinkTable(ta sql.Table, cols columnsCode) gen.Declaration {
	goTypeName := ta.TableName()
	sqlTableName := gen.SQLTableName(goTypeName)

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
		rows, err := db.Query("SELECT %[9]s FROM %[2]s")
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

	func (item %[1]s) Insert(db DB) error {
		_, err := db.Exec(`+"`"+`INSERT INTO %[2]s (
			%[9]s
			) VALUES (
			%[10]s
			);
			`+"`,"+`%[5]s)
		if err != nil {
			return err
		}
		return nil
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
		cols.goScanFields, cols.sqlQuotedColumnNames, cols.goValueFields,
		strings.Join(foreignKeyFields, ", "), strings.Join(foreignKeyComps, " AND "), strings.Join(foreignKeyAccess, ", "),
		cols.sqlColumnNames, cols.sqlPlaceholders,
	)

	return gen.Declaration{
		ID:      string(goTypeName),
		Content: content,
	}
}

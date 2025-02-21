package sqlcrud

import (
	"fmt"

	"github.com/benoitkugler/gomacro/analysis/sql"
	gen "github.com/benoitkugler/gomacro/generator"
)

func (ctx context) generatePrimaryTable(ta sql.Table, cols columnsCode) []gen.Declaration {
	primaryIndex := ta.Primary()
	idTypeName := ctx.typeName(ta.Columns[primaryIndex].Field.Type.Type())
	goTypeName := ta.TableName()
	sqlTableName := gen.SQLTableName(goTypeName)

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
	rows, err := db.Query("SELECT %[10]s FROM %[3]s")
	if err != nil {
		return nil, err
	}
	return Scan%[1]ss(rows)
}

// Select%[1]s returns the entry matching 'id'.
func Select%[1]s(tx DB, id %[2]s) (%[1]s, error) {
	row := tx.QueryRow("SELECT %[10]s FROM %[3]s WHERE id = $1", id)
	return Scan%[1]s(row)
}

// Select%[1]ss returns the entry matching the given 'ids'.
func Select%[1]ss(tx DB, ids ...%[2]s) (%[1]ss, error) {
	rows, err := tx.Query("SELECT %[10]s FROM %[3]s WHERE id = ANY($1)", %[2]sArrayToPQ(ids))
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
		) RETURNING %[10]s;
		`+"`,"+`%[7]s)
	return Scan%[1]s(row)
}

// Update %[1]s in the database and returns the new version.
func (item %[1]s) Update(tx DB) (out %[1]s, err error) {
	row := tx.QueryRow(`+"`"+`UPDATE %[3]s SET (
		%[5]s
		) = (
		%[6]s
		) WHERE id = $%[8]d RETURNING %[10]s;
		`+"`,"+`%[7]s, item.%[9]s)
	return Scan%[1]s(row)
}

// Deletes the %[1]s and returns the item
func Delete%[1]sById(tx DB, id %[2]s) (%[1]s, error) {
	row := tx.QueryRow("DELETE FROM %[3]s WHERE id = $1 RETURNING %[10]s;", id)
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
		cols.goScanFields, cols.sqlColumnNamesNoPrimary, cols.sqlPlaceholdersNoPrimary, cols.goValueFieldsNoPrimary,
		cols.columnsCount, ta.Columns[primaryIndex].Field.Field.Name(),
		cols.sqlColumnNames,
	)

	var out []gen.Declaration

	// generate "join like" queries
	for _, key := range ta.ForeignKeys() {
		fieldName := key.F.Field.Name()
		columnName := sqlColumnName(key.F)
		varName := gen.ToLowerFirst(fieldName)

		keyTypeName := ctx.typeName(key.TargetIDType())

		// add int array converter if required
		if ctx.generateArrayConverter(key) {
			out = append(out, ctx.idArrayConverters(keyTypeName))
		}

		if key.IsUnique {
			content += fmt.Sprintf(`
			// Select%[1]sBy%[2]s return zero or one item, thanks to a UNIQUE SQL constraint.
			func Select%[1]sBy%[2]s(tx DB, %[3]s %[5]s) (item %[1]s, found bool, err error) {
				row := tx.QueryRow("SELECT %[7]s FROM %[4]s WHERE %[6]s = $1", %[3]s)
				item, err = Scan%[1]s(row)
				if err == sql.ErrNoRows {
					return item, false, nil
				}
				return item, true, err
			}	
			`, goTypeName, fieldName, varName, sqlTableName, keyTypeName, columnName,
				cols.sqlColumnNames,
			)
		}

		if !key.IsNullable() {
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
						dict := out[target.%[1]s]
						if dict == nil {
							dict = make(%[2]ss)
						}
						dict[target.Id] = target
						out[target.%[1]s] = dict
					}
					return out
				}	
				`, fieldName, goTypeName, keyTypeName)
			}

			content += fmt.Sprintf(`
			// %[1]ss returns the list of ids of %[1]s
			// contained in this table.
			// They are not garanteed to be distinct.
			func (items %[2]ss) %[1]ss() []%[3]s {
				out := make([]%[3]s, 0, len(items))
				for _, target := range items {
					out = append(out, target.%[1]s)
				}
				return out
			}
			`, fieldName, goTypeName, keyTypeName)
		}

		varNamePlural := varName + "s_" // to avoid potential shadowing
		content += fmt.Sprintf(`
		func Select%[1]ssBy%[2]ss(tx DB, %[3]s ...%[6]s) (%[1]ss, error) {
			rows, err := tx.Query("SELECT %[8]s FROM %[4]s WHERE %[7]s = ANY($1)", %[6]sArrayToPQ(%[3]s))
			if err != nil {
				return nil, err
			}
			return Scan%[1]ss(rows)
		}	

		func Delete%[1]ssBy%[2]ss(tx DB, %[3]s ...%[6]s) ([]%[5]s, error) {
			rows, err := tx.Query("DELETE FROM %[4]s WHERE %[7]s = ANY($1) RETURNING id", %[6]sArrayToPQ(%[3]s))
			if err != nil {
				return nil, err
			}
			return Scan%[5]sArray(rows)
		}	
		`, goTypeName, fieldName, varNamePlural, sqlTableName, idTypeName, keyTypeName, columnName,
			cols.sqlColumnNames,
		)
	}

	return append(out,
		ctx.idArrayConverters(idTypeName),
		gen.Declaration{
			ID:      string(goTypeName),
			Content: content,
		},
	)
}

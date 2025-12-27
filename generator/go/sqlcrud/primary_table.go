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

	return []gen.Declaration{
		ctx.idArrayConverters(idTypeName),
		{
			ID:      string(goTypeName),
			Content: content,
		},
	}
}

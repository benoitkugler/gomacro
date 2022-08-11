// Package sqlcrud generate Go functions
// to read and write from a DB defined using
// the conventions from generator/sql
package sqlcrud

import (
	"fmt"
	"go/types"
	"strings"

	an "github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/analysis/sql"
	"github.com/benoitkugler/gomacro/generator"
	gen "github.com/benoitkugler/gomacro/generator"
)

var jsonValuer = gen.Declaration{
	ID: "__json_valuer",
	Content: `
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
	`,
}

// overriden in tests
var pqImportPath = "github.com/lib/pq"

func Generate(ana *an.Analysis) []gen.Declaration {
	header := fmt.Sprintf(`
	package %s

	// Code generated by gomacro/generator/go/sqlcrud. DO NOT EDIT.

	import (
		"database/sql"
		"%s"
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
	`, ana.Root.Types.Name(), pqImportPath)

	decls := []gen.Declaration{
		{ID: "aa__header", Content: header, Priority: true},
	}

	tables := sql.SelectTables(ana)

	ctx := context{target: ana.Root.Types}
	for _, ta := range tables {
		decls = append(decls, ctx.generateTable(ta)...)
	}

	return decls
}

type context struct {
	target *types.Package
}

func (ctx context) typeName(ty types.Type) string {
	return types.TypeString(ty, generator.NameRelativeTo(ctx.target))
}

// returns true if `ty` is named and defined the target package,
// or is int64
func (ctx context) generateArrayConverter(key sql.ForeignKey) bool {
	ty := key.F.Type.Type()
	if ctx.typeName(ty) == "int64" {
		return true
	}
	if named, isNamed := ty.(*types.Named); isNamed {
		return named.Obj().Pkg() == ctx.target
	}
	return false
}

// return `true` if the column is backed by a named type belonging
// to the target package, along with its local name.
// If not, it return false, assuming that the required methods are already implemented.
func (ctx context) canImplementValuer(column sql.Column) (string, bool) {
	named, ok := column.Field.Type.Type().(*types.Named)
	if !ok {
		panic(fmt.Sprintf("field %s, written as JSON in SQL, is not named: sql.Valuer interface can't be implemented", column.Field.Field.Name()))
	}
	goTypeName := named.Obj().Name()
	return goTypeName, named.Obj().Pkg().Path() == ctx.target.Path()
}

func (ctx context) idArrayConverters(idTypeName string) gen.Declaration {
	out := gen.Declaration{
		ID: "id_array_converter_" + idTypeName,
	}
	if idTypeName == "int64" { // no copy required
		out.Content = fmt.Sprintf(`
		func %[1]sArrayToPQ(ids []%[1]s) pq.Int64Array { return ids }
		`, idTypeName)
	} else {
		out.Content = fmt.Sprintf(`
		func %[1]sArrayToPQ(ids []%[1]s) pq.Int64Array {
			out := make(pq.Int64Array, len(ids))
			for i, v := range ids {
				out[i] = int64(v)
			}
			return out
		}
		`, idTypeName)
	}

	out.Content += fmt.Sprintf(`
	// Scan%[1]sArray scans the result of a query returning a
	// list of ID's.
	func Scan%[1]sArray(rs *sql.Rows) ([]%[1]s, error) {
		defer rs.Close()
		ints := make([]%[1]s, 0, 16)
		var err error
		for rs.Next() {
			var s %[1]s
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
	`, idTypeName)

	// also add Set utility
	out.Content += fmt.Sprintf(`
	type %[1]sSet map[%[1]s]bool 

	func New%[1]sSetFrom(ids []%[1]s) %[1]sSet { 
		out := make(%[1]sSet, len(ids))
		for _, key := range ids {
			out[key] = true
		}
		return out
	}

	func (s %[1]sSet) Add(id %[1]s) { s[id] = true }

	func (s %[1]sSet) Has(id %[1]s) bool { return s[id] }

	func (s %[1]sSet) Keys() []%[1]s {
		out := make([]%[1]s, 0, len(s))
		for k := range s {
			out = append(out, k)
		}
		return out
	}
	`, idTypeName)

	return out
}

func isUnderlyingTime(ty an.Type) bool {
	if named, isNamed := ty.(*an.Named); isNamed {
		ty = named.Underlying
	}
	_, isTime := ty.(*an.Time)
	return isTime
}

func (ctx context) generateTable(ta sql.Table) (decls []gen.Declaration) {
	if ta.Primary() >= 0 { // we have an ID
		decls = append(decls, ctx.generatePrimaryTable(ta)...)
	} else { // "link" table
		decls = append(decls, ctx.generateLinkTable(ta)...)
	}

	decls = append(decls, ctx.generateSelectByUniques(ta))

	// generate the value interface method
	for _, col := range ta.Columns {
		if isUnderlyingTime(col.Field.Type) {
			goTypeName, isLocal := ctx.canImplementValuer(col)
			if isLocal {
				decls = append(decls, gen.Declaration{
					ID: "datetime_value" + goTypeName,
					Content: fmt.Sprintf(`
					func (s *%s) Scan(src interface{}) error {
						var tmp pq.NullTime
						err := tmp.Scan(src)
						if err != nil {
							return err
						}
						*s = %s(tmp.Time)
						return nil
					}
		
					func (s %s) Value() (driver.Value, error) {
						return pq.NullTime{Time: time.Time(s), Valid: true}.Value()
					}
					`, goTypeName, goTypeName, goTypeName),
				})
			}
		} else if arr, isArray := col.SQLType.(sql.Array); isArray {
			goTypeName, isLocal := ctx.canImplementValuer(col)
			if isLocal {
				var pqType string
				switch arr.A.Elem.(*an.Basic).Kind() {
				case an.BKBool:
					pqType = "pq.BoolArray"
				case an.BKInt:
					pqType = "pq.Int64Array"
				case an.BKFloat:
					pqType = "pq.Float64Array"
				case an.BKString:
					pqType = "pq.StringArray"
				}
				decls = append(decls, gen.Declaration{
					ID: "array_value" + goTypeName,
					Content: fmt.Sprintf(`
						func (s *%s) Scan(src interface{}) error  { return (*%s)(s).Scan(src) }
						func (s %s) Value() (driver.Value, error) { return %s(s).Value() }
						`, goTypeName, pqType, goTypeName, pqType),
				})
			}
		} else if _, isJSON := col.SQLType.(sql.JSON); isJSON {
			goTypeName, isLocal := ctx.canImplementValuer(col)
			if isLocal {
				decls = append(decls, gen.Declaration{
					ID: "json_value" + goTypeName,
					Content: fmt.Sprintf(`
						func (s *%s) Scan(src interface{}) error { return loadJSON(s, src) }
						func (s %s) Value() (driver.Value, error) { return dumpJSON(s) }
						`, goTypeName, goTypeName),
				}, jsonValuer)
			}
		}
	}

	return decls
}

func sqlColumnName(fi an.StructField) string {
	return strings.ToLower(fi.Field.Name())
}

func (ctx context) generateSelectByUniques(ta sql.Table) gen.Declaration {
	var content string
	goTypeName := ta.TableName()
	sqlTableName := gen.SQLTableName(goTypeName)
	for _, cols := range ta.AdditionalUniqueCols() {
		comparison := columsComparison(cols)
		funcTitle := columsFuncTitle(cols)
		varNames, varDecls := ctx.columsVarDecls(cols)

		content += fmt.Sprintf(`
		// Select%[1]sBy%[2]s return zero or one item, thanks to a UNIQUE SQL constraint.
		func Select%[1]sBy%[2]s(tx DB, %[3]s) (item %[1]s, found bool, err error) {
			row := tx.QueryRow("SELECT * FROM %[4]s WHERE %[5]s", %[6]s)
			item, err = Scan%[1]s(row)
			if err == sql.ErrNoRows {
				return item, false, nil
			}
			return item, true, err
		}
		`, goTypeName, funcTitle, varDecls, sqlTableName, comparison, varNames)
	}

	return gen.Declaration{
		ID:      string(goTypeName) + "_unique_selects",
		Content: content,
	}
}

// assume placholders are $1, $2, etc...
func columsComparison(cols []sql.Column) string {
	chunks := make([]string, len(cols))
	for i, c := range cols {
		chunks[i] = fmt.Sprintf("%s = $%d", c.Field.Field.Name(), i+1)
	}
	return strings.Join(chunks, " AND ")
}

func columsFuncTitle(cols []sql.Column) string {
	chunks := make([]string, len(cols))
	for i, c := range cols {
		chunks[i] = c.Field.Field.Name()
	}
	return strings.Join(chunks, "And")
}

func (ctx context) columsVarDecls(cols []sql.Column) (string, string) {
	varNames := make([]string, len(cols))
	varDecls := make([]string, len(cols))
	for i, c := range cols {
		varName := gen.ToLowerFirst(c.Field.Field.Name())
		varNames[i] = varName
		varDecls[i] = fmt.Sprintf("%s %s", varName, ctx.typeName(c.Field.Field.Type()))
	}
	return strings.Join(varNames, ", "), strings.Join(varDecls, ", ")
}

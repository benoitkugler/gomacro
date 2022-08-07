// Package sqlcrud generate Go functions
// to read and write from a DB defined using
// the conventions from generator/sql
package sqlcrud

import (
	"fmt"
	"go/types"

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

// returns true if `ty` is named and defined the target package
func (ctx context) isNamedLocal(ty types.Type) bool {
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

	func (s %[1]sSet) Add(id %[1]s) { s[id] = true }

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

func (ctx context) generateTable(ta sql.Table) (decls []gen.Declaration) {
	if ta.Primary() >= 0 { // we have an ID
		decls = append(decls, ctx.generatePrimaryTable(ta)...)
	} else { // "link" table
		decls = append(decls, ctx.generateLinkTable(ta)...)
	}

	// generate the value interface method
	for _, col := range ta.Columns {
		if _, isTime := col.Field.Type.(*an.Time); isTime {
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

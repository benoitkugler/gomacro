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

// returns true if `ty` is named and defined in the target package,
// or is int64
func (ctx context) generateArrayConverter(key sql.ForeignKey) bool {
	ty := key.TargetIDType()
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
		panic(fmt.Sprintf("field %s (with type %T), is not named: sql.Valuer interface can't be implemented", column.Field.Field.Name(), column.SQLType))
	}
	goTypeName := named.Obj().Name()
	targetPath := ctx.target.Path()
	columnPath := ""
	if pkg := named.Obj().Pkg(); pkg != nil {
		columnPath = pkg.Path()
	}
	return goTypeName, columnPath == targetPath
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

func isUnderlyingTime(ty an.Type) (*an.Time, bool) {
	if named, isNamed := ty.(*an.Named); isNamed {
		ty = named.Underlying
	}
	t, isTime := ty.(*an.Time)
	return t, isTime
}

// return the field name of sql.NullInt64 like types
func isLocalNullInt64(col sql.Column) *types.Var {
	named, ok := col.Field.Type.Type().(*types.Named)
	if !ok {
		return nil
	}

	if field := sql.IsNullXXX(named); field != nil && sql.IsInt64(field.Type()) {
		return field
	}

	return nil
}

func (ctx context) generateTable(ta sql.Table) (decls []gen.Declaration) {
	code := newColumnsCode(ta)
	if ta.Primary() >= 0 { // we have an ID
		decls = append(decls, ctx.generatePrimaryTable(ta, code)...)
	} else { // "link" table
		decls = append(decls, ctx.generateLinkTable(ta, code)...)
	}

	decls = append(decls, ctx.generateSelectByUniques(ta, code))
	decls = append(decls, ctx.generateSelectByKeys(ta, code))

	// generate the value interface method
	for _, col := range ta.Columns {
		if ty, ok := isUnderlyingTime(col.Field.Type); ok {
			goTypeName, isLocal := ctx.canImplementValuer(col)
			if isLocal {
				if ty.IsDate {
					decls = append(decls, gen.Declaration{
						ID: "date_value" + goTypeName,
						Content: fmt.Sprintf(`
						func (s *%s) Scan(src interface{}) error {
							var tmp pq.NullTime
							err := tmp.Scan(src)
							if err != nil {
								return err
							}
							*s = NewDateFrom(tmp.Time)
							return nil
						}
			
						func (s %s) Value() (driver.Value, error) {
							return pq.NullTime{Time: s.Time(), Valid: true}.Value()
						}
						`, goTypeName, goTypeName),
					})
				} else {
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
			}
		} else if arr, isArray := col.SQLType.(sql.Array); isArray {
			goTypeName, isLocal := ctx.canImplementValuer(col)
			if isLocal {
				decls = append(decls, ctx.arrayConverters(goTypeName, arr))
			}
		} else if composite, isComposite := col.SQLType.(sql.Composite); isComposite {
			goTypeName, isLocal := ctx.canImplementValuer(col)
			if isLocal {
				decls = append(decls, ctx.compositeConverters(goTypeName, composite), jsonValuer)
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
		} else if field := isLocalNullInt64(col); field != nil {
			goTypeName, isLocal := ctx.canImplementValuer(col)
			if isLocal {
				decls = append(decls, ctx.optionnalIdConverters(goTypeName, field))
			}
		}
	}

	return decls
}

func sqlColumnName(fi an.StructField) string {
	return strings.ToLower(fi.Field.Name())
}

func (ctx context) generateSelectByUniques(ta sql.Table, code columnsCode) gen.Declaration {
	var content string
	goTypeName := ta.TableName()
	sqlTableName := gen.SQLTableName(goTypeName)
	for _, cols := range ta.AdditionalUniqueCols() {
		comparison := columnsComparison(cols)
		funcTitle := columsFuncTitle(cols)
		varNames, varDecls := ctx.columsVarDecls(cols)

		content += fmt.Sprintf(`
		// Select%[1]sBy%[2]s return zero or one item, thanks to a UNIQUE SQL constraint.
		func Select%[1]sBy%[2]s(tx DB, %[3]s) (item %[1]s, found bool, err error) {
			row := tx.QueryRow("SELECT %[7]s FROM %[4]s WHERE %[5]s", %[6]s)
			item, err = Scan%[1]s(row)
			if err == sql.ErrNoRows {
				return item, false, nil
			}
			return item, true, err
		}
		`, goTypeName, funcTitle, varDecls, sqlTableName, comparison, varNames,
			code.sqlColumnNames,
		)
	}

	return gen.Declaration{
		ID:      string(goTypeName) + "_unique_selects",
		Content: content,
	}
}

func (ctx context) generateSelectByKeys(ta sql.Table, code columnsCode) gen.Declaration {
	var content string
	goTypeName := ta.TableName()
	sqlTableName := gen.SQLTableName(goTypeName)
	for _, cols := range ta.SelectKeys() {
		comparison := columnsComparison(cols)
		funcTitle := columsFuncTitle(cols)
		varNames, varDecls := ctx.columsVarDecls(cols)

		content += fmt.Sprintf(`
		// Select%[1]ssBy%[2]s selects the items matching the given fields.
		func Select%[1]ssBy%[2]s(tx DB, %[3]s) (item %[1]ss, err error) {
			rows, err := tx.Query("SELECT %[7]s FROM %[4]s WHERE %[5]s", %[6]s)
			if err != nil {
				return nil, err
			}
			return Scan%[1]ss(rows)
		}

		// Delete%[1]ssBy%[2]s deletes the item matching the given fields, returning 
		// the deleted items.
		func Delete%[1]ssBy%[2]s(tx DB, %[3]s) (item %[1]ss, err error) {
			rows, err := tx.Query("DELETE FROM %[4]s WHERE %[5]s RETURNING %[7]s", %[6]s)
			if err != nil {
				return nil, err
			}
			return Scan%[1]ss(rows)
		}
		`, goTypeName, funcTitle, varDecls, sqlTableName, comparison, varNames,
			code.sqlColumnNames)
	}

	return gen.Declaration{
		ID:      string(goTypeName) + "_delete_by_keys",
		Content: content,
	}
}

// assume placholders are $1, $2, etc...
func columnsComparison(cols []sql.Column) string {
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

func (ctx context) optionnalIdConverters(goTypeName string, field *types.Var) gen.Declaration {
	return gen.Declaration{
		ID: "nullable_valuer" + goTypeName,
		Content: fmt.Sprintf(`
			func (s *%[1]s) Scan(src interface{}) error {
				var tmp sql.NullInt64
				err := tmp.Scan(src)
				if err != nil {
					return err
				}
				*s = %[1]s{
					Valid: tmp.Valid,
					%[2]s: %[3]s(tmp.Int64),
				}
				return nil
			}
			
			func (s %[1]s) Value() (driver.Value, error) {
				return sql.NullInt64{
					Int64: int64(s.%[2]s), 
					Valid: s.Valid}.Value()
			}
			`, goTypeName, field.Name(), ctx.typeName(field.Type())),
	}
}

func (ctx context) compositeConverters(goTypeName string, composite sql.Composite) gen.Declaration {
	st := composite.Type().(*an.Struct)
	placholders, selectors := make([]string, len(st.Fields)), make([]string, len(st.Fields))
	scanFields := make([]string, len(st.Fields))
	for i, field := range st.Fields {
		name := field.Field.Name()
		placholders[i] = "%d"
		selectors[i] = "s." + name
		scanFields[i] = fmt.Sprintf(`
			val%s, err := strconv.Atoi(fields[%d])
			if err != nil {
				return err
			}
			s.%s = %s(val%s)
		`, name, i, name, ctx.typeName(field.Field.Type()), name)
	}

	return gen.Declaration{
		ID: "composite_value" + goTypeName,
		Content: fmt.Sprintf(`
			func (s *%s) Scan(src interface{}) error { 
				bs, ok := src.([]byte)
				if !ok {
					return fmt.Errorf("unsupported type %%T", src)
				}
				fields := strings.Split(string(bs[1:len(bs)-1]), ",")
				if len(fields) != %d {
					return fmt.Errorf("unsupported number of fields %%d", len(fields))
				}
				%s
				return nil 
			}
			func (s %s) Value() (driver.Value, error) { 
				bs := fmt.Appendf(nil, "(%s)", %s)
				return driver.Value(bs), nil
			}
			`, goTypeName, len(st.Fields), strings.Join(scanFields, "\n"),
			goTypeName, strings.Join(placholders, ", "), strings.Join(selectors, ", ")),
	}
}

// elems are Basic or (integer) Enums
func (ctx context) arrayConverters(goTypeName string, arr sql.Array) gen.Declaration {
	var (
		pqType             string
		elemUnderlyingType *types.Basic
		elemKind           an.BasicKind

		elemType types.Type // noy nil if conversion is required
	)

	if basic, isBasic := arr.A.Elem.(*an.Basic); isBasic {
		elemUnderlyingType = basic.B
		elemKind = basic.Kind()
	} else if enum, isEnum := arr.A.Elem.(*an.Enum); isEnum {
		elemType = enum.Type()
		elemUnderlyingType = enum.Underlying()
		elemKind = enum.Kind()
	} else {
		panic("unsupported array element")
	}

	switch elemKind {
	case an.BKBool:
		pqType = "pq.BoolArray"
	case an.BKInt:
		switch elemUnderlyingType.Kind() {
		case types.Int64, types.Uint64, types.Int, types.Uint:
			pqType = "pq.Int64Array"
		default:
			pqType = "pq.Int32Array"
		}
	case an.BKFloat:
		pqType = "pq.Float64Array"
	case an.BKString:
		pqType = "pq.StringArray"
	}
	isGoArray := arr.A.Len >= 0

	var scanCode, valueCode string
	if isGoArray { // convert from slice to array
		scanCode = fmt.Sprintf(`var tmp %s
					err := tmp.Scan(src)
					if err != nil {
						return err
					}
					if len(tmp) != %d {
						return fmt.Errorf("unexpected length %%d", len(tmp))
					}
					copy(s[:], tmp)
					return nil
					`, pqType, arr.A.Len)
		valueCode = fmt.Sprintf("return %s(s[:]).Value()", pqType)
	} else {
		if elemType != nil { // we have to build a temporary array
			pqElemType := strings.ToLower(strings.TrimSuffix(strings.TrimPrefix(pqType, "pq."), "Array"))

			scanCode = fmt.Sprintf(`var tmp %s
					err := tmp.Scan(src)
					if err != nil {
						return err
					}
					*s = make([]%s, len(tmp))
					for i, v := range tmp {
						(*s)[i] = %s(v)
					}
					return nil`,
				pqType, ctx.typeName(elemType), ctx.typeName(elemType))
			valueCode = fmt.Sprintf(`tmp := make(%s, len(s))
				for i, v := range s {
						tmp[i] = %s(v)
					}
				return tmp.Value()`,
				pqType, pqElemType)
		} else { // nothing to do, nice !
			scanCode = fmt.Sprintf("return (*%s) (s).Scan(src)", pqType)
			valueCode = fmt.Sprintf("return %s(s).Value()", pqType)
		}
	}

	return gen.Declaration{
		ID: "array_value" + goTypeName,
		Content: fmt.Sprintf(`
		func (s *%s) Scan(src interface{}) error  { 
			%s
		}
		func (s %s) Value() (driver.Value, error) { 
			%s
		}`, goTypeName, scanCode, goTypeName, valueCode),
	}
}

// columnsCode exposes various SQL or Go code
// generated from columns
type columnsCode struct {
	// required for create/update statements

	goScanFields                          string
	goValueFields, goValueFieldsNoPrimary string

	// required for insert statements

	sqlQuotedColumnNames                      string
	sqlColumnNames, sqlColumnNamesNoPrimary   string
	sqlPlaceholders, sqlPlaceholdersNoPrimary string
}

func newColumnsCode(ta sql.Table) columnsCode {
	var (
		scanFields        = make([]string, len(ta.Columns))
		valueFields       = make([]string, len(ta.Columns))
		quotedColumnNames = make([]string, len(ta.Columns))
		columnNames       = make([]string, len(ta.Columns))
		placeholders      = make([]string, len(ta.Columns))

		valueFieldsNoPrimary  []string
		columnNamesNoPrimary  []string
		placeholdersNoPrimary []string
	)
	primaryIndex := ta.Primary()
	for i, col := range ta.Columns {
		fieldName := col.Field.Field.Name()
		columnName := sqlColumnName(col.Field)
		scanFields[i] = fmt.Sprintf("&item.%s,", fieldName)
		valueFields[i] = fmt.Sprintf("item.%s", fieldName)

		quotedColumnNames[i] = fmt.Sprintf("%q,", columnName)
		columnNames[i] = columnName
		placeholders[i] = fmt.Sprintf("$%d", i+1)

		if i != primaryIndex {
			valueFieldsNoPrimary = append(valueFieldsNoPrimary, fmt.Sprintf("item.%s", fieldName))
			columnNamesNoPrimary = append(columnNamesNoPrimary, columnName)
			// placeholders like $1 $2 ...
			placeholdersNoPrimary = append(placeholdersNoPrimary, fmt.Sprintf("$%d", len(placeholdersNoPrimary)+1))
		}
	}

	return columnsCode{
		goScanFields:         strings.Join(scanFields, "\n"),
		goValueFields:        strings.Join(valueFields, ", "),
		sqlQuotedColumnNames: strings.Join(quotedColumnNames, "\n"),
		sqlColumnNames:       strings.Join(columnNames, ", "),
		sqlPlaceholders:      strings.Join(placeholders, ", "),

		goValueFieldsNoPrimary:   strings.Join(valueFieldsNoPrimary, ", "),
		sqlColumnNamesNoPrimary:  strings.Join(columnNamesNoPrimary, ", "),
		sqlPlaceholdersNoPrimary: strings.Join(placeholdersNoPrimary, ", "),
	}
}

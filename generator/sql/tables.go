package sql

import (
	"fmt"
	"regexp"
	"strings"

	an "github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/analysis/sql"
	gen "github.com/benoitkugler/gomacro/generator"
)

const (
	prefixDeclHeader         = "aa_"
	prefixDeclCreate         = "ab_"
	prefixDeclConstraint     = "ac_"
	prefixDeclJSONConstraint = "zz_"
	prefixDeclCompositeType  = "aaa_"
)

// Generate returns the SQL statements required to create
// the tables contained in `Source`, defined by Go structs.
func Generate(ana *an.Analysis) []gen.Declaration {
	var decls []gen.Declaration

	tables := sql.SelectTables(ana)

	nameReplacer := gen.NewTableNameReplacer(tables)

	var constraints []string
	for _, ta := range tables {

		// table creation / JSON validations
		decls = append(decls, generateTable(ta)...)

		// explicit (user provided) constraints
		for _, constraint := range ta.CustomConstraints {
			constraints = append(constraints, generateCustomConstraint(ana, ta, nameReplacer, constraint))
		}

		// implicit constraints (like foreign keys)
		for _, foreign := range ta.ForeignKeys() {
			constraints = append(constraints, generateForeignConstraint(ta.TableName(), foreign))
		}

		for _, column := range ta.Columns {
			if value, ok := column.Field.IsSQLGuard(); ok {
				constraints = append(constraints, generateQuardConstraint(ana, ta, column, value)...)
			}
		}
	}

	decls = append(decls,
		gen.Declaration{
			ID:       prefixDeclHeader + "header",
			Content:  "-- Code genererated by gomacro/generator/sql. DO NOT EDIT.",
			Priority: true,
		},
		gen.Declaration{
			ID:       prefixDeclConstraint + "constraints",
			Content:  "-- constraints\n" + strings.Join(constraints, "\n"),
			Priority: true,
		})

	return decls
}

func generateQuardConstraint(ana *an.Analysis, ta sql.Table, column sql.Column, value string) []string {
	value = gen.ReplaceEnums(ana, value)
	return []string{
		fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s;", gen.SQLTableName(ta.TableName()), column.Field.Field.Name(), value),
		fmt.Sprintf("ALTER TABLE %s ADD CHECK(%s = %s);", gen.SQLTableName(ta.TableName()), column.Field.Field.Name(), value),
	}
}

func generateForeignConstraint(sourceTable sql.TableName, fk sql.ForeignKey) string {
	onDelete := ""
	if action := fk.OnDelete(); action != "" {
		onDelete = "ON DELETE " + action
	}
	return fmt.Sprintf("ALTER TABLE %s ADD FOREIGN KEY(%s) REFERENCES %s %s;",
		gen.SQLTableName(sourceTable), fk.F.Field.Name(), gen.SQLTableName(fk.Target), onDelete)
}

var reReferences = regexp.MustCompile(`REFERENCES (\w+)`)

func generateCustomConstraint(ana *an.Analysis, ta sql.Table, rep gen.TableNameReplacer, content string) string {
	content = reReferences.ReplaceAllStringFunc(content, func(s string) string {
		ref, name, _ := strings.Cut(s, " ")
		return ref + " " + gen.SQLTableName(sql.TableName(name))
	})

	content = rep.Replace(content)
	content = gen.ReplaceEnums(ana, content)

	if strings.HasPrefix(content, "ADD") {
		return fmt.Sprintf("ALTER TABLE %s %s;", gen.SQLTableName(ta.TableName()), content)
	}

	return fmt.Sprintf("%s;", content)
}

func generateTable(ta sql.Table) []gen.Declaration {
	var (
		decls    []gen.Declaration
		colTypes = make([]string, len(ta.Columns))
	)
	tablePkg := ta.Name.Obj().Pkg()
	for i, f := range ta.Columns {
		statement, colDecls := createStmt(f, ta.Primary() == i)

		colTypes[i] = "\t" + statement

		// add the composite type decl
		if composite, isComposite := f.SQLType.(sql.Composite); isComposite {
			isLocal := tablePkg == composite.Type().(*an.Struct).Name.Obj().Pkg()
			if isLocal {
				decls = append(decls, gen.Declaration{
					ID:       prefixDeclCompositeType + composite.Name(),
					Content:  compositeDecl(composite),
					Priority: true,
				})
			}
		}

		// add the eventual JSON validation function
		if js, isJSON := f.SQLType.(sql.JSON); isJSON {
			jsonDecls, jsonFuncName := jsonValidations(js)

			colName := f.Field.Field.Name()
			id := prefixDeclJSONConstraint + jsonFuncName + gen.SQLTableName(ta.TableName()) + colName
			decls = append(decls, jsonDecls...)
			decls = append(decls, gen.Declaration{
				ID:       id,
				Content:  fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s_gomacro CHECK (%s(%s));", gen.SQLTableName(ta.TableName()), colName, jsonFuncName, colName),
				Priority: false,
			})
		}

		decls = append(decls, colDecls...)
	}

	tableName := gen.SQLTableName(ta.TableName())

	decl := gen.Declaration{
		ID: prefixDeclCreate + ta.Name.Obj().String(),
		Content: fmt.Sprintf(`
		CREATE TABLE %s (
		%s
		);`, tableName, strings.Join(colTypes, ",\n")),
		Priority: true,
	}

	return append(decls, decl)
}

func createStmt(col sql.Column, isPrimary bool) (string, []gen.Declaration) {
	var (
		typeDecl string
		decls    []gen.Declaration
	)
	if isPrimary { // special case
		typeDecl = "serial PRIMARY KEY"
	} else {
		ct := typeConstraint(col)
		typeDecl = col.SQLType.Name() + " " + ct
	}

	colName := col.Field.Field.Name()

	// we defer json and foreign contraints in separate declaration
	// including used provided constraints
	return fmt.Sprintf("%s %s", colName, typeDecl), decls
}

func typeConstraint(field sql.Column) string {
	switch ty := field.SQLType.(type) {
	case sql.Builtin:
		if ty.IsNullable() {
			return ""
		}
		return "NOT NULL"
	case sql.Enum:
		return fmt.Sprintf(" CHECK (%s IN %s) NOT NULL", field.Field.Field.Name(), enumTuple(ty.E))
	case sql.Array:
		if L := ty.A.Len; L >= 0 {
			return fmt.Sprintf(" CHECK (array_length(%s, 1) = %d) NOT NULL", field.Field.Field.Name(), L)
		}
		return ""
	case sql.Composite:
		return "NOT NULL"
	case sql.JSON:
		return "NOT NULL"
	default:
		panic(sql.ExhaustiveSQLTypeSwitch)
	}
}

// enumTuple returns a tuple of valid values
// compatible with SQL syntax.
// Ex: ('red', 'blue', 'green')
func enumTuple(e *an.Enum) string {
	chunks := make([]string, len(e.Members))
	for i, val := range e.Members {
		chunks[i] = val.Const.Val().ExactString()
	}
	out := fmt.Sprintf("(%s)", strings.Join(chunks, ", "))
	return strings.ReplaceAll(out, `"`, `'`) // SQL uses single quote
}

func compositeDecl(cp sql.Composite) string {
	st := cp.Type().(*an.Struct)
	fields := make([]string, len(st.Fields))
	for i, field := range st.Fields {
		fields[i] = fmt.Sprintf("%s %s", field.JSONName(), cp.SQLType(i).Name())
	}
	return fmt.Sprintf("CREATE TYPE %s AS (%s);", cp.Name(), strings.Join(fields, ", "))
}

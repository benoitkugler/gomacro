package sql

import (
	"fmt"
	"strings"

	an "github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/analysis/sql"
	gen "github.com/benoitkugler/gomacro/generator"
)

const (
	prefixDeclCreate     = "aa_"
	prefixDeclConstraint = "ab_"
)

// Generate returns the SQL statements required to create
// the tables contained in `Source`, defined by Go structs.
func Generate(ana *an.Analysis) []gen.Declaration {
	var decls []gen.Declaration

	tables := sql.SelectTables(ana)

	nameReplacer := tableNameReplacer(tables)

	var constraints []string
	for _, ta := range tables {
		// table creation / JSON validations
		decls = append(decls, generateTable(ta)...)

		// implicit constraints (like foreign keys)
		for _, foreign := range ta.ForeignKeys() {
			constraints = append(constraints, generateForeignConstraint(ta.TableName(), foreign))
		}

		// explicit (user provided) constraints
		for _, constraint := range ta.CustomConstraints {
			constraints = append(constraints, generateCustomConstraint(nameReplacer, ta.TableName(), constraint))
		}
	}

	decls = append(decls, gen.Declaration{
		ID:       prefixDeclConstraint + "constraints",
		Content:  "-- constraints\n" + strings.Join(constraints, "\n"),
		Priority: true,
	})

	return decls
}

func tableNameReplacer(tables []sql.Table) *strings.Replacer {
	var reps []string
	for _, ta := range tables {
		name := ta.TableName()
		reps = append(reps, string(name), gen.SQLTableName(name))
	}
	return strings.NewReplacer(reps...)
}

func generateForeignConstraint(sourceTable sql.TableName, fk sql.ForeignKey) string {
	onDelete := ""
	if action := fk.OnDelete(); action != "" {
		onDelete = "ON DELETE " + action
	}
	return fmt.Sprintf("ALTER TABLE %s ADD FOREIGN KEY(%s) REFERENCES %s %s;",
		gen.SQLTableName(sourceTable), fk.F.Field.Name(), gen.SQLTableName(fk.Target), onDelete)
}

func generateCustomConstraint(rep *strings.Replacer, sourceTable sql.TableName, content string) string {
	content = rep.Replace(content)
	return fmt.Sprintf("ALTER TABLE %s %s;", gen.SQLTableName(sourceTable), content)
}

func generateTable(ta sql.Table) []gen.Declaration {
	var (
		decls    []gen.Declaration
		colTypes = make([]string, len(ta.Columns))
	)
	for i, f := range ta.Columns {
		statement, colDecls := createStmt(f, ta.Primary() == i)

		colTypes[i] = "\t" + statement
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

	// add the eventual JSON validation function
	if js, isJSON := col.SQLType.(sql.JSON); isJSON {
		var jsonFuncName string
		decls, jsonFuncName = jsonValidations(js)
		typeDecl += fmt.Sprintf(" CONSTRAINT %s_%s CHECK (%s(%s))", colName, jsonFuncName, jsonFuncName, colName)
	}

	// we defer foreign contraints in separate declaration
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
		chunks[i] = val.Const.Val().String()
	}
	out := fmt.Sprintf("(%s)", strings.Join(chunks, ", "))
	return strings.ReplaceAll(out, `"`, `'`) // SQL uses single quote
}

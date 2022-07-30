package sql

import (
	"fmt"
	"strings"

	an "github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/analysis/sql"
	gen "github.com/benoitkugler/gomacro/generator"
)

// convertTableName uses a familiar SQL convention
func convertTableName(name string) string {
	return gen.ToSnakeCase(name) + "s"
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

	tableName := convertTableName(ta.Name.Obj().Name())

	decl := gen.Declaration{
		ID: ta.Name.Obj().String(),
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

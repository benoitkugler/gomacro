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

		// explicit (user provided) constraints
		for _, constraint := range ta.CustomConstraints {
			constraints = append(constraints, generateCustomConstraint(ana, ta, nameReplacer, constraint))
		}

		// implicit constraints (like foreign keys)
		for _, foreign := range ta.ForeignKeys() {
			constraints = append(constraints, generateForeignConstraint(ta.TableName(), foreign))
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

type replacer map[string]string

func tableNameReplacer(tables []sql.Table) replacer {
	out := make(replacer)
	for _, ta := range tables {
		name := ta.TableName()
		out[string(name)] = gen.SQLTableName(name)
	}
	return out
}

var reWords = regexp.MustCompile(`([\w]+)`)

// Replace replace words occurences
func (rp replacer) Replace(input string) string {
	return reWords.ReplaceAllStringFunc(input, func(word string) string {
		if subs, has := rp[word]; has {
			return subs
		}
		return word
	})
}

func generateForeignConstraint(sourceTable sql.TableName, fk sql.ForeignKey) string {
	onDelete := ""
	if action := fk.OnDelete(); action != "" {
		onDelete = "ON DELETE " + action
	}
	return fmt.Sprintf("ALTER TABLE %s ADD FOREIGN KEY(%s) REFERENCES %s %s;",
		gen.SQLTableName(sourceTable), fk.F.Field.Name(), gen.SQLTableName(fk.Target), onDelete)
}

var reEnums = regexp.MustCompile(`#\[(\w+)\.(\w+)\]`)

func generateCustomConstraint(ana *an.Analysis, ta sql.Table, rep replacer, content string) string {
	content = rep.Replace(content)
	// replace enum values
	content = reEnums.ReplaceAllStringFunc(content, func(s string) string {
		s = s[2 : len(s)-1] // trim starting #[ and leading ]
		typeName, varName, _ := strings.Cut(s, ".")
		enum := ana.GetByName(ta.Name, typeName).(*an.Enum)
		enumValue := enum.Get(varName)
		return fmt.Sprintf("%s /* %s.%s */", enumValue.Const.Val().ExactString(), typeName, varName)
	})

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
	for i, f := range ta.Columns {
		statement, colDecls := createStmt(f, ta.Primary() == i)

		colTypes[i] = "\t" + statement

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

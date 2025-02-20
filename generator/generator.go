package generator

import (
	"go/types"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/analysis/sql"
)

// Declaration is a top level declaration to write to the generated file.
type Declaration struct {
	ID       string // uniquely identifies the item, used to avoid duplicated declarations
	Content  string // actual code to write
	Priority bool   // if true, the declaration is written at the begining of the file
}

// WriteDeclarations remove duplicates and aggregate the declarations,
// sorting them by ID
func WriteDeclarations(decls []Declaration) string {
	sort.Slice(decls, func(i, j int) bool { return decls[i].ID < decls[j].ID })
	sort.SliceStable(decls, func(i, j int) bool { return decls[i].Priority && !decls[j].Priority })

	keys := map[string]bool{}
	var out strings.Builder
	for _, decl := range decls {
		if alreadyHere := keys[decl.ID]; !alreadyHere {
			keys[decl.ID] = true
			out.WriteString(decl.Content)
			out.WriteByte('\n')
		}
	}
	return out.String()
}

// Origin return the file where the type is defined,
// suitable to be included as comment.
func Origin(ty analysis.Type) string { return ty.Type().String() }

// NameRelativeTo is the same as types.Relative to, but
// use the (shorter) package name instead of path.
func NameRelativeTo(pkg *types.Package) types.Qualifier {
	if pkg == nil {
		return nil
	}
	return func(other *types.Package) string {
		if pkg == other {
			return "" // same package; unqualified
		}
		return other.Name()
	}
}

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func ToLowerFirst(str string) string {
	if str == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(str)
	return string(unicode.ToLower(r)) + str[n:]
}

// SQLTableName uses a familiar SQL convention for table names,
// shared by generator/sql and generator/go/sqlcrud
func SQLTableName(name sql.TableName) string {
	return ToSnakeCase(string(name)) + "s"
}

type TableNameReplacer map[string]string

func NewTableNameReplacer(tables []sql.Table) TableNameReplacer {
	out := make(TableNameReplacer)
	for _, ta := range tables {
		name := ta.TableName()
		out[string(name)] = SQLTableName(name)
	}
	return out
}

var reWords = regexp.MustCompile(`([\w]+)`)

// Replace replace words occurences
func (rp TableNameReplacer) Replace(input string) string {
	return reWords.ReplaceAllStringFunc(input, func(word string) string {
		if subs, has := rp[word]; has {
			return subs
		}
		return word
	})
}

// Cache is a cache used to handled recursive types.
type Cache map[*types.Named]bool

// Check returns `true` is `typ` is already in the cache,
// udpating it if not.
// Non named types are ignored.
func (c Cache) Check(typ analysis.Type) bool {
	if named, isNamed := typ.Type().(*types.Named); isNamed {
		if c[named] {
			return true
		}
		c[named] = true
	}
	return false
}

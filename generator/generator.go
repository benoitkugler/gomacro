package generator

import (
	"go/types"
	"sort"
	"strings"
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

// Cache is a cache used to handled recursive types.
type Cache map[*types.Named]bool

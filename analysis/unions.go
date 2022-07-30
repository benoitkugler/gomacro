package analysis

import (
	"go/types"
	"log"

	"golang.org/x/tools/go/packages"
)

// unionsMap is the set of named types which are
// considered as union types.
type unionsMap map[*types.Named][]*types.Named

// Union is deduced from interfaces, with the limitation
// that only types defined in the same package are considered
// as members of the union
type Union struct {
	name *types.Named // with underlying type *types.Interface

	// The types implementing this interface, sorted by name.
	// By construction, their Type() method will always return a *types.Named,
	// and Obj().Name() should be used as an identifier tag
	// accros the generators.
	Members []Type
}

func (u *Union) Type() types.Type { return u.name }

func allNamedTypes(pa *packages.Package) (out []*types.Named) {
	scope := pa.Types.Scope()
	for _, name := range scope.Names() {
		obj, ok := scope.Lookup(name).(*types.TypeName)
		if !ok {
			continue
		}

		if named, isNamed := obj.Type().(*types.Named); isNamed {
			out = append(out, named)
		}
	}
	return out
}

func fetchPkgUnions(pa *packages.Package) unionsMap {
	out := make(unionsMap)

	// get all the named types (candidates for members)
	candidates := allNamedTypes(pa)

	// walk through all the interface types
	for _, c := range candidates {
		itf, ok := c.Underlying().(*types.Interface)
		if !ok {
			continue
		}

		var members []*types.Named
		// walk again through the candidates
		for _, member := range candidates {
			// do not add interfaces as member of an union
			if _, isItf := member.Underlying().(*types.Interface); isItf {
				continue
			}

			if types.Implements(member, itf) {
				members = append(members, member)
			}
		}

		if len(members) == 0 {
			log.Printf("empty union type " + c.Obj().Pkg().Path() + "." + c.Obj().Name())
		}

		out[c] = members
	}
	return out
}

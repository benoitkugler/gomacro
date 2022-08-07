package httpapi

import (
	"go/types"

	"github.com/benoitkugler/gomacro/analysis"
)

// Endpoint describes one server endpoint
type Endpoint struct {
	Url      string
	Method   string
	Contract Contract
}

type TypedParam struct {
	type_ *types.Basic // used during parsing
	Type  *analysis.Basic
	Name  string
}

func (tp *TypedParam) resolveType(an *analysis.Analysis) {
	tp.Type = an.Types[tp.type_].(*analysis.Basic)
}

// Contract describes the expected and returned
// types for one endpoint.
type Contract struct {
	Name string

	// field used during parsing
	inputT  types.Type
	returnT types.Type

	InputBody analysis.Type

	// Return may be nil
	Return analysis.Type

	InputForm        Form
	InputQueryParams []TypedParam
}

type Form struct {
	File       string // empty means no file
	ValueNames []string
}

func (f Form) IsZero() bool {
	return f.File == "" && len(f.ValueNames) == 0
}

// AsTypedValues returns the name of the form parameters with type String
func (f Form) AsTypedValues() []TypedParam {
	out := make([]TypedParam, len(f.ValueNames))
	for i, v := range f.ValueNames {
		out[i] = TypedParam{Name: v, Type: analysis.String}
	}
	return out
}

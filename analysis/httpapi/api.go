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
	type_ types.Type // used during parsing
	Type  analysis.Type
	Name  string
}

// Contract describes the expected and returned
// types for one endpoint.
type Contract struct {
	// field used during parsing
	inputT  types.Type
	returnT types.Type

	Input, Return analysis.Type

	Form        Form
	Name        string
	QueryParams []TypedParam
}

type Form struct {
	File   string // empty means no file
	Values []string
}

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
	Comment  SpecialComment
}

type TypedParam struct {
	type_ types.Type // used during parsing
	Type  analysis.Type
	Name  string
}

func (tp *TypedParam) resolveType(an *analysis.Analysis) {
	tp.Type = an.Types[tp.type_]
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

	IsReturnBlob bool // [Return] is a []byte, interpreted as a file

	IsReturnStream bool // JSON stream of type [Return]
}

type Form struct {
	File       string // empty means no file
	ValueNames []string
	JSON       TypedParam
}

func (f Form) IsZero() bool {
	return f.File == "" && len(f.ValueNames) == 0 && f.JSON.Name == ""
}

// AsTypedValues returns the name of the form parameters with type String
func (f Form) AsTypedValues() []TypedParam {
	out := make([]TypedParam, len(f.ValueNames))
	for i, v := range f.ValueNames {
		out[i] = TypedParam{Name: v, Type: analysis.String}
	}
	return out
}

type SpecialComment uint8

const (
	_ SpecialComment = iota
	Ignore
)

func newSpecialComment(comment string) SpecialComment {
	switch comment {
	case "":
		return 0
	case "ignore":
		return Ignore
	default:
		panic("invalid special comment " + comment)
	}
}

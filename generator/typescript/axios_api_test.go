package typescript

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/analysis/httpapi"
	"github.com/benoitkugler/gomacro/generator"
)

func TestGenerate1(t *testing.T) {
	apis := []httpapi.Endpoint{
		{
			Url: "/samlskm/", Method: http.MethodPost, Contract: httpapi.Contract{
				Name:      "M1",
				InputBody: &analysis.Array{Elem: analysis.Bool, Len: 5},
				Return:    &analysis.Array{Elem: analysis.Int, Len: -1},
			},
		},
		{
			Url: "/samlskm/:param1", Method: http.MethodGet, Contract: httpapi.Contract{
				Name: "M2",
				InputQueryParams: []httpapi.TypedParam{
					{Name: "arg1", Type: analysis.String},
					{Name: "arg2", Type: analysis.Int},
					{Name: "arg2bis", Type: analysis.Float},
					{Name: "arg3", Type: analysis.Bool},
				},
				Return: &analysis.Array{Elem: analysis.Int, Len: -1},
			},
		},
	}
	fmt.Println(GenerateAxios(apis))
}

func TestGenerateMaps(t *testing.T) {
	api := httpapi.Endpoint{
		Url:    "/samlskm/",
		Method: http.MethodPost,
		Contract: httpapi.Contract{
			Name:      "M1",
			InputBody: &analysis.Array{Elem: analysis.Bool, Len: 5},
			Return:    &analysis.Map{Key: analysis.String, Elem: analysis.Int},
		},
	}
	fmt.Println(generateMethod(api))
}

func TestGenerateTypes(t *testing.T) {
	an, err := analysis.NewAnalysisFromFile("../../testutils/testsource/defs.go")
	if err != nil {
		t.Fatal(err)
	}

	var (
		decls []generator.Declaration
		cache = make(generator.Cache)
	)
	for _, ty := range an.Source {
		decls = append(decls, generate(an.Types[ty], cache)...)
	}
	out := generator.WriteDeclarations(decls)

	fn := "test/gen.ts"
	err = os.WriteFile(fn, []byte(out), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	var fmts generator.Formatters
	if err := fmts.FormatFile(generator.TypeScript, fn); err != nil {
		t.Fatal(err)
	}
}

package typescript

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/analysis/httpapi"
	"github.com/benoitkugler/gomacro/generator"
)

func Test(t *testing.T) {
	fn := "../../analysis/httpapi/test/routes.go"
	pack, err := analysis.LoadSource(fn)
	if err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs(fn)
	if err != nil {
		t.Fatal(err)
	}

	apis := httpapi.ParseEcho(pack, abs, "")
	code := GenerateURLs(apis)
	err = os.WriteFile("test/urls_gen.ts", []byte(code), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	var fmts generator.Formatters
	if err := fmts.FormatFile(generator.TypeScript, "test/urls_gen.ts"); err != nil {
		t.Fatal(err)
	}
}

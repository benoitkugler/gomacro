package httpapi

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/benoitkugler/gomacro/analysis"
)

func TestParse(t *testing.T) {
	fn := "test/routes.go"
	pack, err := analysis.LoadSource(fn)
	if err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs(fn)
	if err != nil {
		t.Fatal(err)
	}

	ti := time.Now()
	apis := ParseEcho(pack, abs)
	fmt.Println("Resolved in ", time.Since(ti))
	if len(apis) != 19 {
		t.Fatalf("expected 19 endpoints, got %d", len(apis))
	}

	if apis[9].Contract.InputForm.JSON.type_ == nil {
		t.Fatal()
	}
	if apis[9].Contract.InputForm.JSON.Type == nil {
		t.Fatal()
	}
	if len(apis[9].Contract.InputForm.ValueNames) != 1 {
		t.Fatal()
	}
	if apis[9].Contract.InputForm.ValueNames[0].Type == nil {
		t.Fatal()
	}

	if len(apis[15].Contract.InputQueryParams) == 0 {
		t.Fatal("generic not supported")
	}

	if len(apis[16].Contract.InputQueryParams) != 1 {
		t.Fatal()
	}

	if !apis[17].Contract.IsReturnStream {
		t.Fatal()
	}
	if !apis[18].IsUrlOnly {
		t.Fatal()
	}
	if apis[11].Contract.InputForm.Files != "documents" {
		t.Fatalf("expected multiple-file field to be parsed, got %q", apis[11].Contract.InputForm.Files)
	}
}

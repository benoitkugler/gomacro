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
	if len(apis) != 14 {
		t.Fatal()
	}
}

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

	file := selectFileByPath(pack, abs)
	ti := time.Now()
	apis := Parse(pack, file)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(time.Since(ti))
	fmt.Println(apis)
}

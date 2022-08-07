package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/analysis/httpapi"
	"github.com/benoitkugler/gomacro/generator"
	"github.com/benoitkugler/gomacro/generator/dart"
	"github.com/benoitkugler/gomacro/generator/go/gounions"
	"github.com/benoitkugler/gomacro/generator/go/randdata"
	"github.com/benoitkugler/gomacro/generator/go/sqlcrud"
	"github.com/benoitkugler/gomacro/generator/sql"
	"github.com/benoitkugler/gomacro/generator/typescript"
	"golang.org/x/tools/go/packages"
)

// command line based API

type mode string

const (
	goUnionsGen   = "go/unions"
	goSqlcrudGen  = "go/sqlcrud"
	goRanddataGen = "go/randdata"
	sqlGen        = "sql"
	typescriptGen = "typescript"
	dartGen       = "dart"
)

var fmts generator.Formatters

type action struct {
	Mode   mode
	Output string
}

type Actions []action

func (i *Actions) String() string {
	if i == nil {
		return ""
	}
	return fmt.Sprint(*i)
}

func (i *Actions) Set(value string) error {
	chuncks := strings.Split(value, ":")
	if len(chuncks) != 2 {
		return fmt.Errorf("expected colon separated <mode>:<output>, got %s", value)
	}
	m := action{Mode: mode(chuncks[0]), Output: chuncks[1]}
	switch m.Mode {
	case goUnionsGen, goSqlcrudGen, goRanddataGen,
		sqlGen, typescriptGen, dartGen:
	default:
		return fmt.Errorf("invalid mode %s", m.Mode)
	}
	if m.Output == "" {
		return fmt.Errorf("output not specified for mode %s", m.Mode)
	}
	*i = append(*i, m)
	return nil
}

func runActions(source string, pkg *packages.Package, actions Actions) error {
	fullPath, err := filepath.Abs(source)
	if err != nil {
		return err
	}

	log.Printf("running actions for %s", fullPath)

	ana := analysis.NewAnalysisFromFile(pkg, source)

	for _, m := range actions {
		var (
			code   string
			format generator.Format // format if true
		)
		switch m.Mode {
		case goUnionsGen:
			code = generator.WriteDeclarations(gounions.Generate(ana))
			format = generator.Go
		case goSqlcrudGen:
			code = generator.WriteDeclarations(sqlcrud.Generate(ana))
			format = generator.Go
		case goRanddataGen:
			code = generator.WriteDeclarations(randdata.Generate(ana))
			format = generator.Go
		case sqlGen:
			code = generator.WriteDeclarations(sql.Generate(ana))
			format = generator.Psql
		case typescriptGen:
			api := httpapi.ParseEcho(ana.Root, fullPath)
			code = typescript.GenerateAxios(api)
			format = generator.TypeScript
		case dartGen:
			code = generator.WriteDeclarations(dart.Generate(ana))
			format = generator.Dart
		default:
			panic(m.Mode)
		}

		err := os.WriteFile(m.Output, []byte(code), os.ModePerm)
		if err != nil {
			return err
		}

		err = fmts.FormatFile(format, m.Output)
		if err != nil {
			return fmt.Errorf("formatting %s failed: generated code is probably incorrect: %s", m.Output, err)
		}

		log.Printf("Code for action %s written in %s", m.Mode, m.Output)
	}
	return nil
}

func runFromArgs(source string, actions Actions) error {
	pkg, err := analysis.LoadSource(source)
	if err != nil {
		return err
	}
	err = runActions(source, pkg, actions)
	return err
}

// configuration file based API

// Config maps a list of files to the actions to apply
type Config map[string]Actions

func runFromConfig(configFile string) error {
	f, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer f.Close()

	var conf Config
	if err = json.NewDecoder(f).Decode(&conf); err != nil {
		return err
	}

	// fetch the packages for each file in one call
	var files []string
	for file := range conf {
		files = append(files, file)
	}
	sort.Strings(files) // ensure deterministic execution order

	pkgs, err := analysis.LoadSources(files)
	if err != nil {
		return err
	}

	log.Println("Source loading done.")

	for i, file := range files {
		actions := conf[file]
		pkg := pkgs[i]
		err = runActions(file, pkg, actions)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	configFilePtr := flag.String("conf", "", "JSON config file defining which actions to execute")
	sourcePtr := flag.String("source", "", "go source file to convert")
	var actions Actions
	flag.Var(&actions, "actions", "list of actions <mode>:<output>")

	flag.Parse()

	source, configFile := *sourcePtr, *configFilePtr
	if source == "" && configFile == "" {
		log.Fatal("Please define an input source file or a configuration file.")
	} else if source != "" && configFile != "" {
		log.Fatal("Please define either an input source file or a configuration file, not both.")
	}

	if configFile != "" {
		if err := runFromConfig(configFile); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := runFromArgs(source, actions); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("Done.")
}

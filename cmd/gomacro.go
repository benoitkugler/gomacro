package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

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
	goUnionsGen        = "go/unions"
	goSqlcrudGen       = "go/sqlcrud"
	goRanddataGen      = "go/randdata"
	sqlGen             = "sql"
	typescriptApiGen   = "typescript/api"
	typescriptTypesGen = "typescript/types"
	dartGen            = "dart"
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
		sqlGen, typescriptApiGen, dartGen:
	default:
		return fmt.Errorf("invalid mode %s", m.Mode)
	}
	if m.Output == "" {
		return fmt.Errorf("output not specified for mode %s", m.Mode)
	}
	*i = append(*i, m)
	return nil
}

type outputFile struct {
	format  generator.Format
	file    string
	content string
}

// special case for dart actions, which are returned for latter processsing
func runActions(source string, pkg *packages.Package, actions Actions) (*analysis.Analysis, []outputFile, error) {
	fullPath, err := filepath.Abs(source)
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("Running actions for %s\n", source)

	ana := analysis.NewAnalysisFromFile(pkg, source)

	hasDart := false
	var outs []outputFile
	for _, m := range actions {
		var (
			code   string
			format generator.Format // format if true
			output = m.Output
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
		case typescriptTypesGen:
			code = generator.WriteDeclarations(typescript.Generate(ana))
			format = generator.TypeScript
		case typescriptApiGen:
			api := httpapi.ParseEcho(ana.Root, fullPath)
			code = typescript.GenerateAxios(api)
			format = generator.TypeScript
		case dartGen:
			hasDart = true
			// code = generator.WriteDeclarations(dart.Generate(ana))
			// format = generator.Dart
		default:
			panic(m.Mode)
		}

		if code != "" {
			outs = append(outs, outputFile{format: format, file: output, content: code})
		}

		// err := os.WriteFile(output, []byte(code), os.ModePerm)
		// if err != nil {
		// 	return nil, err
		// }

		// fmt.Printf("\tCode for action <%s> written to %s (pending formatting).\n", m.Mode, output)

		// go func() {
		// 	err = fmts.FormatFile(format, output)
		// 	if err != nil {
		// 		panic(fmt.Sprintf("formatting %s failed: generated code is probably incorrect: %s", output, err))
		// 	}
		// 	wg.Done()
		// }()

	}
	if hasDart {
		return ana, outs, nil
	}
	return nil, outs, nil
}

func saveOutputs(commonDir, dartOutputDir string, dartAnalysis []*analysis.Analysis, outputs []outputFile) error {
	for _, out := range dart.Generate(commonDir, dartAnalysis) {
		outputs = append(outputs, outputFile{
			format:  generator.Dart,
			file:    filepath.Join(dartOutputDir, out.Filename),
			content: generator.WriteDeclarations(out.Content),
		})
	}

	fmt.Println("Code generated. Saving and formatting...")

	var wg sync.WaitGroup
	wg.Add(len(outputs))
	for _, out := range outputs {
		output := out.file
		format := out.format
		err := os.WriteFile(output, []byte(out.content), os.ModePerm)
		if err != nil {
			return err
		}

		fmt.Printf("\tCode written to %s (pending formatting).\n", output)

		go func() {
			err = fmts.FormatFile(format, output)
			if err != nil {
				panic(fmt.Sprintf("formatting %s failed: generated code is probably incorrect: %s", output, err))
			}
			wg.Done()
		}()
	}
	fmt.Println("Waiting for formatters...")
	wg.Wait()
	return nil
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

	dartOutputDir := conf["_dart"][0].Output
	delete(conf, "_dart")

	// fetch the packages for each file in one call
	var files []string
	for file := range conf {
		files = append(files, file)
	}
	sort.Strings(files) // ensure deterministic execution order

	fmt.Println("Type-checking source files...")
	pkgs, commonDir, err := analysis.LoadSources(files)
	if err != nil {
		return err
	}
	fmt.Println("Source loading done. Root directory:", commonDir)

	var (
		allOutputs []outputFile
		dartAnas   []*analysis.Analysis
	)
	for i, file := range files {
		actions := conf[file]
		pkg := pkgs[i]
		dartAna, outs, err := runActions(file, pkg, actions)
		if err != nil {
			return err
		}
		allOutputs = append(allOutputs, outs...)
		if dartAna != nil {
			dartAnas = append(dartAnas, dartAna)
		}
	}

	return saveOutputs(commonDir, dartOutputDir, dartAnas, allOutputs)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide a configuration file.")
	}

	configFile := os.Args[1]

	if err := runFromConfig(configFile); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Done.")
}

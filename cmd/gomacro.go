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
	Mode            mode
	Output          string
	RestrictHTTPApi string // only for [typescriptApiGen]
	UrlOnly         bool   // only for [typescriptApiGen]
}

func newAction(value string) (action, error) {
	md, output, ok := strings.Cut(value, ":")
	if !ok {
		return action{}, fmt.Errorf("expected colon separated <mode>:<output>, got %s", value)
	}
	m := action{Mode: mode(md), Output: output}
	switch m.Mode {
	case goUnionsGen, goSqlcrudGen, goRanddataGen,
		sqlGen, typescriptApiGen, typescriptTypesGen, dartGen:
	default:
		const usage = `
		Supported modes : 
		"go/unions","go/sqlcrud","go/randdata","sql","typescript/api","typescript/types","dart"
	`
		return action{}, fmt.Errorf("invalid mode %s %s", m.Mode, usage)
	}
	if m.Output == "" {
		return action{}, fmt.Errorf("output not specified for mode %s", m.Mode)
	}

	return m, nil
}

type Actions []action

func (acs Actions) hasDart() bool {
	for _, ac := range acs {
		if ac.Mode == dartGen {
			return true
		}
	}
	return false
}

func (i *Actions) String() string {
	if i == nil {
		return ""
	}
	return fmt.Sprint(*i)
}

func (i *Actions) Set(value string) error {
	m, err := newAction(value)
	if err != nil {
		return err
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
func runActions(source string, pkg *packages.Package, actions Actions, dartOnly, generateSets bool) (*analysis.Analysis, []outputFile, error) {
	if dartOnly && !actions.hasDart() {
		return nil, nil, nil
	}

	fullPath, err := filepath.Abs(source)
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("Running actions for %s...\n", source)

	ana := analysis.NewAnalysisFromFile(pkg, source)

	fmt.Println("Code analysis completed. Generating..")

	hasDart := false
	var outs []outputFile
	for _, act := range actions {
		var (
			code   string
			format generator.Format // format if true
			output = act.Output
		)

		if dartOnly && act.Mode != dartGen {
			continue
		}

		switch act.Mode {
		case goUnionsGen:
			code = generator.WriteDeclarations(gounions.Generate(ana))
			format = generator.Go
		case goSqlcrudGen:
			code = generator.WriteDeclarations(sqlcrud.Generate(ana, generateSets))
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
			fmt.Println("Parsing http routes...")
			api := httpapi.ParseEcho(ana.Pkg, fullPath, act.RestrictHTTPApi)
			fmt.Println("Done. Generating", len(api), "routes")
			if act.UrlOnly {
				code = typescript.GenerateURLs(api)
			} else {
				code = typescript.GenerateAxios(api)
			}
			format = generator.TypeScript
		case dartGen:
			hasDart = true
			// code = generator.WriteDeclarations(dart.Generate(ana))
			// format = generator.Dart
		default:
			panic(act.Mode)
		}

		if code != "" {
			outs = append(outs, outputFile{format: format, file: output, content: code})
		}
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

func newConfigFromJSON(configFile string) (Config, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var conf Config
	if err = json.NewDecoder(f).Decode(&conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func (conf Config) run(dartOnly, generateSets bool) error {
	dartOutputDir := ""
	if dart, ok := conf["_dart"]; ok {
		dartOutputDir = dart[0].Output
		delete(conf, "_dart")
	}

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
		dartAna, outs, err := runActions(file, pkg, actions, dartOnly, generateSets)
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
	isConfig := flag.Bool("config", false, "Use a config file")
	isDartOnly := flag.Bool("dart-only", false, "Only run Dart actions")
	restrictHTTPApi := flag.String("http-api", "", "Only generate API for endpoints with this prefix")
	generateSetsID := flag.Bool("generate-sets", false, "Generate a convenient Set type")
	httpURLOnly := flag.Bool("url-only", false, "Generates URL instead of Axios calls")
	flag.Parse()

	fileArgs := flag.Args()
	if len(fileArgs) < 1 {
		log.Fatal("Please provide a configuration file or an input file.")
	}

	var (
		conf Config
		err  error
	)
	if *isConfig { // config mode
		configFile := fileArgs[0]
		conf, err = newConfigFromJSON(configFile)
		if err != nil {
			log.Fatal(err)
		}
	} else { // single file mode
		inputFile := fileArgs[0]
		conf = make(Config)
		for _, actionString := range fileArgs[1:] {
			action, err := newAction(actionString)
			if err != nil {
				log.Fatal(err)
			}
			action.RestrictHTTPApi = *restrictHTTPApi
			action.UrlOnly = *httpURLOnly
			conf[inputFile] = append(conf[inputFile], action)
		}
	}

	if err := conf.run(*isDartOnly, *generateSetsID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Done.")
}

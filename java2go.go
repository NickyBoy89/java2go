package main

import (
	"context"
	"flag"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"

	"github.com/NickyBoy89/java2go/symbol"
	log "github.com/sirupsen/logrus"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
)

var (
	// Stores a global list of Java annotations to exclude from the generated code
	excludedAnnotations = make(map[string]bool)
)

var (
	writeFiles              bool
	quiet                   bool
	displayAST              bool
	parseFilesSynchronously bool
	outputDirectory         string
	ignoredAnnotations      string

	cpuProfile string
)

type SourceFile struct {
	Name    string
	Source  []byte
	Ast     *sitter.Node
	Symbols *symbol.FileScope
}

func main() {
	flag.BoolVar(&writeFiles, "w", false, "Whether to write the files to disk instead of stdout")
	flag.BoolVar(&quiet, "q", false, "Don't write to stdout on successful parse")
	flag.BoolVar(&displayAST, "ast", false, "Print out go's pretty-printed ast, instead of source code")
	flag.BoolVar(&parseFilesSynchronously, "sync", false, "Parse the files one by one, instead of in parallel")
	flag.StringVar(&outputDirectory, "outDir", ".", "Specify a directory for the generated files")
	flag.StringVar(&ignoredAnnotations, "exclude-annotations", "", "A comma-separated list of annotations to exclude from the final code generation")

	flag.StringVar(&cpuProfile, "cpuprofile", "", "write cpu profile to `file`")

	flag.Parse()

	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	for _, annotation := range strings.Split(ignoredAnnotations, ",") {
		excludedAnnotations[annotation] = true
	}

	// All the files to parse
	files := []SourceFile{}

	log.Info("Collecting files...")

	// Collect all the files and read them into memory
	for _, file := range flag.Args() {
		err := filepath.WalkDir(file, fs.WalkDirFunc(
			func(path string, d fs.DirEntry, err error) error {
				// Only include java files
				if filepath.Ext(path) == ".java" && !d.IsDir() {
					sourceCode, err := os.ReadFile(path)
					if err != nil {
						log.WithFields(log.Fields{
							"file":  path,
							"error": err,
						}).Panic("Error reading source file")
					}

					files = append(files, SourceFile{
						Name:   path,
						Source: sourceCode,
					})
				}

				return nil
			},
		))

		if err != nil {
			log.WithFields(log.Fields{
				"file":  file,
				"error": err,
			}).Fatal("Error walking directory or file")
		}
	}

	if len(files) == 0 {
		log.Warn("No files specified to convert")
		return
	}

	// Parse the ASTs of all the files

	log.Info("Parsing ASTs...")

	sem := make(chan struct{}, runtime.NumCPU())

	var wg sync.WaitGroup
	wg.Add(len(files))

	for index := range files {
		sem <- struct{}{}

		go func(index int) {
			parser := sitter.NewParser()
			parser.SetLanguage(java.GetLanguage())
			tree, err := parser.ParseCtx(context.Background(), nil, files[index].Source)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Panic("Error parsing tree-sitter AST")
			}

			parser.Close()

			files[index].Ast = tree.RootNode()

			<-sem
			wg.Done()
		}(index)
	}

	// We might still have some parsing jobs, so wait on them
	wg.Wait()

	// Generate the symbol tables for the files

	log.Info("Generating symbol tables...")

	for index, file := range files {
		if file.Ast.HasError() {
			log.WithFields(log.Fields{
				"fileName": file.Name,
			}).Warn("AST parse error in file, skipping file")
			continue
		}

		symbols := symbol.ParseSymbols(file.Ast, file.Source)

		files[index].Symbols = symbols

		symbol.GlobalScope.Packages[symbols.Package].AddSymbolsFromFile(symbols)
	}

	// Go back through the symbol tables and fill in anything that could not be resolved

	log.Info("Resolving symbols...")

	for _, file := range files {

		// Resolve all the fields in that respective class
		for _, field := range file.Symbols.BaseClass.Fields {

			// Since a private global variable is able to be accessed in the package, it must be renamed
			// to avoid conflicts with other global variables

			packageScope := symbol.GlobalScope.FindPackage(file.Symbols.Package)

			symbol.ResolveDefinition(field, file.Symbols, symbol.GlobalScope)

			// Rename the field if its name conflits with any keyword
			for i := 0; symbol.IsReserved(field.Name) || len(packageScope.FindStaticField().ByName(field.Name)) > 0; i++ {
				field.Rename(field.Name + strconv.Itoa(i))
			}
		}
		for _, method := range file.Symbols.BaseClass.Methods {
			// Resolve the return type, as well as the body of the method
			symbol.ResolveChildren(method, file.Symbols, symbol.GlobalScope)

			for i := 0; symbol.IsReserved(method.Name); /* || method.MethodExistsIn(file.Symbols.BaseClass)*/ i++ {
				method.Rename(method.Name + strconv.Itoa(i))
			}
			// Resolve all the paramters of the method
			for _, param := range method.Parameters {
				symbol.ResolveDefinition(param, file.Symbols, symbol.GlobalScope)

				for i := 0; symbol.IsReserved(param.Name); i++ {
					param.Rename(param.Name + strconv.Itoa(i))
				}
			}
		}
	}

	// Transpile the files

	log.Info("Converting files...")

	for _, file := range files {
		log.Infof("Converting file \"%s\"", file.Name)

		// Write to stdout by default
		var output io.Writer = os.Stdout

		// Write to a `.go` file in the same directory
		outputFile := file.Name[:len(file.Name)-len(filepath.Ext(file.Name))] + ".go"
		outputPath := outputDirectory + "/" + outputFile

		if writeFiles {
			err := os.MkdirAll(path.Dir(outputPath), 0755)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"path":  outputPath,
				}).Panic("Error creating output directory")
			}

			// Write the output to a file
			output, err = os.Create(outputPath)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"file":  outputPath,
				}).Panic("Error creating output file")
			}
		} else if quiet {
			// Otherwise, throw away the output
			output = io.Discard
		}

		// The converted AST, in Go's AST representation
		parsed := ParseNode(file.Ast, file.Source, Ctx{classScope: file.Symbols.BaseClass}).(ast.Node)

		// Print the generated AST
		if displayAST {
			ast.Print(token.NewFileSet(), parsed)
		}

		// Output the parsed AST, into the source specified earlier
		if err := printer.Fprint(output, token.NewFileSet(), parsed); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Panic("Error printing generated code")
		}

		if writeFiles {
			output.(*os.File).Close()
		}
	}
}

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/NickyBoy89/java2go/parsing"
	"github.com/NickyBoy89/java2go/symbol"
	log "github.com/sirupsen/logrus"
)

// Stores a global list of Java annotations to exclude from the generated code
var excludedAnnotations = make(map[string]bool)

// Command-line arguments
var (
	writeFiles              bool
	dryRun                  bool
	displayAST              bool
	symbolAware             bool
	parseFilesSynchronously bool
)

var (
	outputDirectory    string
	ignoredAnnotations string
)

func main() {
	flag.BoolVar(&writeFiles, "w", false, "Whether to write the files to disk instead of stdout")
	flag.BoolVar(&dryRun, "q", false, "Don't write to stdout on successful parse")
	flag.BoolVar(&displayAST, "ast", false, "Print out go's pretty-printed ast, instead of source code")
	flag.BoolVar(&parseFilesSynchronously, "sync", false, "Parse the files one by one, instead of in parallel")
	flag.BoolVar(&symbolAware, "symbols", true, `Whether the program is aware of the symbols of the parsed code
Results in better code generation, but can be disabled for a more direct translation
or to fix crashes with the symbol handling`,
	)
	flag.StringVar(&outputDirectory, "output", ".", "Specify a directory for the generated files")
	flag.StringVar(&ignoredAnnotations, "exclude-annotations", "", "A comma-separated list of annotations to exclude from the final code generation")

	flag.Parse()

	for _, annotation := range strings.Split(ignoredAnnotations, ",") {
		excludedAnnotations[annotation] = true
	}

	// All the files to parse
	var files []parsing.SourceFile

	log.Info("Collecting files...")

	// Collect all the files and read them into memory
	for _, dirName := range flag.Args() {
		sources, err := parsing.ReadSourcesInDir(dirName)
		if err != nil {
			log.WithField("error", err).Fatal("Error reading directory")
		}
		files = append(files, sources...)
	}

	if len(files) == 0 {
		log.Warn("No files specified to convert")
	}

	// Parse the ASTs of all the files

	log.Info("Parsing ASTs...")

	var wg sync.WaitGroup
	wg.Add(len(files))

	for index := range files {
		parseFunc := func(ind int) {
			if err := files[ind].ParseAST(); err != nil {
				log.WithField("error", err).Error("Error parsing AST")
			}
			wg.Done()
		}

		if parseFilesSynchronously {
			parseFunc(index)
		} else {
			go parseFunc(index)
		}
	}

	// We might still have some parsing jobs, so wait on them
	wg.Wait()

	for _, file := range files {
		if file.Ast == nil {
			panic("Not all files have asts")
		}
	}

	// Generate the symbol tables for the files
	if symbolAware {
		log.Info("Generating symbol tables...")

		for index, file := range files {
			if file.Ast.HasError() {
				log.WithFields(log.Fields{
					"fileName": file.Name,
				}).Warn("AST parse error in file, skipping file")
				continue
			}

			symbols := files[index].ParseSymbols()
			// Add the symbols to the global symbol table
			symbol.AddSymbolsToPackage(symbols)
		}

		// Go back through the symbol tables and fill in anything that could not be resolved

		log.Info("Resolving symbols...")

		for _, file := range files {
			if !file.Ast.HasError() {
				ResolveFile(file)
			}
		}
	}

	// Transpile the files

	log.Info("Converting files...")

	for _, file := range files {
		if dryRun {
			log.Infof("Not converting file \"%s\"", file.Name)
			continue
		}

		log.Infof("Converting file \"%s\"", file.Name)

		// Write to stdout by default
		var output io.Writer = os.Stdout
		if writeFiles {
			// Write to a `.go` file in the same directory
			outputFile := fmt.Sprintf("%s/%s",
				outputDirectory,
				strings.TrimSuffix(file.Name, filepath.Ext(file.Name))+".go",
			)

			err := os.MkdirAll(outputDirectory, 0755)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"path":  outputFile,
				}).Panic("Error creating output directory")
			}

			// Write the output to a file
			output, err = os.Create(outputFile)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"file":  outputFile,
				}).Panic("Error creating output file")
			}
		}

		// The converted AST, in Go's AST representation
		var initialContext Ctx
		if symbolAware {
			initialContext.currentFile = file.Symbols
			initialContext.currentClass = file.Symbols.BaseClass
		}

		parsed := ParseNode(file.Ast, file.Source, initialContext).(ast.Node)

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

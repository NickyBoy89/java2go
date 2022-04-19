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
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
)

var (
	// Stores a global list of Java annotations to exclude from the generated code
	excludedAnnotations = make(map[string]struct{})
)

func main() {
	writeFlag := flag.Bool("w", false, "Whether to write the files to disk instead of stdout")
	quiet := flag.Bool("q", false, "Don't write to stdout on successful parse")
	astFlag := flag.Bool("ast", false, "Print out go's pretty-printed ast, instead of source code")
	syncFlag := flag.Bool("sync", false, "Parse the files sequentially, instead of multi-threaded")
	outDirFlag := flag.String("outDir", ".", "Specify a directory for the generated files")

	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

	excludeAnnotationsFlag := flag.String("exclude-annotations", "", "A comma-separated list of annotations to exclude from the final code generation")

	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	for _, annotation := range strings.Split(*excludeAnnotationsFlag, ",") {
		excludedAnnotations[annotation] = struct{}{}
	}

	// All the files to parse
	fileNames := []string{}

	for _, file := range flag.Args() {
		err := filepath.WalkDir(file, fs.WalkDirFunc(func(path string, d fs.DirEntry, err error) error {
			// Only include java files
			if filepath.Ext(path) == ".java" && !d.IsDir() {
				fileNames = append(fileNames, path)
			}

			return nil
		}))

		if err != nil {
			log.WithFields(log.Fields{
				"file":  file,
				"error": err,
			}).Fatal("Error walking directory or file")
		}
	}

	if len(fileNames) == 0 {
		log.Warn("No files specified to convert")
	}

	var wg sync.WaitGroup
	wg.Add(len(fileNames))

	// Sem determines the number of files parsed in parallel
	sem := make(chan struct{}, runtime.NumCPU())

	parsedAsts := make(chan struct {
		ast   *sitter.Node
		index int
	})

	asts := make([]*sitter.Node, len(fileNames))

	// Parse all the files into their tree-sitter representations
	for ind, filePath := range fileNames {
		go func(index int, path string) {
			sourceCode, err := os.ReadFile(path)
			if err != nil {
				log.WithFields(log.Fields{
					"file":  path,
					"error": err,
				}).Panic("Error reading source file")
			}

			sem <- struct{}{}
			// Release the semaphore when done
			defer func() { <-sem }()
			defer wg.Done()
			parser := sitter.NewParser()
			defer parser.Close()
			parser.SetLanguage(java.GetLanguage())
			tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Panic("Error parsing tree-sitter AST")
			}

			var test struct {
				ast   *sitter.Node
				index int
			}

			test.ast = tree.RootNode()
			test.index = index

			parsedAsts <- test
		}(ind, filePath)
	}

	var n int
	for p := range parsedAsts {
		asts[p.index] = p.ast
		n++
		if n == len(fileNames) {
			close(parsedAsts)
		}
	}

	globalPackages := make(map[string]*PackageScope)
	classDefinitions := make([]*ClassScope, len(fileNames))

	// Generate symbol tables
	for ind, filePath := range fileNames {
		sourceCode, err := os.ReadFile(filePath)
		if err != nil {
			log.WithFields(log.Fields{
				"file":  filePath,
				"error": err,
			}).Panic("Error reading source file")
		}

		classDef := ExtractDefinitions(asts[ind], sourceCode)
		classDefinitions[ind] = classDef
		classPackage := classDef.Package
		if classPackage == "" {
			classPackage = "main"
		}
		if _, exist := globalPackages[classDef.Package]; !exist {
			globalPackages[classDef.Package] = &PackageScope{files: make(map[string]*ClassScope)}
		}
		globalPackages[classDef.Package].files[classPackage] = classDef
	}

	_ = syncFlag

	// Start looking through the files
	for ind, filePath := range fileNames {
		sourceCode, err := os.ReadFile(filePath)
		if err != nil {
			log.WithFields(log.Fields{
				"file":  filePath,
				"error": err,
			}).Panic("Error reading source file")
		}

		log.Infof("Converting file \"%s\"", filePath)

		// Write to stdout by default
		var output io.Writer = os.Stdout

		outputFile := filePath[:len(filePath)-len(filepath.Ext(filePath))] + ".go"
		outputPath := *outDirFlag + "/" + outputFile

		if *writeFlag {
			if err := os.MkdirAll(path.Dir(outputPath), 0755); err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"path":  outputPath,
				}).Panic("Error creating output directory")
			}

			// Write the output to another file
			output, err := os.Create(outputPath)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"file":  outputPath,
				}).Panic("Error creating output file")
			}
			defer output.Close()
		} else if *quiet {
			output = io.Discard
		}

		// The converted AST, in Go's AST representation
		parsed := ParseNode(asts[ind], sourceCode, Ctx{classScope: classDefinitions[ind]}).(ast.Node)

		// Print the generated AST
		if *astFlag {
			ast.Print(token.NewFileSet(), parsed)
		}

		// Output the parsed AST, into the source specified earlier
		if err := printer.Fprint(output, token.NewFileSet(), parsed); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Panic("Error printing generated code")
		}
	}
}

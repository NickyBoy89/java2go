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
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"

	stdpath "path"

	log "github.com/sirupsen/logrus"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
)

var (
	// Stores a global list of Java annotations to exclude from the generated code
	excludedAnnotations = make(map[string]struct{})
)

func main() {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())

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

	// Sem determines the number of files parsed in parallel
	sem := make(chan struct{}, runtime.NumCPU())

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
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(fileNames))

	// Start looking through the files
	for _, path := range fileNames {
		sourceCode, err := os.ReadFile(path)
		if err != nil {
			log.WithFields(log.Fields{
				"file":  path,
				"error": err,
			}).Panic("Error reading source file")
		}

		tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Panic("Error parsing tree-sitter AST")
		}

		n := tree.RootNode()

		log.Infof("Converting file \"%s\"", path)

		// Write to stdout by default
		var output io.Writer = os.Stdout

		outputFile := path[:len(path)-len(filepath.Ext(path))] + ".go"
		outputPath := *outDirFlag + "/" + outputFile

		if *writeFlag {
			if err := os.MkdirAll(stdpath.Dir(outputPath), 0755); err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"path":  outputPath,
				}).Panic("Error creating output directory")
			}

			// Write the output to another file
			output, err = os.Create(outputPath)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"file":  outputPath,
				}).Panic("Error creating output file")
			}
			defer output.(*os.File).Close()
		} else if *quiet {
			output = io.Discard
		}

		// Acquire a semaphore
		sem <- struct{}{}

		parseFunc := func() {
			// Release the semaphore when done
			defer func() { <-sem }()

			defer wg.Done()

			parsedAst := ParseNode(n, sourceCode, Ctx{}).(ast.Node)

			// Print the generated AST
			if *astFlag {
				ast.Print(token.NewFileSet(), parsedAst)
			}

			if err := printer.Fprint(output, token.NewFileSet(), parsedAst); err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Panic("Error printing generated code")
			}
		}

		// If we don't want this to run in parallel
		if *syncFlag {
			parseFunc()
		} else {
			go parseFunc()
		}
	}

	wg.Wait()
}

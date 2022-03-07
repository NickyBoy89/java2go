package main

import (
	"context"
	"flag"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	log "github.com/sirupsen/logrus"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
	"gitlab.nicholasnovak.io/snapdragon/java2go/dot"
)

func main() {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())

	writeFlag := flag.Bool("w", false, "Whether to write the files to disk instead of stdout")
	quiet := flag.Bool("q", false, "Don't write to stdout on successful parse")
	astFlag := flag.Bool("ast", false, "Print out go's pretty-printed ast, instead of source code")
	dependencyTreeFlag := flag.Bool("dependency-tree", false, "Output a dependency tree of all classes, in graphviz dot format")
	flag.Parse()

	// Sem determines the number of files parsed in parallel
	sem := make(chan struct{}, 4)

	// Collects the list of all the names of the files found
	fileNames := []string{}

	for _, file := range flag.Args() {
		// For every file given in the args, traverse its directory recursively
		err := filepath.WalkDir(file, fs.WalkDirFunc(func(path string, d fs.DirEntry, err error) error {
			// Make sure that the file we are parsing is a `java` file, and not a directory
			if filepath.Ext(path) != ".java" || d.IsDir() {
				return nil
			}

			fileNames = append(fileNames, path)

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

	// Open a graphviz dot file for adding dependencies to
	dotfile, err := dot.New("graph.dot")
	if err != nil {
		panic(err)
	}
	defer dotfile.Close()

	for _, path := range fileNames {
		sourceCode, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
		if err != nil {
			panic(err)
		}

		n := tree.RootNode()

		if *dependencyTreeFlag {
			// The extracted file contains the name, full package, and relative imports
			extracted := ExtractImports(n, sourceCode)

			var graph dot.GraphItem = dotfile

			// Generate all the subgraphs for the package declaration
			for _, item := range extracted.Package.Scope {
				graph = graph.Subgraph(item)
			}

			edges := []string{}

			// Add the class's name, connected to all of its imports
			for _, imp := range extracted.Imports {
				var subgraph dot.GraphItem = dotfile
				for ind, item := range imp.Scope {
					if ind == len(imp.Scope)-1 {
						subgraph.AddNode(item)
						edges = append(edges, item)
					} else {
						subgraph = subgraph.Subgraph(item)
					}
				}
			}

			graph.AddNode(extracted.Name, edges...)

			dotfile.DeleteSubgraph("java")
			dotfile.DeleteSubgraph("javax")
			dotfile.DeleteSubgraph("com")
			dotfile.DeleteSubgraph("it")

			wg.Done()
			continue
		}

		log.Infof("Converting file \"%s\"", path)
		// Write to stdout by default
		var output io.Writer = os.Stdout

		// If write is specified, then write everything to the corresponding files
		if *writeFlag {
			outFile := path[:len(path)-len(filepath.Ext(path))] + ".go"
			output, err = os.Create(outFile)
			if err != nil {
				panic(fmt.Errorf("Error creating file %v: %v", outFile, err))
			}
		} else {
			// If quiet is specified, then discard the output
			if *quiet {
				output = io.Discard
			}
		}

		// If ast flag is specified, then print out go's formatted ast
		if *astFlag {
			ast.Print(token.NewFileSet(), ParseNode(n, sourceCode, Ctx{}).(ast.Node))
		}

		go func() {
			sem <- struct{}{}
			// Release the semaphore when done
			defer func() { <-sem }()
			err = printer.Fprint(output, token.NewFileSet(), ParseNode(n, sourceCode, Ctx{}).(ast.Node))
			if err != nil {
				panic(err)
			}
			wg.Done()
		}()
	}

	wg.Wait()

	dotfile.WriteToFile()
}

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

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
)

func main() {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())

	writeFlag := flag.Bool("w", false, "Whether to write the files to disk instead of stdout")
	quiet := flag.Bool("q", false, "Don't write to stdout on successful parse")
	astFlag := flag.Bool("ast", false, "Print out go's pretty-printed ast, instead of source code")
	flag.Parse()

	for _, file := range flag.Args() {
		// For every file given in the args, traverse its directory recursively
		err := filepath.WalkDir(file, fs.WalkDirFunc(func(path string, d fs.DirEntry, err error) error {
			// Make sure that the file we are parsing is a `java` file, and not a directory
			if filepath.Ext(path) != ".java" || d.IsDir() {
				return nil
			}
			sourceCode, err := os.ReadFile(path)
			if err != nil {
				panic(err)
			}
			tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
			if err != nil {
				panic(err)
			}

			n := tree.RootNode()

			fmt.Printf("Converting file \"%s\" ... ", path)
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

			err = printer.Fprint(output, token.NewFileSet(), ParseNode(n, sourceCode, Ctx{}).(ast.Node))
			if err != nil {
				panic(err)
			}
			fmt.Println("Success!")
			return nil
		}))
		if err != nil {
			panic(err)
		}
	}
}

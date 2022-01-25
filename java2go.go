package main

import (
	"context"
	"flag"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
)

func main() {
	writeFlag := flag.Bool("w", false, "Whether to write the files to disk instead of stdout")
	flag.Parse()
	for _, fileName := range flag.Args() {
		parser := sitter.NewParser()
		parser.SetLanguage(java.GetLanguage())

		sourceCode, err := os.ReadFile(fileName)
		if err != nil {
			panic(err)
		}
		tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
		if err != nil {
			panic(err)
		}

		n := tree.RootNode()

		if *writeFlag {
			outFile := fileName[:len(fileName)-len(filepath.Ext(fileName))] + ".go"
			out, err := os.Create(outFile)
			if err != nil {
				panic(fmt.Errorf("Error creating file %v: %v", outFile, err))
			}

			err = printer.Fprint(out, token.NewFileSet(), ParseNode(n, sourceCode, Ctx{}).(ast.Node))
			if err != nil {
				panic(err)
			}

			fmt.Printf("Successfully converted \"%v\" to \"%v\"\n", fileName, outFile)

			// Close the file to prevent future writes
			out.Close()
			continue
		}

		err = printer.Fprint(os.Stdout, token.NewFileSet(), ParseNode(n, sourceCode, Ctx{}).(ast.Node))
		if err != nil {
			panic(err)
		}
	}
}

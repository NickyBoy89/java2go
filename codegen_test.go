package main

import (
	"context"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
)

// ParseSourceAst parses a given source file and returns the tree-sitter root
// node for the AST associated with that file
func ParseSourceAst(fileName string) (*sitter.Node, []byte) {
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

	return tree.RootNode(), sourceCode
}

func ParseAst(fileName string) ast.Node {
	root, source := ParseSourceAst(fileName)
	return ParseNode(root, source, Ctx{}).(ast.Node)
}

func TestIncDec(t *testing.T) {
	// This tests the increment and decrement handling on increment and decrement
	// statements, as well as expressions

	err := printer.Fprint(os.Stdout, token.NewFileSet(), ParseAst("testfiles/IncrementDecrement.java"))
	if err != nil {
		t.Fatal(err)
	}
}

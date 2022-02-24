package main

import (
	"bytes"
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

// This tests the increment and decrement handling on increment and decrement
// statements, as well as expressions
func TestIncDec(t *testing.T) {
	var generated bytes.Buffer
	err := printer.Fprint(&generated, token.NewFileSet(), ParseAst("testfiles/IncrementDecrement.java"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(generated.String())
}

// This tests the variable assignment handling, using it as a statement as
// well as an expression
func TestAssignments(t *testing.T) {
	var generated bytes.Buffer
	err := printer.Fprint(&generated, token.NewFileSet(), ParseAst("testfiles/VariableAssignments.java"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(generated.String())
}

// This tests two alternate ways of calling the new constructor
func TestAlternateNewCall(t *testing.T) {
	var generated bytes.Buffer
	err := printer.Fprint(&generated, token.NewFileSet(), ParseAst("testfiles/SelectorNewExpression.java"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(generated.String())
}

// This tests for various combinations of init, cond, and post parts of for loops
func TestScrambledForLoops(t *testing.T) {
	var generated bytes.Buffer
	err := printer.Fprint(&generated, token.NewFileSet(), ParseAst("testfiles/ScrambledForLoops.java"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(generated.String())
}

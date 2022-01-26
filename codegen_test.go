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
	"gitlab.nicholasnovak.io/snapdragon/java2go/diffmatchpatch"
)

func ParseCodeInput(input []byte) []byte {

	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, input)
	if err != nil {
		panic(err)
	}

	var output bytes.Buffer

	err = printer.Fprint(&output, token.NewFileSet(), ParseNode(tree.RootNode(), input, Ctx{}).(ast.Node))
	if err != nil {
		panic(err)
	}
	return output.Bytes()
}

func TestSimpleTest(t *testing.T) {
	input, err := os.ReadFile("testfiles/Test.java")
	if err != nil {
		t.Fatalf("Could not read file \"testfiles/Test.java\": %v", err)
	}
	expected, err := os.ReadFile("testfiles/output/Test.go")
	if err != nil {
		t.Fatalf("Could not read \"testfiles/output/Test.go\": %v", err)
	}

	if string(ParseCodeInput(input)) != string(expected) {
		diff := diffmatchpatch.New()
		difference := diff.DiffMain(string(ParseCodeInput(input)), string(expected), false)
		t.Log("Input and expected did not match:")
		t.Error(diff.DiffPrettyText(difference))
	}
}

package main

import (
	"context"
	"os"
	"reflect"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
)

func loadFile(fileName string) ([]byte, *sitter.Tree) {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())

	source, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		panic(err)
	}
	return source, tree
}

func TestSimpleDeclaration(t *testing.T) {
	source, tree := loadFile("testfiles/typechecks/SimpleDeclaration.java")

	expected := TypeInformation{
		types: map[string]string{
			"main":     "",
			"args":     "[]string",
			"variable": "int32",
		},
	}

	info, err := ExtractTypeInformation(tree.RootNode(), source)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(info, expected) {
		t.Errorf("Actual: %v did not meet expected: %v", info, expected)
	}
}

func TestMethodDeclaration(t *testing.T) {
	source, tree := loadFile("testfiles/typechecks/MethodConstructorDeclaration.java")

	expected := TypeInformation{
		types: map[string]string{
			"sayHello": "string",
			"squared":  "int32",
			"n":        "int32",
			"someNum":  "float64",
		},
	}

	info, err := ExtractTypeInformation(tree.RootNode(), source)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(info, expected) {
		t.Errorf("Actual: %v did not meet expected: %v", info, expected)
	}
}

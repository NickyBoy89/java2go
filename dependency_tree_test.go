package main

import (
	"reflect"
	"testing"
)

func TestLinkedListImports(t *testing.T) {
	expected := []string{"java.lang.AssertionError"}
	ast, source := ParseSourceAst("testfiles/IntLinkedList.java")
	className, actualImports := ExtractImports(ast, source, "")
	if !reflect.DeepEqual(actualImports, expected) {
		t.Errorf("Expected: %v, Actual: %v", expected, actualImports)
	}

	if className != "IntLinkedList" {
		t.Errorf("Expected: %v, Actual: %v", "IntLinkedList", className)
	}
}

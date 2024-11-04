package nodeutil

import (
	"fmt"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

func AssertTypeIs(node *sitter.Node, expectedType string) {
	if node.Kind() != expectedType {
		panic(fmt.Sprintf("assertion failed: Type of node differs from expected: %s, got: %s", expectedType, node.Kind()))
	}
}

package ng

import (
	"go/ast"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

func IsComment(node sitter.Node) bool {
	switch node.Kind() {
	case "block_comment", "line_comment":
		return true
	}

	return false
}

func ParseBlockComment(node sitter.Node) *ast.CommentGroup {
	return &ast.CommentGroup{
		List: []*ast.Comment{
			{
				Text: node.Utf8Text(source),
			},
		},
	}
}

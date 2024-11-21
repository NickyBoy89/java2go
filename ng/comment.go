package ng

import sitter "github.com/tree-sitter/go-tree-sitter"

func IsComment(node sitter.Node) bool {
	switch node.Kind() {
	case "block_comment", "line_comment":
		return true
	}

	return false
}

package main

import (
	"fmt"
	"go/ast"

	sitter "github.com/smacker/go-tree-sitter"
)

func ParseStmt(node *sitter.Node, source []byte, ctx Ctx) ast.Stmt {
	if stmt := TryParseStmt(node, source, ctx); stmt != nil {
		return stmt
	}
	panic(fmt.Errorf("Unhandled stmt type: %v", node.Type()))
}

func TryParseStmt(node *sitter.Node, source []byte, ctx Ctx) ast.Stmt {
	switch node.Type() {
	case "assignment_expression":
		// Try to handle this assignment as a statement, and if not, don't do anything
		if stmt, ok := ParseNode(node, source, ctx).(ast.Stmt); ok {
			return stmt
		}
	case "update_expression":
		if stmt, ok := ParseNode(node, source, ctx).(ast.Stmt); ok {
			return stmt
		}
	}
	return nil
}

func ParseStmts(node *sitter.Node, source []byte, ctx Ctx) []ast.Stmt {
	if stmts := TryParseStmts(node, source, ctx); stmts != nil {
		return stmts
	}
	panic(fmt.Errorf("Unhandled stmts type: %v", node.Type()))
}

func TryParseStmts(node *sitter.Node, source []byte, ctx Ctx) []ast.Stmt {
	switch node.Type() {
	case "assignment_expression":
		if stmts, ok := ParseNode(node, source, ctx).([]ast.Stmt); ok {
			return stmts
		}
	}
	return nil
}

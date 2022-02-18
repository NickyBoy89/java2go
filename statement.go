package main

import (
	"fmt"
	"go/ast"
	"go/token"

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
	case "local_variable_declaration":
		// Ignore the name of the type being declared, because we are going to
		// infer that when the variable gets assigned
		if node.NamedChild(0).Type() == "modifiers" {
			return ParseStmt(node.NamedChild(2), source, ctx)
		}
		return ParseStmt(node.NamedChild(1), source, ctx)
	case "variable_declarator":
		var names, values []ast.Expr

		// If there is only one node, then that node is just a name
		if node.NamedChildCount() == 1 {
			names = append(names, ParseExpr(node.NamedChild(0), source, ctx))
		}

		// Loop through every pair of name and value
		for ind := 0; ind < int(node.NamedChildCount())-1; ind += 2 {
			names = append(names, ParseExpr(node.NamedChild(ind), source, ctx))
			values = append(values, ParseExpr(node.NamedChild(ind+1), source, ctx))
		}

		return &ast.AssignStmt{Lhs: names, Tok: token.DEFINE, Rhs: values}
	case "assignment_expression":
		// Try to handle this assignment as a statement, and if not, don't do anything
		if stmt, ok := ParseNode(node, source, ctx).(ast.Stmt); ok {
			return stmt
		}
	case "update_expression":
		if node.Child(0).Type() == "identifier" {
			return &ast.IncDecStmt{
				X:   ParseExpr(node.Child(0), source, ctx),
				Tok: StrToToken(node.Child(1).Content(source)),
			}
		}

		return &ast.IncDecStmt{
			X:   ParseExpr(node.Child(1), source, ctx),
			Tok: StrToToken(node.Child(0).Content(source)),
		}
	case "return_statement":
		if node.NamedChildCount() < 1 {
			return &ast.ReturnStmt{Results: []ast.Expr{}}
		}
		return &ast.ReturnStmt{Results: []ast.Expr{ParseExpr(node.NamedChild(0), source, ctx)}}
	case "labeled_statement":
		return &ast.LabeledStmt{
			Label: ParseNode(node.NamedChild(0), source, ctx).(*ast.Ident),
			Stmt:  ParseNode(node.NamedChild(1), source, ctx).(ast.Stmt),
		}
	case "break_statement":
		return &ast.BranchStmt{Tok: token.BREAK}
	case "continue_statement":
		return &ast.BranchStmt{Tok: token.CONTINUE}
	case "throw_statement":
		return &ast.ExprStmt{X: &ast.CallExpr{
			Fun:  &ast.Ident{Name: "panic"},
			Args: []ast.Expr{ParseExpr(node.NamedChild(0), source, ctx)},
		}}
	case "if_statement":
		var cond ast.Expr
		var body *ast.BlockStmt
		var elseStmt ast.Stmt

		for _, c := range Children(node) {
			switch c.Type() {
			case "parenthesized_expression":
				cond = ParseExpr(c, source, ctx)
			case "block": // First block is the `if`, second is the `else`
				if body == nil {
					body = ParseNode(c, source, ctx).(*ast.BlockStmt)
				} else {
					elseStmt = ParseNode(c, source, ctx).(*ast.BlockStmt)
				}
			}
		}

		return &ast.IfStmt{
			Cond: cond,
			Body: body,
			Else: elseStmt,
		}
	case "for_statement":
		return &ast.ForStmt{
			Init: ParseStmt(node.NamedChild(0), source, ctx),
			Cond: ParseExpr(node.NamedChild(1), source, ctx),
			Post: ParseStmt(node.NamedChild(2), source, ctx),
			Body: ParseNode(node.NamedChild(3), source, ctx).(*ast.BlockStmt),
		}
	case "while_statement":
		return &ast.ForStmt{
			Cond: ParseExpr(node.NamedChild(0), source, ctx),
			Body: ParseNode(node.NamedChild(1), source, ctx).(*ast.BlockStmt),
		}
	case "do_statement":
		// A do statement is handled as a blank for loop with the condition
		// inserted as a break condition in the final part of the loop
		body := ParseNode(node.NamedChild(0), source, ctx).(*ast.BlockStmt)

		body.List = append(body.List, &ast.IfStmt{
			Cond: &ast.UnaryExpr{
				X: &ast.ParenExpr{
					X: ParseExpr(node.NamedChild(1), source, ctx),
				},
			},
			Body: &ast.BlockStmt{List: []ast.Stmt{&ast.BranchStmt{Tok: token.BREAK}}},
		})

		return &ast.ForStmt{
			Body: body,
		}
	case "switch_statement":
		return &ast.SwitchStmt{
			Tag:  ParseExpr(node.NamedChild(0), source, ctx),
			Body: ParseNode(node.NamedChild(1), source, ctx).(*ast.BlockStmt),
		}
	case "expression_statement":
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

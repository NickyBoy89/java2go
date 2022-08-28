package main

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/NickyBoy89/java2go/astutil"
	"github.com/NickyBoy89/java2go/nodeutil"
	log "github.com/sirupsen/logrus"
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
	case "ERROR":
		log.WithFields(log.Fields{
			"parsed":    node.Content(source),
			"className": ctx.className,
		}).Warn("Statement parse error")
		return &ast.BadStmt{}
	case "comment":
		return &ast.BadStmt{}
	case "local_variable_declaration":
		variableType := astutil.ParseType(node.ChildByFieldName("type"), source)
		variableDeclarator := node.ChildByFieldName("declarator")

		// If a variable is being declared, but not set to a value
		// Ex: `int value;`
		if variableDeclarator.NamedChildCount() == 1 {
			return &ast.DeclStmt{
				Decl: &ast.GenDecl{
					Tok: token.VAR,
					Specs: []ast.Spec{
						&ast.ValueSpec{
							Names: []*ast.Ident{ParseExpr(variableDeclarator.ChildByFieldName("name"), source, ctx).(*ast.Ident)},
							Type:  variableType,
						},
					},
				},
			}
		}

		ctx.lastType = variableType

		declaration := ParseStmt(variableDeclarator, source, ctx).(*ast.AssignStmt)

		// Now, if a variable is assigned to `null`, we can't infer its type, so
		// don't throw out the type information associated with it
		var containsNull bool

		// Go through the values and see if there is a `null_literal`
		for _, child := range nodeutil.NamedChildrenOf(variableDeclarator) {
			if child.Type() == "null_literal" {
				containsNull = true
				break
			}
		}

		names := make([]*ast.Ident, len(declaration.Lhs))
		for ind, decl := range declaration.Lhs {
			names[ind] = decl.(*ast.Ident)
		}

		// If the declaration contains null, declare it with the `var` keyword instead
		// of implicitly
		if containsNull {
			return &ast.DeclStmt{
				Decl: &ast.GenDecl{
					Tok: token.VAR,
					Specs: []ast.Spec{
						&ast.ValueSpec{
							Names:  names,
							Type:   variableType,
							Values: declaration.Rhs,
						},
					},
				},
			}
		}

		return declaration
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
		assignVar := ParseExpr(node.Child(0), source, ctx)
		assignVal := ParseExpr(node.Child(2), source, ctx)

		// Unsigned right shift
		if node.Child(1).Content(source) == ">>>=" {
			return &ast.ExprStmt{X: &ast.CallExpr{
				Fun:  &ast.Ident{Name: "UnsignedRightShiftAssignment"},
				Args: []ast.Expr{assignVar, assignVal},
			}}
		}

		return &ast.AssignStmt{
			Lhs: []ast.Expr{assignVar},
			Tok: StrToToken(node.Child(1).Content(source)),
			Rhs: []ast.Expr{assignVal},
		}
	case "update_expression":
		if node.Child(0).IsNamed() {
			return &ast.IncDecStmt{
				X:   ParseExpr(node.Child(0), source, ctx),
				Tok: StrToToken(node.Child(1).Content(source)),
			}
		}

		return &ast.IncDecStmt{
			X:   ParseExpr(node.Child(1), source, ctx),
			Tok: StrToToken(node.Child(0).Content(source)),
		}
	case "resource_specification":
		return ParseStmt(node.NamedChild(0), source, ctx)
	case "resource":
		var offset int
		if node.NamedChild(0).Type() == "modifiers" {
			offset = 1
		}
		return &ast.AssignStmt{
			Lhs: []ast.Expr{ParseExpr(node.NamedChild(1+offset), source, ctx)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{ParseExpr(node.NamedChild(2+offset), source, ctx)},
		}
	case "method_invocation":
		return &ast.ExprStmt{X: ParseExpr(node, source, ctx)}
	case "constructor_body", "block":
		body := &ast.BlockStmt{}
		for _, line := range nodeutil.NamedChildrenOf(node) {
			if line.Type() == "comment" {
				continue
			}
			if stmt := TryParseStmt(line, source, ctx); stmt != nil {
				body.List = append(body.List, stmt)
			} else {
				// Try statements are ignored, so they return a list of statements
				body.List = append(body.List, ParseNode(line, source, ctx).([]ast.Stmt)...)
			}
		}
		return body
	case "expression_statement":
		if stmt := TryParseStmt(node.NamedChild(0), source, ctx); stmt != nil {
			return stmt
		}
		return &ast.ExprStmt{X: ParseExpr(node.NamedChild(0), source, ctx)}
	case "explicit_constructor_invocation":
		// This is when a constructor calls another constructor with the use of
		// something such as `this(args...)`
		return &ast.ExprStmt{
			&ast.CallExpr{
				Fun:  &ast.Ident{Name: "New" + ctx.className},
				Args: ParseNode(node.NamedChild(1), source, ctx).([]ast.Expr),
			},
		}
	case "return_statement":
		if node.NamedChildCount() < 1 {
			return &ast.ReturnStmt{Results: []ast.Expr{}}
		}
		return &ast.ReturnStmt{Results: []ast.Expr{ParseExpr(node.NamedChild(0), source, ctx)}}
	case "labeled_statement":
		return &ast.LabeledStmt{
			Label: ParseExpr(node.NamedChild(0), source, ctx).(*ast.Ident),
			Stmt:  ParseStmt(node.NamedChild(1), source, ctx),
		}
	case "break_statement":
		if node.NamedChildCount() > 0 {
			return &ast.BranchStmt{Tok: token.BREAK, Label: ParseExpr(node.NamedChild(0), source, ctx).(*ast.Ident)}
		}
		return &ast.BranchStmt{Tok: token.BREAK}
	case "continue_statement":
		if node.NamedChildCount() > 0 {
			return &ast.BranchStmt{Tok: token.CONTINUE, Label: ParseExpr(node.NamedChild(0), source, ctx).(*ast.Ident)}
		}
		return &ast.BranchStmt{Tok: token.CONTINUE}
	case "throw_statement":
		return &ast.ExprStmt{X: &ast.CallExpr{
			Fun:  &ast.Ident{Name: "panic"},
			Args: []ast.Expr{ParseExpr(node.NamedChild(0), source, ctx)},
		}}
	case "if_statement":
		var other ast.Stmt
		if node.ChildByFieldName("alternative") != nil {
			other = ParseStmt(node.ChildByFieldName("alternative"), source, ctx)
		}
		return &ast.IfStmt{
			Cond: ParseExpr(node.ChildByFieldName("condition"), source, ctx),
			Body: ParseStmt(node.ChildByFieldName("consequence"), source, ctx).(*ast.BlockStmt),
			Else: other,
		}
	case "enhanced_for_statement":
		// An enhanced for statement has the following fields:
		// variables for the variable being declared (ex: int n)
		// then the expression that is being ranged over
		// and finally, the block of the expression

		total := int(node.NamedChildCount())

		return &ast.RangeStmt{
			// We don't need the type of the variable for the range expression
			Key:   &ast.Ident{Name: "_"},
			Value: ParseExpr(node.NamedChild(total-3), source, ctx),
			Tok:   token.DEFINE,
			X:     ParseExpr(node.NamedChild(total-2), source, ctx),
			Body:  ParseStmt(node.NamedChild(total-1), source, ctx).(*ast.BlockStmt),
		}
	case "for_statement":
		var init, post ast.Stmt
		if node.ChildByFieldName("init") != nil {
			init = ParseStmt(node.ChildByFieldName("init"), source, ctx)
		}
		if node.ChildByFieldName("update") != nil {
			post = ParseStmt(node.ChildByFieldName("update"), source, ctx)
		}
		var cond ast.Expr
		if node.ChildByFieldName("condition") != nil {
			cond = ParseExpr(node.ChildByFieldName("condition"), source, ctx)
		}

		return &ast.ForStmt{
			Init: init,
			Cond: cond,
			Post: post,
			Body: ParseStmt(node.ChildByFieldName("body"), source, ctx).(*ast.BlockStmt),
		}
	case "while_statement":
		return &ast.ForStmt{
			Cond: ParseExpr(node.NamedChild(0), source, ctx),
			Body: ParseStmt(node.NamedChild(1), source, ctx).(*ast.BlockStmt),
		}
	case "do_statement":
		// A do statement is handled as a blank for loop with the condition
		// inserted as a break condition in the final part of the loop
		body := ParseStmt(node.NamedChild(0), source, ctx).(*ast.BlockStmt)

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
			Body: ParseStmt(node.NamedChild(1), source, ctx).(*ast.BlockStmt),
		}
	case "switch_block":
		switchBlock := &ast.BlockStmt{}
		var currentCase *ast.CaseClause
		for _, c := range nodeutil.NamedChildrenOf(node) {
			switch c.Type() {
			case "switch_label":
				// When a new switch label comes, append it to the switch block
				if currentCase != nil {
					switchBlock.List = append(switchBlock.List, currentCase)
				}
				currentCase = ParseNode(c, source, ctx).(*ast.CaseClause)
			default:
				if exprs := TryParseStmts(c, source, ctx); exprs != nil {
					currentCase.Body = append(currentCase.Body, exprs...)
				} else {
					currentCase.Body = append(currentCase.Body, ParseStmt(c, source, ctx))
				}
			}
		}

		return switchBlock
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
	case "try_statement":
		if stmts, ok := ParseNode(node, source, ctx).([]ast.Stmt); ok {
			return stmts
		}
	}
	return nil
}

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
	case "ERROR":
		return &ast.BadStmt{}
	case "local_variable_declaration":
		var varTypeIndex int

		// The first child can either be modifiers e.g `final int var = 1`, or
		// just the variable's type
		if node.NamedChild(0).Type() == "modifiers" {
			varTypeIndex = 1
		}

		// The variable declarator does not have a value
		if node.NamedChild(varTypeIndex+1).NamedChildCount() == 1 {
			return &ast.DeclStmt{
				Decl: &ast.GenDecl{
					Tok: token.VAR,
					Specs: []ast.Spec{
						&ast.ValueSpec{
							Names: []*ast.Ident{ParseExpr(node.NamedChild(varTypeIndex+1).NamedChild(0), source, ctx).(*ast.Ident)},
							Type:  ParseExpr(node.NamedChild(varTypeIndex), source, ctx),
						},
					},
				},
			}
		}

		// Ignore the type being declared, because it is going to be inferred
		return ParseStmt(node.NamedChild(varTypeIndex+1), source, ctx)
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
		if node.Child(1).Content(source) == ">>>=" {
			return &ast.ExprStmt{X: &ast.CallExpr{
				Fun:  &ast.Ident{Name: "UnsignedRightShiftAssignment"},
				Args: []ast.Expr{ParseExpr(node.Child(0), source, ctx), ParseExpr(node.Child(2), source, ctx)},
			}}
		}
		return &ast.AssignStmt{
			Lhs: []ast.Expr{ParseExpr(node.Child(0), source, ctx)},
			Tok: StrToToken(node.Child(1).Content(source)),
			Rhs: []ast.Expr{ParseExpr(node.Child(2), source, ctx)},
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
	case "method_invocation":
		return &ast.ExprStmt{X: ParseExpr(node, source, ctx)}
	case "constructor_body", "block":
		body := &ast.BlockStmt{}
		for _, line := range Children(node) {
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
					body = ParseStmt(c, source, ctx).(*ast.BlockStmt)
				} else {
					elseStmt = ParseStmt(c, source, ctx).(*ast.BlockStmt)
				}
			}
		}

		return &ast.IfStmt{
			Cond: cond,
			Body: body,
			Else: elseStmt,
		}
	case "enhanced_for_statement":
		// An enhanced for statement has the following fields:
		// variables for the variable being declared (ex: int n)
		// then the expression that is being ranged over
		// and finally, the block of the expression

		total := int(node.NamedChildCount())

		return &ast.RangeStmt{
			Key:   ParseExpr(node.NamedChild(total-4), source, ctx),
			Value: ParseExpr(node.NamedChild(total-3), source, ctx),
			Tok:   token.DEFINE,
			X:     ParseExpr(node.NamedChild(total-2), source, ctx),
			Body:  ParseStmt(node.NamedChild(total-1), source, ctx).(*ast.BlockStmt),
		}
	case "for_statement":
		// The different parts of the for loop
		var cond ast.Expr
		var init, post ast.Stmt

		var ignoreInit, ignoreCond, ignorePost bool

		var multipleAssignment bool

		// If the init is a statement, then it will have a semicon associated with
		// it, instead of an expression, where the semicolon will be separate
		var initClosed bool

		for _, c := range UnnamedChildren(node) {
			if c.IsNamed() {
				if multipleAssignment {
					otherAssign := ParseStmt(c, source, ctx).(*ast.AssignStmt)
					init.(*ast.AssignStmt).Lhs = append(init.(*ast.AssignStmt).Lhs, otherAssign.Lhs...)
					init.(*ast.AssignStmt).Rhs = append(init.(*ast.AssignStmt).Rhs, otherAssign.Rhs...)
					multipleAssignment = false
				} else if !ignoreInit && init == nil {
					if c.Child(int(c.ChildCount())-1).Content(source) == ";" {
						initClosed = true
					}
					ignoreInit = true
					init = ParseStmt(c, source, ctx)
				} else if !ignoreCond && cond == nil {
					ignoreCond = true
					cond = ParseExpr(c, source, ctx)
				} else if !ignorePost && post == nil {
					ignorePost = true
					post = ParseStmt(c, source, ctx)
				}
			} else {
				switch c.Content(source) {
				case ";":
					// If we encounter a semicolon

					// If there has been
					// If the init is missing, there will be only a semicolon left
					if !initClosed {
						initClosed = true
						ignoreInit = true
					} else if !ignoreInit && init == nil {
						ignoreInit = true
					} else if !ignoreCond && cond == nil {
						ignoreCond = true
					}
				case ",":
					switch init.(type) {
					case *ast.AssignStmt:
						multipleAssignment = true
					default:
						panic("Init had multiple non-assignments in init")
					}
				case ")":
					// Once the parenthesies close, stop parsing the for loop
					break
				}
			}
		}

		return &ast.ForStmt{
			Init: init,
			Cond: cond,
			Post: post,
			Body: ParseStmt(node.NamedChild(int(node.NamedChildCount())-1), source, ctx).(*ast.BlockStmt),
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
		for _, c := range Children(node) {
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

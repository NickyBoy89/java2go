package ng

import (
	"go/ast"
	"go/token"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

func ParseReturnStatement(node sitter.Node) *ast.ReturnStmt {
	res := []ast.Expr{}

	// If there is a return value
	if node.NamedChildCount() > 0 {
		// TODO: Find why `field_access` is allowed into this function
		// Maybe something to do with supertypes?
		expr, err := ParseExpression(*node.NamedChild(0))
		if err != nil {
			panic(err)
		}
		res = append(res, expr)
	}

	return &ast.ReturnStmt{
		Results: res,
	}
}

func ParseIfStatement(node sitter.Node) (*ast.IfStmt, error) {
	cond, err := ParseParenthesizedExpression(*node.ChildByFieldName("condition"))
	if err != nil {
		return nil, err
	}

	res, err := ParseStatement(*node.ChildByFieldName("consequence"))
	if err != nil {
		return nil, err
	}

	var elseCase ast.Stmt
	elseNode := node.ChildByFieldName("alternative")
	if elseNode != nil {
		s, err := ParseStatement(*elseNode)
		if err != nil {
			return nil, err
		}
		elseCase = s.(ast.Stmt)
	}

	return &ast.IfStmt{
		Init: nil,
		Cond: cond,
		Body: &ast.BlockStmt{
			// TODO: Handle declaration statements
			List: []ast.Stmt{res.(ast.Stmt)},
		},
		Else: elseCase,
	}, nil
}

func ParseExpressionStatement(node sitter.Node) (*ast.ExprStmt, error) {
	expr, err := ParseExpression(*node.NamedChild(0))
	return &ast.ExprStmt{
		X: expr,
	}, err
}

// TODO: Update all call sides for this function
func ParseVariableDeclaratorList(node sitter.Node) []*ast.Ident {
	decls := []*ast.Ident{}
	for _, decl := range node.ChildrenByFieldName("declarator", node.Walk()) {
		decls = append(decls, ParseVariableDeclarator(decl))
	}

	return decls
}

// `local_variable_declaration`
// Structure:
// local_variable_declaration: $ => seq(
//
//	optional($.modifiers),
//	field('type', $._unannotated_type),
//	$._variable_declarator_list,
//	';',
//
// ),
func ParseLocalVariableDeclaration(node sitter.Node) (*ast.DeclStmt, error) {
	mods, err := HandleModifiers(node)
	if err != nil {
		return nil, err
	}
	_ = mods

	typ, err := ParseUnannotatedType(*node.ChildByFieldName("type"))
	if err != nil {
		return nil, err
	}

	return &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names:  ParseVariableDeclaratorList(node),
					Type:   typ,
					Values: []ast.Expr{},
				},
			},
		},
	}, nil
}

// `for_statement`
// Structure:
// for_statement: $ => seq(
//
//	'for', '(',
//	choice(
//	  field('init', $.local_variable_declaration),
//	  seq(
//	    commaSep(field('init', $.expression)),
//	    ';',
//	  ),
//	),
//	field('condition', optional($.expression)), ';',
//	commaSep(field('update', $.expression)), ')',
//	field('body', $.statement),
//
// ),
func ParseForStatement(node sitter.Node) (*ast.ForStmt, error) {
	s, err := ParseStatement(*node.ChildByFieldName("body"))
	if err != nil {
		return nil, err
	}

	// TODO: Implement the rest of the for loop
	return &ast.ForStmt{
		Init: nil,
		Cond: nil,
		Post: nil,
		Body: &ast.BlockStmt{
			// TODO: Fix this for non-inline for loops
			List: []ast.Stmt{s.(ast.Stmt)},
		},
	}, nil
}

func ParseThrowStatement(node sitter.Node) (*ast.ExprStmt, error) {
	expr, err := ParseExpression(*node.NamedChild(0))
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun:  ast.NewIdent("panic"),
			Args: []ast.Expr{expr},
		},
	}, err
}

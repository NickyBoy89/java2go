package ng

import (
	"go/ast"

	"github.com/NickyBoy89/java2go/ng/codegen"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

// `expression` struture:
// $.assignment_expression,
// $.binary_expression,
// $.instanceof_expression,
// $.lambda_expression,
// $.ternary_expression,
// $.update_expression,
// $.primary_expression,
// $.unary_expression,
// $.cast_expression,
// $.switch_expression,
func ParseExpression(node sitter.Node) (ast.Expr, error) {
	switch node.Kind() {
	case "assignment_expression":
		return ParseAssignmentExpression(node)
	case "binary_expression":
		return ParseBinaryExpression(node)
	case "instanceof_expression":
	case "lambda_expression":
	case "ternary_expression":
	case "update_expression":
		return ParseUpdateExpression(node)
	case "primary_expression":
	case "unary_expression":
		return ParseUnaryExpression(node)
	case "cast_expression":
	case "switch_expression":
	}

	return ParsePrimaryExpression(node)

	panic("TODO: Unhandled expression: " + node.Kind())
}

// `primary_expression` structure:
// &._literal,
// $.class_literal,
// $.this,
// $.identifier,
// $._reserved_identifier,
// $.parenthesized_expression,
// $.object_creation_expression,
// $.field_access,
// $.array_access,
// $.method_invocation,
// $.method_reference,
// $.array_creation_expression,
// $.template_expression,
func ParsePrimaryExpression(node sitter.Node) (ast.Expr, error) {
	expr, err := TryParsePrimaryExpression(node)
	if err != nil {
		return nil, err
	}

	if expr == nil {
		panic("TODO: Unhandled primary expression: " + node.Kind())
	}

	return expr, err
}

func TryParsePrimaryExpression(node sitter.Node) (ast.Expr, error) {
	// TODO:
	// &._literal,
	// $._reserved_identifier,

	if lit := TryParseLiteral(node); lit != nil {
		return lit, nil
	}

	switch node.Kind() {
	case "class_literal":
	case "this":
		return ParseThis(node), nil
	case "identifier":
		return ast.NewIdent(ParseIdentifier(node)), nil
	case "parenthesized_expression":
		return ParseParenthesizedExpression(node)
	case "object_creation_expression":
		return ParseObjectCreationExpression(node)
	case "field_access":
		return ParseFieldAccess(node), nil
	case "array_access":
		return ParseArrayAccess(node)
	case "method_invocation":
		return ParseMethodInvocation(node)
	case "method_reference":
	case "array_creation_expression":
		return ParseArrayCreationExpression(node), nil
	case "template_expression":
	}

	return nil, nil
}

func ParseParenthesizedExpression(node sitter.Node) (*ast.ParenExpr, error) {
	expr, err := ParseExpression(*node.NamedChild(0))
	return &ast.ParenExpr{
		X: expr,
	}, err
}

func ParseBinaryExpression(node sitter.Node) (*ast.BinaryExpr, error) {
	l, err := ParseExpression(*node.ChildByFieldName("left"))
	if err != nil {
		return nil, err
	}
	r, err := ParseExpression(*node.ChildByFieldName("right"))
	if err != nil {
		return nil, err
	}
	return &ast.BinaryExpr{
		X:  l,
		Y:  r,
		Op: codegen.StrToToken(node.ChildByFieldName("operator").Kind()),
	}, nil
}

// `field_access` structure:
// field_access: $ => seq(
//
//	field('object', choice($.primary_expression, $.super)),
//	optional(seq(
//	  '.',
//	  $.super,
//	)),
//	'.',
//	field('field', choice($.identifier, $._reserved_identifier, $.this)),
//
// ),
func ParseFieldAccess(node sitter.Node) *ast.SelectorExpr {
	objNode := *node.ChildByFieldName("object")

	if nxt := objNode.NextNamedSibling(); nxt != nil && nxt.Kind() == "super" {
		panic("TODO: Superclass subclass access not implemented")
	}

	if objNode.Kind() == "super" {
		panic("TODO: Superclass access not implemented")
	}

	obj, err := ParsePrimaryExpression(objNode)
	if err != nil {
		panic(err)
	}

	var sel *ast.Ident

	switch fieldNode := *node.ChildByFieldName("field"); fieldNode.Kind() {
	case "this":
		sel = ast.NewIdent("this")
	case "identifier":
		sel = ast.NewIdent(ParseIdentifier(fieldNode))
	default:
		sel = parseReservedIdentifier(fieldNode)
	}

	// X.Sel
	return &ast.SelectorExpr{
		X:   obj,
		Sel: sel,
	}
}

// `update_expression` is a bit difficult
// Structure:
// seq($.expression, '++'),
// seq($.expression, '--'),
// seq('++', $.expression),
// seq('--', $.expression),
//
// Go only supports statements of the form `expr += 1` instead of `expr++`
// However, this is not an expression, and cannot be used as a value, so we
// have to replace this with the use of an inline function
func ParseUpdateExpression(node sitter.Node) (*ast.CallExpr, error) {
	var updateFunctionName *ast.Ident

	// Post-update expression, e.g. `i++`
	if node.Child(0).IsNamed() {
		updateFunctionName = ast.NewIdent("PostUpdate")
	} else {
		updateFunctionName = ast.NewIdent("PreUpdate")
	}

	expr, err := ParseExpression(*node.NamedChild(0))

	return &ast.CallExpr{
		Fun:  updateFunctionName,
		Args: []ast.Expr{expr},
	}, err
}

// `assignment_expression` is also a bit difficult
// Structure:
// assignment_expression: $ => prec.right(PREC.ASSIGN, seq(
//
//	field('left', choice(
//	  $.identifier,
//	  $._reserved_identifier,
//	  $.field_access,
//	  $.array_access,
//	)),
//	field('operator', choice('=', '+=', '-=', '*=', '/=', '&=', '|=', '^=', '%=', '<<=', '>>=', '>>>=')),
//	field('right', $.expression),
//
// )),
//
// Go's assignments are only in the form of statements, so they cannot be used
// as values, similar to Java's. This makes statements such as `x = a = 1` invalid.
//
// We replace these with an inline function
func ParseAssignmentExpression(node sitter.Node) (*ast.CallExpr, error) {
	leftNode := *node.ChildByFieldName("left")

	var err error
	var left ast.Expr

	switch leftNode.Kind() {
	case "field_access":
		left = ParseFieldAccess(leftNode)
	case "array_access":
		left, err = ParseArrayAccess(leftNode)
	default:
		ident := tryParseReservedIdentifier(leftNode)
		if ident == nil {
			left = ast.NewIdent(ParseIdentifier(leftNode))
		} else {
			left = ident
		}
	}

	if err != nil {
		panic(err)
	}

	rightExpr, err := ParseExpression(*node.ChildByFieldName("right"))
	if err != nil {
		panic(err)
	}

	// TODO: Handle the rest of the function call
	return &ast.CallExpr{
		Fun: ast.NewIdent("AssignmentExpression"),
		Args: []ast.Expr{
			left,
			codegen.AstString(node.ChildByFieldName("operator").Kind()),
			rightExpr,
		},
	}, nil
}

// `method_invocation`
// Structure:
// method_invocation: $ => seq(
//
//	choice(
//	  field('name', choice($.identifier, $._reserved_identifier)),
//	  seq(
//	    field('object', choice($.primary_expression, $.super)),
//	    '.',
//	    optional(seq(
//	      $.super,
//	      '.',
//	    )),
//	    field('type_arguments', optional($.type_arguments)),
//	    field('name', choice($.identifier, $._reserved_identifier)),
//	  ),
//	),
//	field('arguments', $.argument_list),
//
// ),
func ParseMethodInvocation(node sitter.Node) (*ast.CallExpr, error) {
	objNode := node.ChildByFieldName("object")
	if objNode != nil {
		if objNode.Kind() == "super" {
			panic("TODO: Implement superclasses")
		}
		obj, err := ParsePrimaryExpression(*objNode)
		if err != nil {
			return nil, err
		}

		if objNode.NextNamedSibling().Kind() == "super" {
			panic("TODO: Implement second superclass")
		}

		UnimplementedField(node, "type_arguments")

		_ = obj
		// TODO: Implement full object support for methods
	}

	name := ParseIdentifier(*node.ChildByFieldName("name"))

	args, err := ParseArgumentList(*node.ChildByFieldName("arguments"))
	if err != nil {
		return nil, err
	}

	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(MethodReceiverName),
			Sel: ast.NewIdent(name),
		},
		Args: args,
	}, nil
}

// TODO: Change this to call into generated code constructors
func ParseObjectCreationExpression(node sitter.Node) (*ast.CallExpr, error) {
	// TODO: Implement this
	return &ast.CallExpr{
		Fun:  ast.NewIdent("object_creation_expression"),
		Args: []ast.Expr{},
	}, nil
}

func ParseArrayAccess(node sitter.Node) (*ast.IndexExpr, error) {
	arr, err := ParsePrimaryExpression(*node.ChildByFieldName("array"))
	if err != nil {
		return nil, err
	}

	expr, err := ParseExpression(*node.ChildByFieldName("index"))
	if err != nil {
		return nil, err
	}

	return &ast.IndexExpr{
		X:     arr,
		Index: expr,
	}, nil
}

func ParseUnaryExpression(node sitter.Node) (*ast.UnaryExpr, error) {
	expr, err := ParseExpression(*node.ChildByFieldName("operand"))
	return &ast.UnaryExpr{
		Op: codegen.StrToToken(node.ChildByFieldName("operator").Kind()),
		X:  expr,
	}, err
}

// array_creation_expression: $ => prec.right(seq(
//
//	'new',
//	repeat($._annotation),
//	field('type', $._simple_type),
//	choice(
//		seq(
//			field('dimensions', repeat1($.dimensions_expr)),
//			field('dimensions', optional($.dimensions)),
//		),
//		seq(
//			field('dimensions', $.dimensions),
//			field('value', $.array_initializer),
//		),
//	),
//
// )),
func ParseArrayCreationExpression(node sitter.Node) *ast.CallExpr {
	// TODO: Handle annotations

	arrayType, err := ParseSimpleType(*node.ChildByFieldName("type"))
	if err != nil {
		panic(err)
	}

	for _, dimensionNode := range node.ChildrenByFieldName("dimensions", node.Walk()) {
		switch dimensionNode.Kind() {
		case "dimensions_expr":
		case "dimensions":
		}
	}

	// TODO: Handle dimensions
	// initialElementCount := ParseArrayInitializer(*node.ChildByFieldName("value")

	return &ast.CallExpr{
		Fun: ast.NewIdent("make"),
		Args: []ast.Expr{
			&ast.ArrayType{
				Elt: arrayType,
			},
			ast.NewIdent("0"),
		},
	}
}

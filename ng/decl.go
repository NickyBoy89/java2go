package ng

import (
	"go/ast"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// `constructor_declaration` structure:
// 1. optional($.modifiers),
// 2. $._constructor_declarator,
//  1. field('type_parameters', optional($.type_parameters)),
//  2. field('name', $.identifier),
//  3. field('parameters', $.formal_parameters),
//
// 3. optional($.throws),
// 4. field('body', $.constructor_body),
func ParseConstructorDeclaration(node sitter.Node) (ast.Decl, error) {
	mods, err := HandleModifiers(node)
	if err != nil {
		return nil, err
	}
	_ = mods

	UnimplementedField(node, "type_parameters")

	nameNode := *node.ChildByFieldName("name")
	params, err := ParseFormalParameters(*node.ChildByFieldName("parameters"))
	if err != nil {
		return nil, err
	}

	body, err := ParseConstructorBody(*node.ChildByFieldName("body"))
	if err != nil {
		return nil, err
	}

	name := ParseIdentifier(nameNode)

	return &ast.FuncDecl{
		Doc:  nil,
		Name: &ast.Ident{Name: name},
		Recv: nil,
		Type: &ast.FuncType{
			Params: params,
			Results: &ast.FieldList{
				List: []*ast.Field{},
			},
		},
		Body: body,
	}, nil
}

// `method_declaration` structure:
// 1. optional($.modifiers),
// 2. $._method_header,
// _method_header: $ => seq(
//
//	  optional(seq(
//	    field('type_parameters', $.type_parameters),
//	    repeat($._annotation),
//	  )),
//	  field('type', $._unannotated_type),
//	  $._method_declarator,
//	  optional($.throws),
//	),
//
// 3. choice(field('body', $.block), ';'),
//
// The parsing of a field is a bit ambigous here, because in Java, a field
// declaration is a top-level declaration, while in Go, we want it to live as
// a struct
func ParseFieldDeclaration(node sitter.Node) (*ast.Field, error) {
	mods, err := HandleModifiers(node)
	if err != nil {
		return nil, err
	}
	_ = mods

	fieldType, err := ParseUnannotatedType(*node.ChildByFieldName("type"))
	if err != nil {
		return nil, err
	}

	_ = fieldType

	fieldVars := []*ast.Ident{}

	cursor := node.Walk()
	for _, decl := range node.ChildrenByFieldName("declarator", cursor) {
		fieldVars = append(fieldVars, ParseVariableDeclarator(decl))
	}

	// A field declaration in a struct type
	return &ast.Field{
		Names: fieldVars,
		Type:  fieldType,
	}, nil
}

func ParseMethodDeclaration(node sitter.Node) (*ast.FuncDecl, error) {
	mods, err := HandleModifiers(node)
	if err != nil {
		return nil, err
	}

	// TODO: Handle reserved identifiers
	name := node.ChildByFieldName("name").Utf8Text(source)
	name = HandleAccessModifierRename(name, mods)

	// TODO: Handle type parameters
	UnimplementedField(node, "type_parameters")

	bodyNode := node.ChildByFieldName("body")
	// An empty method means that the function is meant to be filled by the
	// classes that implement it
	if bodyNode == nil {
		panic("TODO: Handle empty methods with code generation")
	}

	body, err := ParseBlock(*bodyNode)
	if err != nil {
		return nil, err
	}

	// TODO: Method header

	params, err := ParseFormalParameters(*node.ChildByFieldName("parameters"))
	if err != nil {
		return nil, err
	}

	UnimplementedField(node, "dimensions")

	return &ast.FuncDecl{
		Doc:  nil,
		Name: &ast.Ident{Name: name},
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: MethodReceiverName}},
					// TODO: We want the name of the class here, which means that we should pass in the class's context as well
					Type: &ast.StarExpr{X: &ast.Ident{Name: "temp"}},
				},
			},
		},
		Type: &ast.FuncType{
			Params: params,
			Results: &ast.FieldList{
				List: []*ast.Field{},
			},
		},
		Body: body,
	}, nil
}

package ng

import (
	"go/ast"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

func ParseTypeParameters(node sitter.Node) *ast.FieldList {
	typeParams := &ast.FieldList{List: []*ast.Field{}}
	for _, child := range node.NamedChildren(node.Walk()) {
		typeParams.List = append(typeParams.List, ParseTypeParameter(child))
	}

	return typeParams
}

func ParseTypeParameter(node sitter.Node) *ast.Field {
	// TODO: Handle annotation

	var typeName string

	for _, child := range node.NamedChildren(node.Walk()) {
		switch child.Kind() {
		case "type_identifier":
			typeName = ParseIdentifier(child)
		}
	}

	// TODO: Handle type bound

	return &ast.Field{
		Names: []*ast.Ident{
			ast.NewIdent(typeName),
		},
		Type: ast.NewIdent("any"), // TODO: Handle type of generics here
	}
}

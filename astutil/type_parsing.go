package astutil

import (
	"fmt"
	"go/ast"

	sitter "github.com/smacker/go-tree-sitter"
)

func ParseType(node *sitter.Node, source []byte) ast.Expr {
	switch node.Type() {
	case "integral_type":
		switch node.Child(0).Type() {
		case "int":
			return &ast.Ident{Name: "int32"}
		case "short":
			return &ast.Ident{Name: "int16"}
		case "long":
			return &ast.Ident{Name: "int64"}
		case "char":
			return &ast.Ident{Name: "rune"}
		case "byte":
			return &ast.Ident{Name: node.Content(source)}
		}

		panic(fmt.Errorf("Unknown integral type: %v", node.Child(0).Type()))
	case "floating_point_type": // Can be either `float` or `double`
		switch node.Child(0).Type() {
		case "float":
			return &ast.Ident{Name: "float32"}
		case "double":
			return &ast.Ident{Name: "float64"}
		}

		panic(fmt.Errorf("Unknown float type: %v", node.Child(0).Type()))
	case "void_type":
		return &ast.Ident{}
	case "boolean_type":
		return &ast.Ident{Name: "bool"}
	case "generic_type":
		// A generic type is any type that is of the form GenericType<T>
		return &ast.Ident{Name: node.NamedChild(0).Content(source)}
	case "array_type":
		return &ast.ArrayType{Elt: ParseType(node.NamedChild(0), source)}
	case "type_identifier": // Any reference type
		switch node.Content(source) {
		// Special case for strings, because in Go, these are primitive types
		case "String":
			return &ast.Ident{Name: "string"}
		}

		return &ast.StarExpr{
			X: &ast.Ident{Name: node.Content(source)},
		}
	case "scoped_type_identifier":
		// This contains a reference to the type of a nested class
		// Ex: LinkedList.Node
		return &ast.StarExpr{X: &ast.Ident{Name: node.Content(source)}}
	}
	panic("Unknown type to convert: " + node.Type())
}

package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
)

func main() {
	for _, fileName := range os.Args[1:] {
		parser := sitter.NewParser()
		parser.SetLanguage(java.GetLanguage())

		sourceCode, err := os.ReadFile(fileName)
		if err != nil {
			panic(err)
		}
		tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
		if err != nil {
			fmt.Println(err)
		}

		n := tree.RootNode()

		err = printer.Fprint(os.Stdout, token.NewFileSet(), ParseNode(n, sourceCode, "").(ast.Node))
		if err != nil {
			panic(err)
		}
	}
}

func Children(node *sitter.Node) []*sitter.Node {
	count := int(node.NamedChildCount())
	children := make([]*sitter.Node, count)
	for i := 0; i < count; i++ {
		children[i] = node.NamedChild(i)
	}
	return children
}

func Inspect(node *sitter.Node, source []byte) {
	for _, c := range Children(node) {
		fmt.Println(c, c.Content(source))
	}
}

func ParseNode(node *sitter.Node, source []byte, className string) interface{} {
	switch node.Type() {
	// A program contains all the source code, in this case, one `class_declaration`
	case "program":
		program := &ast.File{
			Name:  &ast.Ident{Name: "main"},
			Decls: ParseNode(node.NamedChild(0), source, className).([]ast.Decl),
		}
		return program
	// A class declaration contains the name of the class, and the `class_body`
	// that contains the contents of the class
	case "class_declaration":
		// Since `class_body` contains all the methods and fields in the class, we
		// need to return those, along with the generated struct

		// First go through and generate the struct, with all of its fields
		fields := &ast.FieldList{}
		for _, child := range Children(node) {
			if child.Type() == "field_declaration" {
				fields.List = append(fields.List, ParseNode(child, source, className).(*ast.Field))
			}
		}

		structName := node.NamedChild(0).Content(source) // Class's name is the first node
		className = structName
		decls := []ast.Decl{GenStruct(structName, fields)}

		// Join the generated struct with all the other decls
		return append(decls, ParseNode(node.NamedChild(1), source, className).([]ast.Decl)...)
	case "class_body":
		decls := []ast.Decl{}
		for _, item := range Children(node) {
			if item.Type() != "field_declaration" { // Field declarations have already been handled
				decls = append(decls, ParseNode(item, source, className).(ast.Decl))
			}
		}
		return decls
	case "constructor_declaration":
		mods := node.NamedChild(0)
		_ = mods

		return &ast.FuncDecl{
			Name: &ast.Ident{Name: "New" + ParseNode(node.NamedChild(1), source, className).(*ast.Ident).Name},
			Type: &ast.FuncType{
				Params:  ParseNode(node.NamedChild(2), source, className).(*ast.FieldList),
				Results: &ast.FieldList{},
			},
			Body: ParseNode(node.NamedChild(3), source, className).(*ast.BlockStmt),
		}
	case "method_declaration":
		mods := node.NamedChild(0)
		_ = mods

		return &ast.FuncDecl{
			Name: ParseNode(node.NamedChild(2), source, className).(*ast.Ident),
			Recv: &ast.FieldList{List: []*ast.Field{
				&ast.Field{
					Names: []*ast.Ident{&ast.Ident{Name: ShortName(className)}},
					Type:  &ast.StarExpr{X: &ast.Ident{Name: className}},
				},
			}},
			Type: &ast.FuncType{
				Params: ParseNode(node.NamedChild(3), source, className).(*ast.FieldList),
				Results: &ast.FieldList{List: []*ast.Field{
					&ast.Field{Type: ParseNode(node.NamedChild(1), source, className).(*ast.Ident)},
				}},
			},
			Body: ParseNode(node.NamedChild(4), source, className).(*ast.BlockStmt),
		}
	case "constructor_body", "block":
		body := &ast.BlockStmt{}
		for _, line := range Children(node) {
			body.List = append(body.List, ParseNode(line, source, className).(ast.Stmt))
		}
		return body
	case "expression_statement":
		stmt := ParseNode(node.NamedChild(0), source, className)
		// If the result is already a statement, don't wrap it in a `ExprStmt`
		if s, ok := stmt.(ast.Stmt); ok {
			return s
		}
		return &ast.ExprStmt{X: ParseNode(node.NamedChild(0), source, className).(ast.Expr)}
	case "return_statement":
		return &ast.ReturnStmt{Results: []ast.Expr{ParseNode(node.NamedChild(0), source, className).(ast.Expr)}}
	case "assignment_expression":
		names := []ast.Expr{}
		values := []ast.Expr{}
		for i := 0; i < int(node.NamedChildCount())-1; i++ {
			names = append(names, ParseNode(node.NamedChild(i), source, className).(ast.Expr))
			values = append(values, ParseNode(node.NamedChild(i+1), source, className).(ast.Expr))
		}
		return &ast.AssignStmt{Lhs: names, Tok: token.ASSIGN, Rhs: values}
	case "field_access":
		return &ast.SelectorExpr{
			X:   ParseNode(node.NamedChild(0), source, className).(ast.Expr),
			Sel: ParseNode(node.NamedChild(1), source, className).(*ast.Ident),
		}
	case "formal_parameters":
		params := &ast.FieldList{}
		for _, param := range Children(node) {
			params.List = append(params.List, ParseNode(param, source, className).(*ast.Field))
		}
		return params
	case "formal_parameter":
		return &ast.Field{
			Names: []*ast.Ident{ParseNode(node.NamedChild(1), source, className).(*ast.Ident)},
			Type:  ParseNode(node.NamedChild(0), source, className).(ast.Expr),
		}
	case "this":
		return &ast.Ident{Name: ShortName(className)}
	case "identifier":
		return &ast.Ident{Name: node.Content(source)}
	case "integral_type":
		return &ast.Ident{Name: node.Content(source)}
	default:
		panic(fmt.Sprintf("Unknown node type: %v", node.Type()))
	}
	return nil
}

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

func UnnamedChildren(node *sitter.Node) []*sitter.Node {
	count := int(node.ChildCount())
	children := make([]*sitter.Node, count)
	for i := 0; i < count; i++ {
		children[i] = node.Child(i)
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
			Name: &ast.Ident{Name: "main"},
		}

		for _, c := range Children(node) {
			switch c.Type() {
			case "class_declaration":
				program.Decls = ParseNode(c, source, className).([]ast.Decl)
			case "import_declaration":
				program.Imports = append(program.Imports, ParseNode(c, source, className).(*ast.ImportSpec))
			}
		}
		return program
	// A class declaration contains the name of the class, and the `class_body`
	// that contains the contents of the class
	case "class_declaration":
		// Since `class_body` contains all the methods and fields in the class, we
		// need to return those, along with the generated struct

		var structName string

		var structDecls []ast.Decl

		// First go through and generate the struct, with all of its fields
		fields := &ast.FieldList{}
		for _, child := range Children(node) {
			switch child.Type() {
			case "field_declaration":
				fields.List = append(fields.List, ParseNode(child, source, className).(*ast.Field))
			case "identifier":
				structName = child.Content(source)
			case "class_body":
				structDecls = ParseNode(child, source, className).([]ast.Decl)
			}
		}

		className = structName
		decls := []ast.Decl{GenStruct(structName, fields)}

		// Join the generated struct with all the other decls
		return append(decls, structDecls...)
	case "import_declaration":
		return &ast.ImportSpec{Name: ParseNode(node.NamedChild(0), source, className).(*ast.Ident)}
	case "scoped_identifier":
		return ParseNode(node.NamedChild(0), source, className).(*ast.Ident)
	case "class_body":
		decls := []ast.Decl{}
		for _, item := range Children(node) {
			if item.Type() != "field_declaration" { // Field declarations have already been handled
				// A class declaration will return a list of all the declarations within
				// it, not just a single declaration
				if item.Type() == "class_declaration" {
					decls = append(decls, ParseNode(item, source, className).([]ast.Decl)...)
				} else {
					if item.Type() != "comment" {
						decls = append(decls, ParseNode(item, source, className).(ast.Decl))
					}
				}
			}
		}
		return decls
	case "constructor_declaration":
		mods := node.NamedChild(0)
		_ = mods

		var body *ast.BlockStmt
		var name *ast.Ident
		var params *ast.FieldList

		for _, c := range Children(node) {
			switch c.Type() {
			case "identifier":
				name = ParseNode(c, source, className).(*ast.Ident)
			case "formal_parameters":
				params = ParseNode(c, source, className).(*ast.FieldList)
			case "constructor_body":
				body = ParseNode(c, source, className).(*ast.BlockStmt)
			}
		}

		// Create the object to construct in the constructor
		body.List = append([]ast.Stmt{&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: ShortName(className)}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.Ident{Name: "new"}, Args: []ast.Expr{&ast.Ident{Name: className}}}},
		}}, body.List...)
		// Return the created object
		body.List = append(body.List, &ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: ShortName(className)}}})

		return &ast.FuncDecl{
			Name: &ast.Ident{Name: "New" + name.Name},
			Type: &ast.FuncType{
				Params:  params,
				Results: &ast.FieldList{},
			},
			Body: body,
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
	case "local_variable_declaration":
		// Ignore the name of the type being declared, because we are going to
		// infer that when the variable gets assigned
		return ParseNode(node.NamedChild(1), source, className).(ast.Stmt)
	case "variable_declarator":
		var names, values []ast.Expr
		for ind := 0; ind < int(node.NamedChildCount())-1; ind++ {
			names = append(names, ParseNode(node.NamedChild(ind), source, className).(ast.Expr))
			values = append(values, ParseNode(node.NamedChild(ind+1), source, className).(ast.Expr))
		}
		return &ast.AssignStmt{Lhs: names, Tok: token.DEFINE, Rhs: values}
	case "constructor_body", "block":
		body := &ast.BlockStmt{}
		for _, line := range Children(node) {
			if line.Type() == "comment" {
				continue
			}
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
	case "throw_statement":
		return &ast.ExprStmt{X: &ast.CallExpr{
			Fun:  &ast.Ident{Name: "panic"},
			Args: []ast.Expr{ParseNode(node.NamedChild(0), source, className).(ast.Expr)},
		}}
	case "if_statement":
		var cond ast.Expr
		var body *ast.BlockStmt
		var elseStmt ast.Stmt

		for _, c := range Children(node) {
			switch c.Type() {
			case "parenthesied_expression":
				cond = ParseNode(c, source, className).(ast.Expr)
			case "block": // First block is the `if`, second is the `else`
				if body == nil {
					body = ParseNode(c, source, className).(*ast.BlockStmt)
				} else {
					elseStmt = ParseNode(c, source, className).(*ast.BlockStmt)
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
			Init: ParseNode(node.NamedChild(0), source, className).(ast.Stmt),
			Cond: ParseNode(node.NamedChild(1), source, className).(ast.Expr),
			Post: ParseNode(node.NamedChild(2), source, className).(ast.Stmt),
			Body: ParseNode(node.NamedChild(3), source, className).(*ast.BlockStmt),
		}
	case "assignment_expression":
		names := []ast.Expr{}
		values := []ast.Expr{}
		for i := 0; i < int(node.NamedChildCount())-1; i++ {
			names = append(names, ParseNode(node.NamedChild(i), source, className).(ast.Expr))
			values = append(values, ParseNode(node.NamedChild(i+1), source, className).(ast.Expr))
		}
		return &ast.AssignStmt{Lhs: names, Tok: token.ASSIGN, Rhs: values}
	case "update_expression":
		// The token is not a named node, so we need to access that specifically
		return &ast.IncDecStmt{
			Tok: StringToToken(node.Child(1).Content(source)),
			X:   ParseNode(node.Child(0), source, className).(ast.Expr),
		}
	case "object_creation_expression":
		return &ast.CallExpr{
			Fun:  ParseNode(node.NamedChild(0), source, className).(ast.Expr),
			Args: ParseNode(node.NamedChild(1), source, className).([]ast.Expr),
		}
	case "array_creation_expression":
		// This contains the array type, and then a dimension(s)? expr
		return &ast.ExprStmt{X: &ast.CallExpr{
			Fun: &ast.Ident{Name: "make"},
			Args: []ast.Expr{
				&ast.ArrayType{Elt: ParseNode(node.NamedChild(0), source, className).(ast.Expr)},
				ParseNode(node.NamedChild(1), source, className).(ast.Expr),
			},
		}}
	case "binary_expression":
		return &ast.BinaryExpr{
			X:  ParseNode(node.Child(0), source, className).(ast.Expr),
			Op: StringToToken(node.Child(1).Content(source)),
			Y:  ParseNode(node.Child(2), source, className).(ast.Expr),
		}
	case "field_access":
		return &ast.SelectorExpr{
			X:   ParseNode(node.NamedChild(0), source, className).(ast.Expr),
			Sel: ParseNode(node.NamedChild(1), source, className).(*ast.Ident),
		}
	case "method_invocation":
		return &ast.SelectorExpr{
			X: &ast.CallExpr{
				Fun:  ParseNode(node.NamedChild(1), source, className).(ast.Expr),
				Args: ParseNode(node.NamedChild(2), source, className).([]ast.Expr),
			},
			Sel: ParseNode(node.NamedChild(0), source, className).(*ast.Ident),
		}
	case "argument_list":
		args := []ast.Expr{}
		for _, c := range Children(node) {
			args = append(args, ParseNode(c, source, className).(ast.Expr))
		}
		return args
	case "array_access":
		return &ast.IndexExpr{
			X:     ParseNode(node.NamedChild(0), source, className).(ast.Expr),
			Index: ParseNode(node.NamedChild(1), source, className).(ast.Expr),
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
	case "void_type":
		return &ast.Ident{}
	case "array_type":
		return &ast.ArrayType{Elt: ParseNode(node.NamedChild(0), source, className).(ast.Expr)}
	case "type_identifier":
		return &ast.Ident{Name: node.Content(source)}
	case "null_literal":
		return &ast.Ident{Name: "nil"}
	case "decimal_integer_literal":
		return &ast.Ident{Name: node.Content(source)}
	case "string_literal":
		return &ast.Ident{Name: node.Content(source)}
	case "comment": // Ignore comments
		return nil
	default:
		panic(fmt.Sprintf("Unknown node type: %v", node.Type()))
	}
	return nil
}

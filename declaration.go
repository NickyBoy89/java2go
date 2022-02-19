package main

import (
	"fmt"
	"go/ast"
	"go/token"

	sitter "github.com/smacker/go-tree-sitter"
)

// ParseDecls represents any type that returns a list of top-level declarations,
// this is any class, interface, or enum declaration, and the function panics
// if an unknown node type is passed into it
func ParseDecls(node *sitter.Node, source []byte, ctx Ctx) []ast.Decl {
	if decls := TryParseDecls(node, source, ctx); decls != nil {
		return decls
	}
	panic(fmt.Errorf("Unknown type to parse for decls: %v", node.Type()))
}

// TryParseDecls is the underlying function for ParseDecls, although it returns
// `nil` on an unknown node, allowing for type testing to be done
func TryParseDecls(node *sitter.Node, source []byte, ctx Ctx) []ast.Decl {
	switch node.Type() {
	case "class_declaration":
		// A class declaration contains the name of the class, and the `class_body`
		// that contains the contents of the class

		// Since `class_body` contains all the methods and fields in the class, we
		// need to return those, along with the generated struct

		// Find the class's name first, to name everything
		for _, c := range Children(node) {
			if c.Type() == "identifier" {
				ctx.className = c.Content(source)
			}
		}

		var structDecls []ast.Decl

		// First go through and generate the struct, with all of its fields
		fields := &ast.FieldList{}
		for _, c := range Children(node) {
			switch c.Type() {
			case "class_body":
				structDecls = ParseDecls(c, source, ctx)
				for _, classChild := range Children(c) {
					if classChild.Type() == "field_declaration" {
						fields.List = append(fields.List, ParseNode(classChild, source, ctx).(*ast.Field))
					}
				}
			}
		}

		decls := []ast.Decl{GenStruct(ctx.className, fields)}

		// Join the generated struct with all the other decls
		return append(decls, structDecls...)
	case "class_body":
		decls := []ast.Decl{}
		for _, item := range Children(node) {
			// Skip all the field declarations in the class body, because they
			// have already been handled, as well as the comments
			if item.Type() == "field_declaration" || item.Type() == "comment" {
				continue
			}

			// Parsing a nested class will return a list of decls
			if declList := TryParseDecls(item, source, ctx); declList != nil {
				decls = append(decls, declList...)
			} else {
				// Otherwise, treat it as a decl
				decls = append(decls, ParseDecl(item, source, ctx))
			}
		}
		return decls
	case "interface_declaration":
		//modifiers := ParseNode(node.NamedChild(0), source, ctx)

		ctx.className = node.NamedChild(1).Content(source)

		// NOTE: Fix this to correctly generate an interface
		return []ast.Decl{}
	case "enum_declaration":
		// An enum is treated as both a struct, and a list of values that define
		// the states that the enum can be in

		//modifiers := ParseNode(node.NamedChild(0), source, ctx)

		ctx.className = node.NamedChild(1).Content(source)

		for _, item := range Children(node.NamedChild(2)) {
			switch item.Type() {
			case "enum_body_declarations":
				for _, bodyDecl := range Children(item) {
					_ = bodyDecl
				}
			}
		}

		// NOTE: Fix this to handle an interface correctly
		//decls := []ast.Decl{GenStruct(ctx.className, fields)}
		return []ast.Decl{}
	}
	return nil
}

// ParseDecl handles anything that is declared within a source file, such as a
// method, function, etc...
func ParseDecl(node *sitter.Node, source []byte, ctx Ctx) ast.Decl {
	if decl := TryParseDecl(node, source, ctx); decl != nil {
		return decl
	}
	panic(fmt.Errorf("Unknown node for declaration: %v", node.Type()))
}

// TryParseDecls is the underlying function for ParseDecl, although it returns
// `nil` when an unknown node is passed in
func TryParseDecl(node *sitter.Node, source []byte, ctx Ctx) ast.Decl {
	switch node.Type() {
	case "constructor_declaration":
		var body *ast.BlockStmt
		var name *ast.Ident
		var params *ast.FieldList

		for _, c := range Children(node) {
			switch c.Type() {
			case "identifier":
				name = ParseExpr(c, source, ctx).(*ast.Ident)
			case "formal_parameters":
				params = ParseNode(c, source, ctx).(*ast.FieldList)
			case "constructor_body":
				body = ParseStmt(c, source, ctx).(*ast.BlockStmt)
			}
		}

		// Create the object to construct in the constructor
		body.List = append([]ast.Stmt{&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: ShortName(ctx.className)}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.Ident{Name: "new"}, Args: []ast.Expr{&ast.Ident{Name: ctx.className}}}},
		}}, body.List...)
		// Return the created object
		body.List = append(body.List, &ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: ShortName(ctx.className)}}})

		return &ast.FuncDecl{
			Name: &ast.Ident{Name: "New" + name.Name},
			Type: &ast.FuncType{
				Params: params,
				Results: &ast.FieldList{List: []*ast.Field{&ast.Field{
					Type: &ast.StarExpr{
						X: name,
					},
				}}},
			},
			Body: body,
		}
	case "method_declaration":
		var public, static bool

		// The return type comes as the second node, after the modifiers
		// however, if the method is generic, this gets pushed down one
		returnTypeIndex := 1
		if node.NamedChild(1).Type() == "type_parameters" {
			returnTypeIndex++
		}

		returnType := ParseExpr(node.NamedChild(returnTypeIndex), source, ctx)

		var methodName *ast.Ident

		var params *ast.FieldList

		for _, c := range Children(node) {
			switch c.Type() {
			case "modifiers":
				for _, mod := range UnnamedChildren(c) {
					switch mod.Type() {
					case "public":
						public = true
					case "static":
						static = true
					case "abstract":
						// TODO: Handle abstract methods correctly
						return &ast.BadDecl{}
					}
				}
			case "type_parameters": // For generic types
			case "formal_parameters":
				params = ParseNode(c, source, ctx).(*ast.FieldList)
			case "identifier":
				if returnType == nil {
					continue
				}
				// The next two identifiers determine the return type and name of the method
				if public {
					methodName = CapitalizeIdent(ParseExpr(c, source, ctx).(*ast.Ident))
				} else {
					methodName = LowercaseIdent(ParseExpr(c, source, ctx).(*ast.Ident))
				}
			}
		}

		methodRecv := &ast.FieldList{List: []*ast.Field{
			&ast.Field{
				Names: []*ast.Ident{&ast.Ident{Name: ShortName(ctx.className)}},
				Type:  &ast.StarExpr{X: &ast.Ident{Name: ctx.className}},
			},
		}}

		if static {
			methodRecv = nil
		}

		// If the methodName is nil, then the printer will panic
		if methodName == nil {
			panic("Method's name is nil")
		}

		return &ast.FuncDecl{
			Name: methodName,
			Recv: methodRecv,
			Type: &ast.FuncType{
				Params: params,
				Results: &ast.FieldList{List: []*ast.Field{
					&ast.Field{Type: returnType},
				}},
			},
			Body: ParseStmt(node.NamedChild(int(node.NamedChildCount()-1)), source, ctx).(*ast.BlockStmt),
		}
	case "static_initializer":
		// A block of `static`, which is run before the main function
		return &ast.FuncDecl{
			Name: &ast.Ident{Name: "init"},
			Type: &ast.FuncType{
				Params: &ast.FieldList{List: []*ast.Field{}},
			},
			Body: ParseStmt(node.NamedChild(0), source, ctx).(*ast.BlockStmt),
		}
	}
	return nil
}

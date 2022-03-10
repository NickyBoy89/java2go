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
		// A class declaration contains:
		// * The `modifiers` of the class
		// * An `identifier` for the name of the class
		// * `type_parameters` if the class is a generic class
		// * The `class_body` for the content of the class

		var publicClass bool

		// Static fields are declared as global variables
		globalVariables := &ast.GenDecl{Tok: token.VAR}

		// All the declarations in the class (functions, methods, etc...)
		var structDecls []ast.Decl

		// All generic types parameters that are used in the class
		var genericTypes []ast.Decl

		fields := &ast.FieldList{}

		for _, c := range Children(node) {
			switch c.Type() {
			// Modifiers for the class
			case "modifiers":
				for _, modifier := range UnnamedChildren(c) {
					switch modifier.Type() {
					case "public":
						publicClass = true
					}
				}
			// The class's name
			case "identifier":
				if publicClass {
					ctx.className = ToPublic(c.Content(source))
				} else {
					ctx.className = ToPrivate(c.Content(source))
				}
			// If the class is generic, contains the type parameters
			case "type_parameters":
				// Generate definitions for all the generic types
				genericTypes = ParseDecls(c, source, ctx)
			// The body of the class
			case "class_body":
				// Parse all the declarations in the class's body
				structDecls = ParseDecls(c, source, ctx)

				// Go through the class and extract the fields
				for _, classDecl := range Children(c) {
					if classDecl.Type() == "field_declaration" {
						var public, static bool

						if classDecl.NamedChild(0).Type() == "modifiers" {
							for _, modifier := range UnnamedChildren(classDecl.NamedChild(0)) {
								switch modifier.Type() {
								case "public":
									public = true
								case "static":
									static = true
								}
							}
						}

						field := ParseNode(classDecl, source, ctx).(*ast.Field)
						if public {
							field.Names = []*ast.Ident{CapitalizeIdent(field.Names[0])}
						} else {
							field.Names = []*ast.Ident{LowercaseIdent(field.Names[0])}
						}

						// Static fields are global variables
						if static {
							globalVariables.Specs = append(globalVariables.Specs, &ast.ValueSpec{
								Names: field.Names,
								Type:  field.Type,
							})
						} else {
							fields.List = append(fields.List, field)
						}
					}
				}
			}
		}

		declarations := []ast.Decl{}

		// Add the global variables first
		if len(globalVariables.Specs) > 0 {
			declarations = append(declarations, globalVariables)
		}

		// Generic type declarations
		declarations = append(declarations, genericTypes...)

		// The struct for the class
		declarations = append(declarations, GenStruct(ctx.className, fields))

		// Add in all the other declarations for the class
		declarations = append(declarations, structDecls...)

		return declarations
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
	case "type_parameters":
		var declarations []ast.Decl

		// A list of generic type parameters
		for _, param := range Children(node) {
			switch param.Type() {
			case "type_parameter":
				declarations = append(declarations, GenTypeInterface(param.NamedChild(0).Content(source), []string{"any"}))
			}
		}

		return declarations
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

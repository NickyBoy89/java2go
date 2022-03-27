package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

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

						var static bool

						comments := []*ast.Comment{}

						if classDecl.NamedChild(0).Type() == "modifiers" {
							for _, modifier := range UnnamedChildren(classDecl.NamedChild(0)) {
								switch modifier.Type() {
								case "static":
									static = true
								case "marker_annotation", "annotation":
									comments = append(comments, &ast.Comment{Text: "//" + modifier.Content(source)})
									if _, in := excludedAnnotations[modifier.Content(source)]; in {
										// Skip this field if there is an ignored annotation
										continue
									}
								}
							}
						}

						// The field can either be a `ValueSpec` or a `Field`, based on whether
						// the field is being set to a value
						fieldDecl := ParseNode(classDecl, source, ctx)

						field, isField := fieldDecl.(*ast.Field)

						if len(comments) > 0 && isField {
							field.Doc = &ast.CommentGroup{List: comments}
						}

						// Static fields are global variables
						if static {
							if !isField {
								globalVariables.Specs = append(globalVariables.Specs, fieldDecl.(*ast.ValueSpec))
							} else {
								globalVariables.Specs = append(globalVariables.Specs, &ast.ValueSpec{
									Names: field.Names,
									Type:  field.Type,
								})
							}
						} else {
							if !isField {
								// NOTE: This silently throws away the initial values of some
								// variables in the class's declaration
								valueField := fieldDecl.(*ast.ValueSpec)
								fields.List = append(fields.List, &ast.Field{Names: valueField.Names, Type: valueField.Type})
							} else {
								fields.List = append(fields.List, field)
							}
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
		// NOTE: This will be generated regardless of anything depending on it
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
				decl := ParseDecl(item, source, ctx)
				// Only add good declarations
				if _, bad := decl.(*ast.BadDecl); !bad {
					decls = append(decls, decl)
				}
			}
		}
		return decls
	case "interface_body":
		methods := &ast.FieldList{}

		for _, c := range Children(node) {
			if c.Type() == "method_declaration" {
				parsedMethod := ParseNode(c, source, ctx).(*ast.Field)
				// If the method was ignored with an annotation, it will return a blank
				// field, so ignore that
				if parsedMethod.Type != nil {
					methods.List = append(methods.List, parsedMethod)
				}
			}
		}

		return []ast.Decl{GenInterface(ctx.className, methods)}
	case "interface_declaration":
		decls := []ast.Decl{}

		for _, c := range Children(node) {
			switch c.Type() {
			case "modifiers":
			case "identifier":
				ctx.className = c.Content(source)
			case "interface_body":
				decls = ParseDecls(c, source, ctx)
			}
		}

		// TODO: Fix this to correctly generate an interface
		return decls
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

		// TODO: Fix this to handle an enum correctly
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

		// Store the annotations as comments on the method
		comments := []*ast.Comment{}

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
					case "marker_annotation", "annotation":
						comments = append(comments, &ast.Comment{Text: "//" + mod.Content(source)})
						// If the annotation was on the list of ignored annotations, don't
						// parse the method
						if _, in := excludedAnnotations[mod.Content(source)]; in {
							return &ast.BadDecl{}
						}
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

		var methodRecv *ast.FieldList

		// If the method is not static, define it as a struct's method
		if !static {
			methodRecv = &ast.FieldList{List: []*ast.Field{
				&ast.Field{
					Names: []*ast.Ident{&ast.Ident{Name: ShortName(ctx.className)}},
					Type:  &ast.StarExpr{X: &ast.Ident{Name: ctx.className}},
				},
			}}
		}

		// If the methodName is nil, then the printer will panic
		if methodName == nil {
			panic("Method's name is nil")
		}

		method := &ast.FuncDecl{
			Doc:  &ast.CommentGroup{List: comments},
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

		// Special case for the main method, since this should always be lowercase,
		// and per java rules, have an array of args defined with it
		if strings.ToLower(methodName.Name) == "main" {
			methodName.Name = "main"
			// Remove all of its parameters
			method.Type.Params = nil
			// Add a new variable for the args
			// args := os.Args
			method.Body.List = append([]ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{&ast.Ident{Name: "args"}},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.SelectorExpr{
							X:   &ast.Ident{Name: "os"},
							Sel: &ast.Ident{Name: "Args"},
						},
					},
				},
			}, method.Body.List...)
		}

		return method
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

package main

import (
	"go/ast"
	"go/token"

	sitter "github.com/smacker/go-tree-sitter"
)

// ParseDecls represents any type that returns a list of top-level declarations,
// this is any class, interface, or enum declaration
func ParseDecls(node *sitter.Node, source []byte, ctx Ctx) []ast.Decl {
	switch node.Type() {
	case "class_declaration":
		// TODO: Currently ignores implements and extends with the following tags:
		//"superclass"
		//"interfaces"

		// All the declarations for the class
		declarations := []ast.Decl{}

		// Global variables
		globalVariables := &ast.GenDecl{Tok: token.VAR}

		// Other declarations
		fields := &ast.FieldList{}

		def := ctx.classScope.FindClass(node.ChildByFieldName("name").Content(source))

		ctx.className = def.Name()

		// First, look through the class's body for field declarations
		for _, child := range Children(node.ChildByFieldName("body")) {
			if child.Type() == "field_declaration" {

				var publicField, staticField bool

				comments := []*ast.Comment{}

				// Handle any modifiers that the field might have
				if child.NamedChild(0).Type() == "modifiers" {
					for _, modifier := range UnnamedChildren(child.NamedChild(0)) {
						switch modifier.Type() {
						case "static":
							staticField = true
						case "public":
							publicField = true
						case "marker_annotation", "annotation":
							comments = append(comments, &ast.Comment{Text: "//" + modifier.Content(source)})
							if _, in := excludedAnnotations[modifier.Content(source)]; in {
								// Skip this field if there is an ignored annotation
								continue
							}
						}
					}
				}

				// Parse the field declaration
				// The field can either be a `Field`, or a `ValueSpec` if it was assigned to a value
				field := ParseNode(child, source, ctx)

				if valueField, hasValue := field.(*ast.ValueSpec); hasValue {
					if len(comments) > 0 {
						valueField.Doc = &ast.CommentGroup{List: comments}
					}

					if staticField {
						// Add the name of the current class to scope the variable to the current class
						valueField.Names[0].Name = ctx.className + valueField.Names[0].Name

						if publicField {
							valueField.Names[0] = CapitalizeIdent(valueField.Names[0])
						} else {
							valueField.Names[0] = LowercaseIdent(valueField.Names[0])
						}

						globalVariables.Specs = append(globalVariables.Specs, valueField)
					} else {
						// TODO: If a variable is not static and it is initialized to
						// a value, the value is thrown away
						fields.List = append(fields.List, &ast.Field{Names: valueField.Names, Type: valueField.Type})
					}
				} else {
					if len(comments) > 0 {
						field.(*ast.Field).Doc = &ast.CommentGroup{List: comments}
					}

					if staticField {
						// Add the name of the current class to scope the variable to the current class
						field.(*ast.Field).Names[0].Name = ctx.className + field.(*ast.Field).Names[0].Name

						if publicField {
							field.(*ast.Field).Names[0] = CapitalizeIdent(field.(*ast.Field).Names[0])
						} else {
							field.(*ast.Field).Names[0] = LowercaseIdent(field.(*ast.Field).Names[0])
						}

						globalVariables.Specs = append(globalVariables.Specs, &ast.ValueSpec{Names: field.(*ast.Field).Names, Type: field.(*ast.Field).Type})
					} else {
						fields.List = append(fields.List, field.(*ast.Field))
					}
				}
			}
		}

		// Add everything into the declarations

		if len(globalVariables.Specs) > 0 {
			declarations = append(declarations, globalVariables)
		}

		if node.ChildByFieldName("type_parameters") != nil {
			declarations = append(declarations, ParseDecls(node.ChildByFieldName("type_parameters"), source, ctx)...)
		}

		declarations = append(declarations, GenStruct(ctx.className, fields))

		declarations = append(declarations, ParseDecls(node.ChildByFieldName("body"), source, ctx)...)

		return declarations
	case "class_body":
		decls := []ast.Decl{}
		var child *sitter.Node
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child = node.NamedChild(i)
			switch child.Type() {
			// Skip fields and comments
			case "field_declaration", "comment":
			case "constructor_declaration", "method_declaration", "static_initializer":
				d := ParseDecl(child, source, ctx)
				if _, bad := d.(*ast.BadDecl); !bad {
					decls = append(decls, d)
				}
			case "class_declaration", "interface_declaration", "enum_declaration":
				decls = append(decls, ParseDecls(child, source, ctx)...)
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
	panic("Unknown type to parse for decls: " + node.Type())
}

// ParseDecl parses a top-level declaration within a source file, including
// but not limited to fields and methods
func ParseDecl(node *sitter.Node, source []byte, ctx Ctx) ast.Decl {
	switch node.Type() {
	case "constructor_declaration":
		body := ParseStmt(node.ChildByFieldName("body"), source, ctx).(*ast.BlockStmt)

		body.List = append([]ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: ShortName(ctx.className)}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.Ident{Name: "new"}, Args: []ast.Expr{&ast.Ident{Name: ctx.className}}}},
			},
		}, body.List...)

		body.List = append(body.List, &ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: ShortName(ctx.className)}}})

		def := ctx.classScope.FindMethod(ParseExpr(node.ChildByFieldName("name"), source, ctx).(*ast.Ident).Name)

		return &ast.FuncDecl{
			Name: &ast.Ident{Name: def.Name()},
			Type: &ast.FuncType{
				Params: ParseNode(node.ChildByFieldName("parameters"), source, ctx).(*ast.FieldList),
				Results: &ast.FieldList{List: []*ast.Field{&ast.Field{
					Type: &ast.StarExpr{
						X: &ast.Ident{Name: def.Type()},
					},
				}}},
			},
			Body: body,
		}
	case "method_declaration":
		var static bool

		// Store the annotations as comments on the method
		comments := []*ast.Comment{}

		if node.NamedChild(0).Type() == "modifiers" {
			for _, modifier := range UnnamedChildren(node.NamedChild(0)) {
				switch modifier.Type() {
				case "static":
					static = true
				case "abstract":
					// TODO: Handle abstract methods correctly
					return &ast.BadDecl{}
				case "marker_annotation", "annotation":
					comments = append(comments, &ast.Comment{Text: "//" + modifier.Content(source)})
					// If the annotation was on the list of ignored annotations, don't
					// parse the method
					if _, in := excludedAnnotations[modifier.Content(source)]; in {
						return &ast.BadDecl{}
					}
				}
			}
		}

		// If a function is non-static, it has a method receiver
		var receiver *ast.FieldList
		if !static {
			receiver = &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{&ast.Ident{Name: ShortName(ctx.className)}},
						Type:  &ast.StarExpr{X: &ast.Ident{Name: ctx.className}},
					},
				},
			}
		}

		name := ParseExpr(node.ChildByFieldName("name"), source, ctx).(*ast.Ident)

		def := ctx.classScope.FindMethod(name.Name)

		body := ParseStmt(node.ChildByFieldName("body"), source, ctx).(*ast.BlockStmt)
		params := ParseNode(node.ChildByFieldName("parameters"), source, ctx).(*ast.FieldList)

		// Special case for the main method, because in Java, this method has the
		// command line args passed in as a parameter
		if name.Name == "main" {
			params = nil
			body.List = append([]ast.Stmt{
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
			}, body.List...)
		}

		return &ast.FuncDecl{
			Doc:  &ast.CommentGroup{List: comments},
			Name: &ast.Ident{Name: def.Name()},
			Recv: receiver,
			Type: &ast.FuncType{
				Params: params,
				Results: &ast.FieldList{
					List: []*ast.Field{
						&ast.Field{Type: &ast.Ident{Name: def.Type()}},
					},
				},
			},
			Body: body,
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

	panic("Unknown node type for declaration: " + node.Type())
}

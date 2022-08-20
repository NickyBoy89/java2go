package main

import (
	"go/ast"
	"go/token"

	"github.com/NickyBoy89/java2go/nodeutil"
	"github.com/NickyBoy89/java2go/symbol"
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

		// The declarations and fields for the class
		declarations := []ast.Decl{}
		fields := &ast.FieldList{}

		// Global variables
		globalVariables := &ast.GenDecl{Tok: token.VAR}

		ctx.className = ctx.classScope.FindClass(node.ChildByFieldName("name").Content(source)).Name

		// First, look through the class's body for field declarations
		for _, child := range nodeutil.NamedChildrenOf(node.ChildByFieldName("body")) {
			if child.Type() == "field_declaration" {

				var staticField bool

				comments := []*ast.Comment{}

				// Handle any modifiers that the field might have
				if child.NamedChild(0).Type() == "modifiers" {
					for _, modifier := range nodeutil.UnnamedChildrenOf(child.NamedChild(0)) {
						switch modifier.Type() {
						case "static":
							staticField = true
						case "marker_annotation", "annotation":
							comments = append(comments, &ast.Comment{Text: "//" + modifier.Content(source)})
							if _, in := excludedAnnotations[modifier.Content(source)]; in {
								// Skip this field if there is an ignored annotation
								continue
							}
						}
					}
				}

				// TODO: If a field is initialized to a value, that value is discarded

				field := &ast.Field{}
				if len(comments) > 0 {
					field.Doc = &ast.CommentGroup{List: comments}
				}

				fieldDef := ctx.classScope.FindFieldByName(child.ChildByFieldName("declarator").ChildByFieldName("name").Content(source))

				field.Names, field.Type = []*ast.Ident{&ast.Ident{Name: fieldDef.Name}}, &ast.Ident{Name: fieldDef.Type}

				if staticField {
					globalVariables.Specs = append(globalVariables.Specs, &ast.ValueSpec{Names: field.Names, Type: field.Type})
				} else {
					fields.List = append(fields.List, field)
				}
			}
		}

		// Add the global variables
		if len(globalVariables.Specs) > 0 {
			declarations = append(declarations, globalVariables)
		}

		// Add any type paramters defined in the class
		if node.ChildByFieldName("type_parameters") != nil {
			declarations = append(declarations, ParseDecls(node.ChildByFieldName("type_parameters"), source, ctx)...)
		}

		// Add the struct for the class
		declarations = append(declarations, GenStruct(ctx.className, fields))

		// Add all the declarations that appear in the class
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

		for _, c := range nodeutil.NamedChildrenOf(node) {
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
		ctx.className = ctx.classScope.FindClass(node.ChildByFieldName("name").Content(source)).Name

		return ParseDecls(node.ChildByFieldName("body"), source, ctx)
	case "enum_declaration":
		// An enum is treated as both a struct, and a list of values that define
		// the states that the enum can be in

		ctx.className = ctx.classScope.FindClass(node.ChildByFieldName("name").Content(source)).Name

		// TODO: Handle an enum correctly
		//return ParseDecls(node.ChildByFieldName("body"), source, ctx)
		return []ast.Decl{}
	case "type_parameters":
		var declarations []ast.Decl

		// A list of generic type parameters
		for _, param := range nodeutil.NamedChildrenOf(node) {
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
		paramNode := node.ChildByFieldName("parameters")
		parameterTypes := make([]string, paramNode.NamedChildCount())
		for ind := 0; ind < int(paramNode.NamedChildCount()); ind++ {
			if paramNode.NamedChild(ind).Type() == "spread_parameter" {
				parameterTypes[ind] = paramNode.NamedChild(ind).NamedChild(0).Content(source)
			} else {
				parameterTypes[ind] = paramNode.NamedChild(ind).ChildByFieldName("type").Content(source)
			}
		}

		ctx.localScope = ctx.classScope.FindMethodByName(ParseExpr(node.ChildByFieldName("name"), source, ctx).(*ast.Ident).Name, parameterTypes)

		body := ParseStmt(node.ChildByFieldName("body"), source, ctx).(*ast.BlockStmt)

		body.List = append([]ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: ShortName(ctx.className)}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.Ident{Name: "new"}, Args: []ast.Expr{&ast.Ident{Name: ctx.className}}}},
			},
		}, body.List...)

		body.List = append(body.List, &ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: ShortName(ctx.className)}}})

		return &ast.FuncDecl{
			Name: &ast.Ident{Name: ctx.localScope.Name},
			Type: &ast.FuncType{
				Params: ParseNode(node.ChildByFieldName("parameters"), source, ctx).(*ast.FieldList),
				Results: &ast.FieldList{List: []*ast.Field{&ast.Field{
					Type: &ast.Ident{Name: ctx.localScope.Type},
				}}},
			},
			Body: body,
		}
	case "method_declaration":
		var static bool

		// Store the annotations as comments on the method
		comments := []*ast.Comment{}

		if node.NamedChild(0).Type() == "modifiers" {
			for _, modifier := range nodeutil.UnnamedChildrenOf(node.NamedChild(0)) {
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

		paramNode := node.ChildByFieldName("parameters")
		parameterTypes := make([]string, paramNode.NamedChildCount())
		for ind := 0; ind < int(paramNode.NamedChildCount()); ind++ {
			if paramNode.NamedChild(ind).Type() == "spread_parameter" {
				parameterTypes[ind] = paramNode.NamedChild(ind).NamedChild(0).Content(source)
			} else {
				parameterTypes[ind] = paramNode.NamedChild(ind).ChildByFieldName("type").Content(source)
			}
		}

		ctx.localScope = ctx.classScope.FindMethodByName(name.Name, parameterTypes)
		// TODO: This can look up a method of the same name and type in another class, and it will display incorrect results

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
			Name: &ast.Ident{Name: ctx.localScope.Name},
			Recv: receiver,
			Type: &ast.FuncType{
				Params: params,
				Results: &ast.FieldList{
					List: []*ast.Field{
						&ast.Field{Type: &ast.Ident{Name: ctx.localScope.Type}},
					},
				},
			},
			Body: body,
		}
	case "static_initializer":

		ctx.localScope = &symbol.Definition{}

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

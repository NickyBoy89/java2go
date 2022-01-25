package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"unicode"

	sitter "github.com/smacker/go-tree-sitter"
)

// Children gets all named children of a given node
func Children(node *sitter.Node) []*sitter.Node {
	count := int(node.NamedChildCount())
	children := make([]*sitter.Node, count)
	for i := 0; i < count; i++ {
		children[i] = node.NamedChild(i)
	}
	return children
}

// UnnamedChildren gets all the named + unnamed children of a given node
func UnnamedChildren(node *sitter.Node) []*sitter.Node {
	count := int(node.ChildCount())
	children := make([]*sitter.Node, count)
	for i := 0; i < count; i++ {
		children[i] = node.Child(i)
	}
	return children
}

// Inspect is a function for debugging that prints out every named child of a
// given node and the source code for that child
func Inspect(node *sitter.Node, source []byte) {
	for _, c := range Children(node) {
		fmt.Println(c, c.Content(source))
	}
}

// CapitalizeIdent capitalizes the first letter of a `*ast.Ident` to mark the
// result as a public method or field
func CapitalizeIdent(in *ast.Ident) *ast.Ident {
	return &ast.Ident{Name: string(unicode.ToUpper(rune(in.Name[0]))) + in.Name[1:]}
}

// LowercaseIdent lowercases the first letter of a `*ast.Ident` to mark the
// result as a private method or field
func LowercaseIdent(in *ast.Ident) *ast.Ident {
	return &ast.Ident{Name: string(unicode.ToLower(rune(in.Name[0]))) + in.Name[1:]}
}

// A Ctx is passed into the `ParseNode` function and contains any data that is
// needed down-the-line for parsing, such as the class's name
type Ctx struct {
	// Used to generate the names of all the methods, as well as the names
	// of the constructors
	className string
	// Used when generating arrays, because in Java, these are defined as
	// arrType[] varName = {item, item, item}, and no class name data is defined
	arrayType string
}

// Parses a given tree-sitter node and returns the ast representation for it
// if called on the root of a tree-sitter node, it will return the entire
// generated golang ast as a `ast.Node` type
func ParseNode(node *sitter.Node, source []byte, ctx Ctx) interface{} {
	switch node.Type() {
	// A program contains all the source code, in this case, one `class_declaration`
	case "program":
		program := &ast.File{
			Name: &ast.Ident{Name: "main"},
		}

		for _, c := range Children(node) {
			switch c.Type() {
			case "class_declaration":
				program.Decls = ParseNode(c, source, ctx).([]ast.Decl)
			case "import_declaration":
				program.Imports = append(program.Imports, ParseNode(c, source, ctx).(*ast.ImportSpec))
			}
		}
		return program
	// A class declaration contains the name of the class, and the `class_body`
	// that contains the contents of the class
	case "class_declaration":
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
				structDecls = ParseNode(c, source, ctx).([]ast.Decl)
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
	case "field_declaration":
		return &ast.Field{
			Names: []*ast.Ident{
				// This field is a `variable_declarator` which gets parsed out to a
				// full statement, but we only want the identifier for its type
				ParseNode(node.NamedChild(1).NamedChild(0), source, ctx).(*ast.Ident),
			},
			Type: ParseNode(node.NamedChild(0), source, ctx).(ast.Expr),
		}
	case "import_declaration":
		return &ast.ImportSpec{Name: ParseNode(node.NamedChild(0), source, ctx).(*ast.Ident)}
	case "scoped_identifier":
		return ParseNode(node.NamedChild(0), source, ctx).(*ast.Ident)
	case "class_body":
		decls := []ast.Decl{}
		for _, item := range Children(node) {
			if item.Type() != "field_declaration" { // Field declarations have already been handled
				// A class declaration will return a list of all the declarations within
				// it, not just a single declaration
				if item.Type() == "class_declaration" {
					decls = append(decls, ParseNode(item, source, ctx).([]ast.Decl)...)
				} else {
					if item.Type() != "comment" {
						decls = append(decls, ParseNode(item, source, ctx).(ast.Decl))
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
				name = ParseNode(c, source, ctx).(*ast.Ident)
			case "formal_parameters":
				params = ParseNode(c, source, ctx).(*ast.FieldList)
			case "constructor_body":
				body = ParseNode(c, source, ctx).(*ast.BlockStmt)
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
		for _, mod := range UnnamedChildren(node.NamedChild(0)) {
			switch mod.Type() {
			case "public":
				public = true
			case "static":
				static = true
			}
		}

		methodName := ParseNode(node.NamedChild(2), source, ctx).(*ast.Ident)
		if public {
			methodName = CapitalizeIdent(methodName)
		} else {
			methodName = LowercaseIdent(methodName)
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

		return &ast.FuncDecl{
			Name: methodName,
			Recv: methodRecv,
			Type: &ast.FuncType{
				Params: ParseNode(node.NamedChild(3), source, ctx).(*ast.FieldList),
				Results: &ast.FieldList{List: []*ast.Field{
					&ast.Field{Type: ParseNode(node.NamedChild(1), source, ctx).(ast.Expr)},
				}},
			},
			Body: ParseNode(node.NamedChild(4), source, ctx).(*ast.BlockStmt),
		}
	case "local_variable_declaration":
		// Ignore the name of the type being declared, because we are going to
		// infer that when the variable gets assigned
		return ParseNode(node.NamedChild(1), source, ctx).(ast.Stmt)
	case "variable_declarator":
		var names, values []ast.Expr
		for ind := 0; ind < int(node.NamedChildCount())-1; ind++ {
			names = append(names, ParseNode(node.NamedChild(ind), source, ctx).(ast.Expr))
			values = append(values, ParseNode(node.NamedChild(ind+1), source, ctx).(ast.Expr))
		}
		return &ast.AssignStmt{Lhs: names, Tok: token.DEFINE, Rhs: values}
	case "constructor_body", "block":
		body := &ast.BlockStmt{}
		for _, line := range Children(node) {
			if line.Type() == "comment" {
				continue
			}
			body.List = append(body.List, ParseNode(line, source, ctx).(ast.Stmt))
		}
		return body
	case "expression_statement":
		stmt := ParseNode(node.NamedChild(0), source, ctx)
		// If the result is already a statement, don't wrap it in a `ExprStmt`
		if s, ok := stmt.(ast.Stmt); ok {
			return s
		}
		return &ast.ExprStmt{X: ParseNode(node.NamedChild(0), source, ctx).(ast.Expr)}
	case "return_statement":
		return &ast.ReturnStmt{Results: []ast.Expr{ParseNode(node.NamedChild(0), source, ctx).(ast.Expr)}}
	case "throw_statement":
		return &ast.ExprStmt{X: &ast.CallExpr{
			Fun:  &ast.Ident{Name: "panic"},
			Args: []ast.Expr{ParseNode(node.NamedChild(0), source, ctx).(ast.Expr)},
		}}
	case "if_statement":
		var cond ast.Expr
		var body *ast.BlockStmt
		var elseStmt ast.Stmt

		for _, c := range Children(node) {
			switch c.Type() {
			case "parenthesized_expression":
				cond = ParseNode(c, source, ctx).(ast.Expr)
			case "block": // First block is the `if`, second is the `else`
				if body == nil {
					body = ParseNode(c, source, ctx).(*ast.BlockStmt)
				} else {
					elseStmt = ParseNode(c, source, ctx).(*ast.BlockStmt)
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
			Init: ParseNode(node.NamedChild(0), source, ctx).(ast.Stmt),
			Cond: ParseNode(node.NamedChild(1), source, ctx).(ast.Expr),
			Post: ParseNode(node.NamedChild(2), source, ctx).(ast.Stmt),
			Body: ParseNode(node.NamedChild(3), source, ctx).(*ast.BlockStmt),
		}
	case "assignment_expression":
		names := []ast.Expr{}
		values := []ast.Expr{}
		for i := 0; i < int(node.NamedChildCount())-1; i++ {
			names = append(names, ParseNode(node.NamedChild(i), source, ctx).(ast.Expr))
			values = append(values, ParseNode(node.NamedChild(i+1), source, ctx).(ast.Expr))
		}
		return &ast.AssignStmt{Lhs: names, Tok: token.ASSIGN, Rhs: values}
	case "update_expression":
		// The token is not a named node, so we need to access that specifically
		return &ast.IncDecStmt{
			Tok: StringToToken(node.Child(1).Content(source)),
			X:   ParseNode(node.Child(0), source, ctx).(ast.Expr),
		}
	case "object_creation_expression":
		return &ast.CallExpr{
			// All object creations are usually done by calling the constructor
			//function, which is generated as `"New" + className`
			Fun:  &ast.Ident{Name: "New" + ParseNode(node.NamedChild(0), source, ctx).(*ast.StarExpr).X.(*ast.Ident).Name},
			Args: ParseNode(node.NamedChild(1), source, ctx).([]ast.Expr),
		}
	case "array_creation_expression":
		// The type of the array
		arrayType := ParseNode(node.NamedChild(0), source, ctx).(ast.Expr)
		// The dimensions of the array, which Golang only supports defining one at
		// a time with the use of the builtin `make`
		dimensions := []ast.Expr{&ast.ArrayType{Elt: arrayType}}
		for _, c := range Children(node)[1:] {
			if c.Type() == "dimensions_expr" {
				dimensions = append(dimensions, ParseNode(c, source, ctx).(ast.Expr))
			}
		}

		return &ast.CallExpr{
			Fun:  &ast.Ident{Name: "make"},
			Args: dimensions,
		}
	case "dimensions_expr":
		return &ast.Ident{Name: node.NamedChild(0).Content(source)}
	case "binary_expression":
		return &ast.BinaryExpr{
			X:  ParseNode(node.Child(0), source, ctx).(ast.Expr),
			Op: StringToToken(node.Child(1).Content(source)),
			Y:  ParseNode(node.Child(2), source, ctx).(ast.Expr),
		}
	case "unary_expression":
		return &ast.UnaryExpr{
			Op: StringToToken(node.Child(0).Content(source)),
			X:  ParseNode(node.Child(1), source, ctx).(ast.Expr),
		}
	case "parenthesized_expression":
		return &ast.ParenExpr{
			X: ParseNode(node.NamedChild(0), source, ctx).(ast.Expr),
		}
	case "field_access":
		return &ast.SelectorExpr{
			X:   ParseNode(node.NamedChild(0), source, ctx).(ast.Expr),
			Sel: ParseNode(node.NamedChild(1), source, ctx).(*ast.Ident),
		}
	case "method_invocation":
		// Class methods are called with three nodes, the selector, the identifier,
		// and the list of arguments, so that they form the shape
		// `selector.identifier(list of arguments)`
		// Static methods are only called with the identifier and list of args
		// They look like: `identifier(args)`

		switch node.NamedChildCount() {
		case 3: // Invoking a normal class method
			// This is of the form X.Sel(Args)
			return &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ParseNode(node.NamedChild(0), source, ctx).(ast.Expr),
					Sel: ParseNode(node.NamedChild(1), source, ctx).(*ast.Ident),
				},
				Args: ParseNode(node.NamedChild(2), source, ctx).([]ast.Expr),
			}
		case 2: // Invoking a static method
			return &ast.CallExpr{
				// The name is in the wrong place, so use the selector as the name
				Fun:  ParseNode(node.NamedChild(0), source, ctx).(ast.Expr),
				Args: ParseNode(node.NamedChild(1), source, ctx).([]ast.Expr),
			}
		default:
			panic(fmt.Sprintf("Calling method with unknown number of args: %v", node.NamedChildCount()))
		}

	case "argument_list":
		args := []ast.Expr{}
		for _, c := range Children(node) {
			args = append(args, ParseNode(c, source, ctx).(ast.Expr))
		}
		return args
	case "array_access":
		return &ast.IndexExpr{
			X:     ParseNode(node.NamedChild(0), source, ctx).(ast.Expr),
			Index: ParseNode(node.NamedChild(1), source, ctx).(ast.Expr),
		}
	case "array_initializer":
		items := []ast.Expr{}
		for _, c := range Children(node) {
			items = append(items, ParseNode(c, source, ctx).(ast.Expr))
		}
		return &ast.CompositeLit{
			Type: &ast.ArrayType{
				Elt: &ast.Ident{Name: "int"},
			},
			Elts: items,
		}
	case "formal_parameters":
		params := &ast.FieldList{}
		for _, param := range Children(node) {
			params.List = append(params.List, ParseNode(param, source, ctx).(*ast.Field))
		}
		return params
	case "formal_parameter":
		return &ast.Field{
			Names: []*ast.Ident{ParseNode(node.NamedChild(1), source, ctx).(*ast.Ident)},
			Type:  ParseNode(node.NamedChild(0), source, ctx).(ast.Expr),
		}
	case "this":
		return &ast.Ident{Name: ShortName(ctx.className)}
	case "identifier":
		return &ast.Ident{Name: node.Content(source)}
	case "integral_type":
		return &ast.Ident{Name: node.Content(source)}
	case "void_type":
		return &ast.Ident{}
	case "array_type":
		return &ast.ArrayType{Elt: ParseNode(node.NamedChild(0), source, ctx).(ast.Expr)}
	case "type_identifier": // Any reference type
		return &ast.StarExpr{
			X: &ast.Ident{Name: node.Content(source)},
		}
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

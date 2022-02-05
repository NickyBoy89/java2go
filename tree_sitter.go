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

		//decls := []ast.Decl{GenStruct(ctx.className, fields)}
		return []ast.Decl{}
	case "field_declaration":
		var fieldType ast.Expr
		var fieldName *ast.Ident

		for ind, c := range Children(node) {
			switch c.Type() {
			case "modifiers": // Ignore the modifiers for now
				// The variable's type will always follow the modifiers, if they are present
				fieldType = ParseNode(node.NamedChild(ind+1), source, ctx).(ast.Expr)
				// The value will come one after that
				fieldName = ParseNode(node.NamedChild(ind+2).NamedChild(0), source, ctx).(*ast.Ident)
			}
		}

		// If no modifiers were declared, then declare everything with the default
		// offsets
		if fieldType == nil {
			fieldType = ParseNode(node.NamedChild(0), source, ctx).(ast.Expr)
			fieldName = ParseNode(node.NamedChild(1).NamedChild(0), source, ctx).(*ast.Ident)
		}

		return &ast.Field{
			Names: []*ast.Ident{
				// This field is a `variable_declarator` which gets parsed out to a
				// full statement, but we only want the identifier for its type
				fieldName,
			},
			Type: fieldType,
		}
	case "import_declaration":
		return &ast.ImportSpec{Name: ParseNode(node.NamedChild(0), source, ctx).(*ast.Ident)}
	case "scoped_identifier":
		return ParseNode(node.NamedChild(0), source, ctx).(*ast.Ident)
	case "class_body":
		decls := []ast.Decl{}
		for _, item := range Children(node) {
			if item.Type() != "field_declaration" { // Field declarations have already been handled
				decl := ParseNode(item, source, ctx)
				// Skip comments
				if item.Type() == "comment" {
					continue
				}
				// Parsing a nested class will instead return a list of all the decls
				// contained within the class
				if declList, ok := decl.([]ast.Decl); ok {
					decls = append(decls, declList...)
				} else {
					decls = append(decls, decl.(ast.Decl))
				}
			}
		}
		return decls
	case "constructor_declaration":
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

		// The return type comes as the second node, after the modifiers
		// however, if the method is generic, this gets pushed down one
		returnTypeIndex := 1
		if node.NamedChild(1).Type() == "type_parameters" {
			returnTypeIndex++
		}

		returnType := ParseNode(node.NamedChild(returnTypeIndex), source, ctx).(ast.Expr)

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
					methodName = CapitalizeIdent(ParseNode(c, source, ctx).(*ast.Ident))
				} else {
					methodName = LowercaseIdent(ParseNode(c, source, ctx).(*ast.Ident))
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
			Body: ParseNode(node.NamedChild(int(node.NamedChildCount()-1)), source, ctx).(*ast.BlockStmt),
		}
	case "static_initializer":
		return &ast.FuncDecl{
			Name: &ast.Ident{Name: "init"},
			Type: &ast.FuncType{
				Params: &ast.FieldList{List: []*ast.Field{}},
			},
			Body: ParseNode(node.NamedChild(0), source, ctx).(*ast.BlockStmt),
		}
	case "local_variable_declaration":
		// Ignore the name of the type being declared, because we are going to
		// infer that when the variable gets assigned
		if node.NamedChild(0).Type() == "modifiers" {
			return ParseNode(node.NamedChild(2), source, ctx).(ast.Stmt)
		}
		return ParseNode(node.NamedChild(1), source, ctx).(ast.Stmt)
	case "variable_declarator":
		var names, values []ast.Expr

		// If there is only one node, then that node is just a name
		if node.NamedChildCount() == 1 {
			names = append(names, ParseNode(node.NamedChild(0), source, ctx).(ast.Expr))
		}

		// Loop through every pair of name and value
		for ind := 0; ind < int(node.NamedChildCount())-1; ind += 2 {
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
			// Try statements are ignored, so they return a list of statements
			if _, ok := ParseNode(line, source, ctx).([]ast.Stmt); ok {
				body.List = append(body.List, ParseNode(line, source, ctx).([]ast.Stmt)...)
			} else {
				body.List = append(body.List, ParseNode(line, source, ctx).(ast.Stmt))
			}
		}
		return body
	case "switch_block":
		switchBlock := &ast.BlockStmt{}
		var currentCase *ast.CaseClause
		for _, c := range Children(node) {
			switch c.Type() {
			case "switch_label":
				// When a new switch label comes, append it to the switch block
				if currentCase != nil {
					switchBlock.List = append(switchBlock.List, currentCase)
				}
				currentCase = ParseNode(c, source, ctx).(*ast.CaseClause)
			default:
				currentCase.Body = append(currentCase.Body, ParseNode(c, source, ctx).(ast.Stmt))
			}
		}

		return switchBlock
	case "expression_statement":
		stmt := ParseNode(node.NamedChild(0), source, ctx)
		// If the result is already a statement, don't wrap it in a `ExprStmt`
		if s, ok := stmt.(ast.Stmt); ok {
			return s
		} else if s, ok := stmt.([]ast.Stmt); ok { // Return the assignstmts
			return s
		}
		return &ast.ExprStmt{X: stmt.(ast.Expr)}
	case "return_statement":
		if node.NamedChildCount() < 1 {
			return &ast.ReturnStmt{Results: []ast.Expr{}}
		}
		return &ast.ReturnStmt{Results: []ast.Expr{ParseNode(node.NamedChild(0), source, ctx).(ast.Expr)}}
	case "break_statement":
		return &ast.BranchStmt{Tok: token.BREAK}
	case "throw_statement":
		return &ast.ExprStmt{X: &ast.CallExpr{
			Fun:  &ast.Ident{Name: "panic"},
			Args: []ast.Expr{ParseNode(node.NamedChild(0), source, ctx).(ast.Expr)},
		}}
	case "try_statement":
		// We ignore try statements
		return ParseNode(node.NamedChild(0), source, ctx).(*ast.BlockStmt).List
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
	case "while_statement":
		return &ast.ForStmt{
			Cond: ParseNode(node.NamedChild(0), source, ctx).(ast.Expr),
			Body: ParseNode(node.NamedChild(1), source, ctx).(*ast.BlockStmt),
		}
	case "do_statement":
		// A do statement is handled as a blank for loop with the condition
		// inserted as a break condition in the final part of the loop
		body := ParseNode(node.NamedChild(0), source, ctx).(*ast.BlockStmt)

		body.List = append(body.List, &ast.IfStmt{
			Cond: &ast.UnaryExpr{
				X: &ast.ParenExpr{
					X: ParseNode(node.NamedChild(1), source, ctx).(ast.Expr),
				},
			},
			Body: &ast.BlockStmt{List: []ast.Stmt{&ast.BranchStmt{Tok: token.BREAK}}},
		})

		return &ast.ForStmt{
			Body: body,
		}
	case "switch_statement":
		return &ast.SwitchStmt{
			Tag:  ParseNode(node.NamedChild(0), source, ctx).(ast.Expr),
			Body: ParseNode(node.NamedChild(1), source, ctx).(*ast.BlockStmt),
		}
	case "assignment_expression":
		// A simple variable assignment, ex: `name = value`

		// Stores all the assignments if the statement is a multiple-expression
		assignments := []ast.Stmt{}

		names := []ast.Expr{}
		values := []ast.Expr{}
		for i := 0; i < int(node.NamedChildCount())-1; i++ {
			// Rewrite double assignments, e.g. `variable1 = variable2 = 1` to
			// `variable2 = 1`
			// `variable1 = variable2`
			if node.NamedChild(i+1).Type() == "assignment_expression" {
				// If a value is a multiple assignment, add that assignment before the
				// current one, and add the left side of the value to the current line
				otherAssign := ParseNode(node.NamedChild(i+1), source, ctx)

				if otherStmts, ok := otherAssign.([]ast.Stmt); ok {
					assignments = append(assignments, otherStmts...)
				} else {
					assignments = append(assignments, otherAssign.(ast.Stmt))
				}

				// Assign the value to the latest Lhs expression
				assignments = append(assignments, &ast.AssignStmt{
					Lhs: []ast.Expr{ParseNode(node.NamedChild(i), source, ctx).(ast.Expr)},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{assignments[len(assignments)-1].(*ast.AssignStmt).Lhs[0]},
				})
			} else {
				names = append(names, ParseNode(node.NamedChild(i), source, ctx).(ast.Expr))
				values = append(values, ParseNode(node.NamedChild(i+1), source, ctx).(ast.Expr))
			}
		}

		if len(assignments) > 0 {
			return assignments
		}

		return &ast.AssignStmt{Lhs: names, Tok: token.ASSIGN, Rhs: values}
	case "update_expression":
		// If the unnamed token comes first, then this is a pre-increment, such as
		// ++value
		// other than that, if the token comes second, this looks like: value++

		// The post-increment is not supported in go, so instead, this is faked by
		// passing the value through a function
		if node.Child(0).Type() != "identifier" {
			return &ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.Ident{Name: "PostUpdate"},
					Args: []ast.Expr{
						ParseNode(node.Child(1), source, ctx).(ast.Expr),
					},
				},
			}
			panic("Pre-update")
		}
		return &ast.IncDecStmt{
			Tok: StrToToken(node.Child(1).Content(source)),
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
			Op: StrToToken(node.Child(1).Content(source)),
			Y:  ParseNode(node.Child(2), source, ctx).(ast.Expr),
		}
	case "unary_expression":
		return &ast.UnaryExpr{
			Op: StrToToken(node.Child(0).Content(source)),
			X:  ParseNode(node.Child(1), source, ctx).(ast.Expr),
		}
	case "parenthesized_expression":
		return &ast.ParenExpr{
			X: ParseNode(node.NamedChild(0), source, ctx).(ast.Expr),
		}
	case "ternary_expression":
		// Ternary expressions are represented by a built-in function
		// called `ternary`, which takes in the binary expression, and the two
		// return values

		args := []ast.Expr{}
		for _, c := range Children(node) {
			args = append(args, ParseNode(c, source, ctx).(ast.Expr))
		}
		return &ast.CallExpr{
			Fun:  &ast.Ident{Name: "ternary"},
			Args: args,
		}
	case "lambda_expression":
		return &ast.FuncLit{
			Type: &ast.FuncType{
				Params:  ParseNode(node.NamedChild(0), source, ctx).(*ast.FieldList),
				Results: &ast.FieldList{List: []*ast.Field{}},
			},
			Body: ParseNode(node.NamedChild(1), source, ctx).(*ast.BlockStmt),
		}
	case "cast_expression":
		return &ast.TypeAssertExpr{
			X:    ParseNode(node.NamedChild(1), source, ctx).(ast.Expr),
			Type: ParseNode(node.NamedChild(0), source, ctx).(ast.Expr),
		}
	case "switch_label":
		if node.NamedChildCount() > 0 {
			return &ast.CaseClause{
				List: []ast.Expr{ParseNode(node.NamedChild(0), source, ctx).(ast.Expr)},
			}
		}
		return &ast.CaseClause{}
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
	case "explicit_constructor_invocation":
		// This is when a constructor calls another constructor with the use of
		// something such as `this(args...)`
		return &ast.ExprStmt{
			&ast.CallExpr{
				Fun:  &ast.Ident{Name: "New" + ctx.className},
				Args: ParseNode(node.NamedChild(1), source, ctx).([]ast.Expr),
			},
		}
	case "argument_list":
		args := []ast.Expr{}
		for _, c := range Children(node) {
			args = append(args, ParseNode(c, source, ctx).(ast.Expr))
		}
		return args
	case "array_access":
		// For an array access such as `arr[i++]`, which is not valid, we need to
		// call that statement before the current statement
		if _, isIncDec := ParseNode(node.NamedChild(1), source, ctx).(*ast.IncDecStmt); isIncDec {
			return &ast.IndexExpr{
				X: ParseNode(node.NamedChild(0), source, ctx).(ast.Expr),
				// TODO: Handle this value instead of ignoring it
				Index: nil,
			}
		}
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
		// If the parameter has an annotatioj, ignore that
		if node.NamedChild(0).Type() == "modifiers" {
			return &ast.Field{
				Names: []*ast.Ident{ParseNode(node.NamedChild(2), source, ctx).(*ast.Ident)},
				Type:  ParseNode(node.NamedChild(1), source, ctx).(ast.Expr),
			}
		}
		return &ast.Field{
			Names: []*ast.Ident{ParseNode(node.NamedChild(1), source, ctx).(*ast.Ident)},
			Type:  ParseNode(node.NamedChild(0), source, ctx).(ast.Expr),
		}
	case "inferred_parameters":
		params := &ast.FieldList{}
		for _, param := range Children(node) {
			params.List = append(params.List, &ast.Field{
				Names: []*ast.Ident{ParseNode(param, source, ctx).(*ast.Ident)},
				// When we're not sure what parameters to infer, set them as interface
				// values to avoid a panic
				Type: &ast.Ident{Name: "interface{}"},
			})
		}
		return params
	case "scoped_type_identifier":
		// This contains a reference to the type of a nested class
		// Ex: LinkedList.Node
		return &ast.StarExpr{X: &ast.Ident{Name: node.Content(source)}}
	case "method_reference":
		// This refers to manually selecting a function from a specific class and
		// passing it in as an argument in the `func(className::methodName)` style

		return &ast.SelectorExpr{
			X:   ParseNode(node.NamedChild(0), source, ctx).(ast.Expr),
			Sel: ParseNode(node.NamedChild(0), source, ctx).(*ast.Ident),
		}
	case "this":
		return &ast.Ident{Name: ShortName(ctx.className)}
	case "identifier":
		return &ast.Ident{Name: node.Content(source)}
	case "integral_type":
		return &ast.Ident{Name: node.Content(source)}
	case "floating_point_type": // Can be either `float` or `double`
		return &ast.Ident{Name: node.Content(source)}
	case "void_type":
		return &ast.Ident{}
	case "boolean_type":
		return &ast.Ident{Name: node.Content(source)}
	case "generic_type":
		// A generic type is any type that is of the form GenericType<T>
		return &ast.Ident{Name: node.NamedChild(0).Content(source)}
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
	case "decimal_floating_point_literal":
		// This is something like 1.3D
		return &ast.Ident{Name: node.Content(source)}
	case "string_literal":
		return &ast.Ident{Name: node.Content(source)}
	case "character_literal":
		return &ast.Ident{Name: node.Content(source)}
	case "true", "false":
		return &ast.Ident{Name: node.Content(source)}
	case "comment": // Ignore comments
		return nil
	default:
		panic(fmt.Sprintf("Unknown node type: %v", node.Type()))
	}
	return nil
}

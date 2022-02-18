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
	case "ERROR":
		return &ast.BadStmt{}
	case "program":
		// A program contains all the source code, in this case, one `class_declaration`
		program := &ast.File{
			Name: &ast.Ident{Name: "main"},
		}

		for _, c := range Children(node) {
			switch c.Type() {
			case "class_declaration":
				program.Decls = ParseDecls(c, source, ctx)
			case "import_declaration":
				program.Imports = append(program.Imports, ParseNode(c, source, ctx).(*ast.ImportSpec))
			}
		}
		return program
	case "field_declaration":
		var fieldType ast.Expr
		var fieldName *ast.Ident

		for ind, c := range Children(node) {
			switch c.Type() {
			case "modifiers": // Ignore the modifiers for now
				// The variable's type will always follow the modifiers, if they are present
				fieldType = ParseExpr(node.NamedChild(ind+1), source, ctx)
				// The value will come one after that
				fieldName = ParseExpr(node.NamedChild(ind+2).NamedChild(0), source, ctx).(*ast.Ident)
			}
		}

		// If no modifiers were declared, then declare everything with the default
		// offsets
		if fieldType == nil {
			fieldType = ParseExpr(node.NamedChild(0), source, ctx)
			fieldName = ParseExpr(node.NamedChild(1).NamedChild(0), source, ctx).(*ast.Ident)
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
		return &ast.ImportSpec{Name: ParseExpr(node.NamedChild(0), source, ctx).(*ast.Ident)}
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
					Lhs: []ast.Expr{ParseExpr(node.NamedChild(i), source, ctx)},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{assignments[len(assignments)-1].(*ast.AssignStmt).Lhs[0]},
				})
			} else {
				names = append(names, ParseExpr(node.NamedChild(i), source, ctx))
				values = append(values, ParseExpr(node.NamedChild(i+1), source, ctx))
			}
		}

		if len(assignments) > 0 {
			return assignments
		}

		// Having no declarations in the assign stmt panics the parser
		if len(names) == 0 {
			panic("Assignment with no assignments")
		}

		return &ast.AssignStmt{Lhs: names, Tok: token.ASSIGN, Rhs: values}

	case "super":
		return &ast.BadExpr{}
	case "constructor_body", "block":
		body := &ast.BlockStmt{}
		for _, line := range Children(node) {
			if line.Type() == "comment" {
				continue
			}
			if stmt := TryParseStmt(line, source, ctx); stmt != nil {
				body.List = append(body.List, stmt)
			} else {
				// Try statements are ignored, so they return a list of statements
				body.List = append(body.List, ParseNode(line, source, ctx).([]ast.Stmt)...)
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
		if stmt := TryParseStmt(node.NamedChild(0), source, ctx); stmt != nil {
			return stmt
		}
		if exprs := TryParseStmts(node.NamedChild(0), source, ctx); exprs != nil {
			return exprs
		}
		return &ast.ExprStmt{X: ParseExpr(node.NamedChild(0), source, ctx)}
	case "try_statement":
		// We ignore try statements
		return ParseNode(node.NamedChild(0), source, ctx).(*ast.BlockStmt).List
	case "switch_label":
		if node.NamedChildCount() > 0 {
			return &ast.CaseClause{
				List: []ast.Expr{ParseExpr(node.NamedChild(0), source, ctx)},
			}
		}
		return &ast.CaseClause{}
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
			args = append(args, ParseExpr(c, source, ctx))
		}
		return args

	case "formal_parameters":
		params := &ast.FieldList{}
		for _, param := range Children(node) {
			params.List = append(params.List, ParseNode(param, source, ctx).(*ast.Field))
		}
		return params
	case "formal_parameter":
		// If the parameter has an annotation, ignore that
		if node.NamedChild(0).Type() == "modifiers" {
			return &ast.Field{
				Names: []*ast.Ident{ParseExpr(node.NamedChild(2), source, ctx).(*ast.Ident)},
				Type:  ParseExpr(node.NamedChild(1), source, ctx),
			}
		}
		return &ast.Field{
			Names: []*ast.Ident{ParseExpr(node.NamedChild(1), source, ctx).(*ast.Ident)},
			Type:  ParseExpr(node.NamedChild(0), source, ctx),
		}
	case "spread_parameter":
		// The spread paramater takes a list and separates it into multiple elements
		// Ex: addElements([]int elements...)

		// If the parameter is a reference type (ex: ...[]*Test), then the type is
		// a `StarExpr`, which is passed into the ellipsis
		if _, is := ParseNode(node.NamedChild(0), source, ctx).(*ast.StarExpr); is {
			return &ast.Field{
				Names: []*ast.Ident{ParseExpr(node.NamedChild(1).NamedChild(0), source, ctx).(*ast.Ident)},
				Type: &ast.Ellipsis{
					Elt: ParseExpr(node.NamedChild(0), source, ctx),
				},
			}
		}

		return &ast.Field{
			Names: []*ast.Ident{ParseExpr(node.NamedChild(0), source, ctx).(*ast.Ident)},
			Type: &ast.Ellipsis{
				// This comes as a variable declarator, but we only need need the identifier for the type
				Elt: ParseExpr(node.NamedChild(1).NamedChild(0), source, ctx),
			},
		}
	case "inferred_parameters":
		params := &ast.FieldList{}
		for _, param := range Children(node) {
			params.List = append(params.List, &ast.Field{
				Names: []*ast.Ident{ParseExpr(param, source, ctx).(*ast.Ident)},
				// When we're not sure what parameters to infer, set them as interface
				// values to avoid a panic
				Type: &ast.Ident{Name: "interface{}"},
			})
		}
		return params
	case "comment": // Ignore comments
		return nil
	}
	panic(fmt.Sprintf("Unknown node type: %v", node.Type()))
}

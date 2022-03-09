package main

import (
	"fmt"
	"go/ast"
	"unicode"

	log "github.com/sirupsen/logrus"
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
		log.WithFields(log.Fields{
			"parsed": node,
		}).Warn("Error parsing generic node")
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
	case "try_with_resources_statement":
		// Ignore try with resources statements as well
		// NOTE: This will also ignore the catch clause
		stmts := []ast.Stmt{ParseStmt(node.NamedChild(0), source, ctx)}
		return append(stmts, ParseStmt(node.NamedChild(1), source, ctx).(*ast.BlockStmt).List...)
	case "try_statement":
		// We ignore try statements
		return ParseStmt(node.NamedChild(0), source, ctx).(*ast.BlockStmt).List
	case "synchronized_statement":
		// A synchronized statement contains the variable to be synchronized, as
		// well as the block

		// Ignore the sychronized statement
		return ParseStmt(node.NamedChild(1), source, ctx).(*ast.BlockStmt).List
	case "switch_label":
		if node.NamedChildCount() > 0 {
			return &ast.CaseClause{
				List: []ast.Expr{ParseExpr(node.NamedChild(0), source, ctx)},
			}
		}
		return &ast.CaseClause{}
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

		switch ParseExpr(node.NamedChild(0), source, ctx).(type) {
		case *ast.StarExpr:
			// If the parameter is a reference type (ex: ...[]*Test), then the type is
			// a `StarExpr`, which is passed into the ellipsis
			return &ast.Field{
				Names: []*ast.Ident{ParseExpr(node.NamedChild(1).NamedChild(0), source, ctx).(*ast.Ident)},
				Type: &ast.Ellipsis{
					Elt: ParseExpr(node.NamedChild(0), source, ctx),
				},
			}
		case *ast.ArrayType:
			// Represents something such as `byte[]... name`
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

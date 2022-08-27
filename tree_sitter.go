package main

import (
	"fmt"
	"go/ast"

	"github.com/NickyBoy89/java2go/nodeutil"
	"github.com/NickyBoy89/java2go/symbol"
	log "github.com/sirupsen/logrus"
	sitter "github.com/smacker/go-tree-sitter"
)

// Inspect is a function for debugging that prints out every named child of a
// given node and the source code for that child
func Inspect(node *sitter.Node, source []byte) {
	for _, c := range nodeutil.NamedChildrenOf(node) {
		fmt.Println(c, c.Content(source))
	}
}

// CapitalizeIdent capitalizes the first letter of a `*ast.Ident` to mark the
// result as a public method or field
func CapitalizeIdent(in *ast.Ident) *ast.Ident {
	return &ast.Ident{Name: symbol.Uppercase(in.Name)}
}

// LowercaseIdent lowercases the first letter of a `*ast.Ident` to mark the
// result as a private method or field
func LowercaseIdent(in *ast.Ident) *ast.Ident {
	return &ast.Ident{Name: symbol.Lowercase(in.Name)}
}

// A Ctx is all the context that is needed to parse a single source file
type Ctx struct {
	// Used to generate the names of all the methods, as well as the names
	// of the constructors
	className string

	// Symbols for the current file being parsed
	currentFile  *symbol.FileScope
	currentClass *symbol.ClassScope

	// The symbols of the current
	localScope *symbol.Definition

	// Used when generating arrays, because in Java, these are defined as
	// arrType[] varName = {item, item, item}, and no class name data is defined
	// Can either be of type `*ast.Ident` or `*ast.StarExpr`
	lastType ast.Expr
}

// Clone performs a shallow copy on a `Ctx`, returning a new Ctx with its pointers
// pointing at the same things as the previous Ctx
func (c Ctx) Clone() Ctx {
	return Ctx{
		className:    c.className,
		currentFile:  c.currentFile,
		currentClass: c.currentClass,
		localScope:   c.localScope,
		lastType:     c.lastType,
	}
}

// ParseNode parses a given tree-sitter node and returns the ast representation
//
// This function is called when the node being parsed might not be a direct
// expression or statement, as those are parsed with `ParseExpr` and `ParseStmt`
// respectively
func ParseNode(node *sitter.Node, source []byte, ctx Ctx) interface{} {
	switch node.Type() {
	case "ERROR":
		log.WithFields(log.Fields{
			"parsed":    node.Content(source),
			"className": ctx.className,
		}).Warn("Error parsing generic node")
		return &ast.BadStmt{}
	case "program":
		// A program contains all the source code, in this case, one `class_declaration`
		program := &ast.File{
			Name: &ast.Ident{Name: "main"},
		}

		for _, c := range nodeutil.NamedChildrenOf(node) {
			switch c.Type() {
			case "package_declaration":
				program.Name = &ast.Ident{Name: c.NamedChild(0).NamedChild(int(c.NamedChild(0).NamedChildCount()) - 1).Content(source)}
			case "class_declaration", "interface_declaration":
				program.Decls = ParseDecls(c, source, ctx)
			case "import_declaration":
				program.Imports = append(program.Imports, ParseNode(c, source, ctx).(*ast.ImportSpec))
			}
		}
		return program
	case "field_declaration":
		var public bool

		if node.NamedChild(0).Type() == "modifiers" {
			for _, modifier := range nodeutil.UnnamedChildrenOf(node.NamedChild(0)) {
				if modifier.Type() == "public" {
					public = true
				}
			}
		}

		fieldType := ParseExpr(node.ChildByFieldName("type"), source, ctx)
		fieldName := ParseExpr(node.ChildByFieldName("declarator").ChildByFieldName("name"), source, ctx).(*ast.Ident)
		fieldName.Name = symbol.HandleExportStatus(public, fieldName.Name)

		// If the field is assigned to a value (ex: int field = 1)
		fieldAssignmentNode := node.ChildByFieldName("declarator").ChildByFieldName("value")
		if fieldAssignmentNode != nil {
			return &ast.ValueSpec{
				Names: []*ast.Ident{fieldName},
				Type:  fieldType,
				Values: []ast.Expr{
					ParseExpr(fieldAssignmentNode, source, ctx),
				},
			}
		}

		return &ast.Field{
			Names: []*ast.Ident{fieldName},
			Type:  fieldType,
		}
	case "import_declaration":
		return &ast.ImportSpec{Name: ParseExpr(node.NamedChild(0), source, ctx).(*ast.Ident)}
	case "method_declaration":
		comments := []*ast.Comment{}

		if node.NamedChild(0).Type() == "modifiers" {
			for _, modifier := range nodeutil.UnnamedChildrenOf(node.NamedChild(0)) {
				switch modifier.Type() {
				case "marker_annotation", "annotation":
					comments = append(comments, &ast.Comment{Text: "//" + modifier.Content(source)})
					if _, in := excludedAnnotations[modifier.Content(source)]; in {
						// If this entire method is ignored, we return an empty field, which
						// is handled by the logic that parses a class file
						return &ast.Field{}
					}
				}
			}
		}

		parameterTypes := []string{}

		parameters := &ast.FieldList{}

		for _, param := range nodeutil.NamedChildrenOf(node.ChildByFieldName("parameters")) {
			if param.Type() == "spread_parameter" {
				parameterTypes = append(parameterTypes, param.NamedChild(0).Content(source))
			} else {
				parameterTypes = append(parameterTypes, param.ChildByFieldName("type").Content(source))
			}
			parameters.List = append(parameters.List, ParseNode(param, source, ctx).(*ast.Field))
		}

		def := ctx.currentClass.FindMethodByName(node.ChildByFieldName("name").Content(source), parameterTypes)

		return &ast.Field{
			Doc:   &ast.CommentGroup{List: comments},
			Names: []*ast.Ident{&ast.Ident{Name: def.Name}},
			Type: &ast.FuncType{
				Params: parameters,
				Results: &ast.FieldList{List: []*ast.Field{
					&ast.Field{
						Type: &ast.Ident{Name: def.Type},
					},
				},
				},
			},
		}
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
		for _, c := range nodeutil.NamedChildrenOf(node) {
			args = append(args, ParseExpr(c, source, ctx))
		}
		return args

	case "formal_parameters":
		params := &ast.FieldList{}
		for _, param := range nodeutil.NamedChildrenOf(node) {
			params.List = append(params.List, ParseNode(param, source, ctx).(*ast.Field))
		}
		return params
	case "formal_parameter":
		if ctx.localScope != nil {
			paramDef := ctx.localScope.ParameterByName(node.ChildByFieldName("name").Content(source))
			if paramDef == nil {
				paramDef = &symbol.Definition{
					Name: node.ChildByFieldName("name").Content(source),
					Type: node.ChildByFieldName("type").Content(source),
				}
			}
			return &ast.Field{
				Names: []*ast.Ident{&ast.Ident{Name: paramDef.Name}},
				Type:  &ast.Ident{Name: paramDef.Type},
			}
		}
		return &ast.Field{
			Names: []*ast.Ident{ParseExpr(node.ChildByFieldName("name"), source, ctx).(*ast.Ident)},
			Type:  ParseExpr(node.ChildByFieldName("type"), source, ctx),
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
		for _, param := range nodeutil.NamedChildrenOf(node) {
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

package ng

import (
	"fmt"
	"go/ast"

	"github.com/NickyBoy89/java2go/ng/codegen"
	"github.com/NickyBoy89/java2go/parsing"
	"github.com/NickyBoy89/java2go/symbol"
	mapset "github.com/deckarep/golang-set/v2"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

const (
// decimal_integer_literal
// hex_integer_literal
// octal_integer_literal
// binary_integer_literal
// decimal_floating_point_literal
// hex_floating_point_literal
// true
// false
// character_literal
// string_literal
// null_literal
)

var source []byte

// From: https://github.com/tree-sitter/tree-sitter-java/blob/master/grammar.js#L94
func ParseProgram(p parsing.SourceFile) (ast.Node, error) {
	cursor := p.Ast.Walk()

	code := &ast.File{
		// The package name of the generated file
		// TODO: Change this to handle packages correctly
		Name: &ast.Ident{Name: "main"},
	}

	// TODO: Don't pass this as a global variable
	source = p.Source

	for _, child := range p.Ast.Children(cursor) {
		if IsStatement(child) {
			// TODO: Handle multiple top-level declarations
			if parsed, err := ParseStatement(child); err != nil {
				return nil, err
			} else {
				code.Decls = append(code.Decls, parsed...)
			}
		} else if child.Kind() == "method_declaration" {
			panic("TODO: Handle top-level method declarations")
		} else {
			return nil, fmt.Errorf(errUnknownNodeText, child.Kind())
		}
	}

	// TODO: Handle Name, Decls, Imports

	return code, nil
}

const errUnknownNodeText = "unhandled node kind: %s"

// From: https://github.com/tree-sitter/tree-sitter-java/blob/master/grammar.js#L539
func ParseStatement(node sitter.Node) ([]ast.Decl, error) {

	if IsDeclaration(node) {
		return ParseDeclaration(node)
	}

	// TODO: Handle semicolon
	switch node.Kind() {
	case "labeled_statement":
	case "if_statement":
	case "while_statement":
	case "for_statement":
	case "enhanced_for_statement":
	case "block":
	case "assert_statement":
	case "do_statement":
	case "break_statement":
	case "continue_statement":
	case "return_statement":
	case "yield_statement":
	case "switch_expression": // switch statements and expressions are identical
	case "synchronized_statement":
	case "local_variable_declaration":
	case "throw_statement":
	case "try_statement":
	case "try_with_resources_statement":
	case "expression_statement":
	default:
		return nil, fmt.Errorf(errUnknownNodeText, node.Kind())
	}

	fmt.Printf("statement: %s\n", node.Kind())

	return nil, nil
}

func HasModifiers(node sitter.Node) bool {
	return node.ChildCount() > 0 && node.NamedChild(0).Kind() == "modifiers"
}

func UnimplementedField(node sitter.Node, name string) {
	if node.ChildByFieldName(name) != nil {
		panic("TODO: Unimplemented field " + name)
	}
}

func ParseModifiers(node sitter.Node) (mapset.Set[string], error) {
	cursor := node.Walk()

	mods := mapset.NewSet[string]()

	for _, child := range node.Children(cursor) {
		switch child.Kind() {
		case "public":
		case "protected":
		case "private":
		case "abstract":
		case "static":
		case "final":
		case "strictfp":
		case "default":
		case "synchronized":
		case "native":
		case "transient":
		case "volatile":
		case "sealed":
		case "non-sealed":
		case "annotation", "marker_annotation":
			panic("TODO: Implement annotations")
		default:
			return nil, fmt.Errorf(errUnknownNodeText, node.Kind())
		}

		mods.Add(child.Kind())
	}

	return mods, nil
}

func IdentToString(node sitter.Node, source []byte) string {
	return node.Utf8Text(source)
}

const (
	ModifierPublic       = "public"
	ModifierProtected    = "protected"
	ModifierPrivate      = "private"
	ModifierAbstract     = "abstract"
	ModifierStatic       = "static"
	ModifierFinal        = "final"
	ModifierStrictfp     = "strictfp"
	ModifierDefault      = "default"
	ModifierSynchronized = "synchronized"
	ModifierNative       = "native"
	ModifierTransient    = "transient"
	ModifierVolatile     = "volatile"
	ModifierSealed       = "sealed"
	ModifierNonSealed    = "non-sealed"
)

func HandleModifiers(node sitter.Node) (mapset.Set[string], error) {
	var err error
	mods := mapset.NewSet[string]()

	if HasModifiers(node) {
		mods, err = ParseModifiers(*node.NamedChild(0))
		if err != nil {
			return nil, err
		}
	}

	return mods, nil
}

func HandleAccessModifierRename(ident string, mods mapset.Set[string]) string {
	if mods.Contains(ModifierPublic, ModifierProtected) {
		return symbol.Uppercase(ident)
	} else if mods.Contains(ModifierPrivate) {
		return symbol.Lowercase(ident)
	} else {
		// TODO: Re-implement this with knowledge of default access modifiers and modules
		return symbol.Uppercase(ident)
	}
}

// A class declaration is converted to a struct with some additional changes
//
// Its return value is a list of declarations for its own struct, as well as
// any methods and fields
//
// Notable changes to the class while parsing include:
// 1. Its name is renamed with respect to the access modifier (public, private)
func ParseClassDeclaration(node sitter.Node) ([]ast.Decl, error) {

	decls := []ast.Decl{}

	mods, err := HandleModifiers(node)
	if err != nil {
		return nil, err
	}

	name := node.ChildByFieldName("name").Utf8Text(source)
	name = HandleAccessModifierRename(name, mods)

	// TODO: Add class fields
	decls = append(decls, codegen.NewStruct(name, &ast.FieldList{List: []*ast.Field{}}))

	UnimplementedField(node, "type_parameters")
	UnimplementedField(node, "superclass")
	UnimplementedField(node, "interfaces")
	UnimplementedField(node, "permits")

	// body
	bodyDecls, classFields, err := ParseClassBody(*node.ChildByFieldName("body"))
	if err != nil {
		return nil, err
	}

	_ = classFields

	decls = append(decls, bodyDecls...)

	return decls, nil
}

func ParseIdentifier(node sitter.Node) string {
	return node.Utf8Text(source)
}

func ParseClassBody(node sitter.Node) ([]ast.Decl, *ast.FieldList, error) {
	cursor := node.Walk()

	decls := []ast.Decl{}
	fields := &ast.FieldList{List: []*ast.Field{}}

	for _, child := range node.NamedChildren(cursor) {
		var err error
		var decl ast.Decl

		panic("TODO: Why are we trying to parse specific things in a class")
		switch child.Kind() {
		case "field_declaration":
			if field, err := ParseFieldDeclaration(child); err != nil {
				return nil, nil, err
			} else {
				fields.List = append(fields.List, field)
			}
		case "record_declaration":
			panic("TODO: Unimplemented")
		case "method_declaration":
			decl, err = ParseMethodDeclaration(child)
		case "compact_constructor_declaration": // For records.
			panic("TODO: Unimplemented")
		case "class_declaration":
			// TODO: See if this should instead use ParseDeclaration
			panic("TODO: Unimplemented")
		case "interface_declaration":
			panic("TODO: Unimplemented")
		case "annotation_type_declaration":
			panic("TODO: Unimplemented")
		case "enum_declaration":
			panic("TODO: Unimplemented")
		case "block":
			panic("TODO: Unimplemented")
		case "static_initializer":
			panic("TODO: Unimplemented")
		case "constructor_declaration":
			decl, err = ParseConstructorDeclaration(child)
		default:
			return nil, nil, fmt.Errorf(errUnknownNodeText, child.Kind())
		}

		if err != nil {
			return nil, nil, err
		}

		decls = append(decls, decl)
	}

	return decls, fields, nil
}

func ParseFormalParameters(node sitter.Node) (*ast.FieldList, error) {
	cursor := node.Walk()

	params := &ast.FieldList{List: []*ast.Field{}}

	for _, child := range node.NamedChildren(cursor) {
		switch child.Kind() {
		case "receiver_parameter":
			ParseReceiverParameter(child)
		case "formal_parameter":
			param, err := ParseFormalParameter(child)
			if err != nil {
				return nil, err
			}
			params.List = append(params.List, param)
		case "spread_parameter":

		}
	}

	return params, nil
}

func ParseArrayType(node sitter.Node) (*ast.ArrayType, error) {
	element := *node.ChildByFieldName("element")
	dimensions := *node.ChildByFieldName("dimensions")

	parsedElt, err := ParseUnannotatedType(element)
	if err != nil {
		return nil, err
	}

	// TODO: Handle dimensions
	_ = dimensions
	return &ast.ArrayType{
		Elt: parsedElt,
	}, nil
}

func ParseUnannotatedType(node sitter.Node) (ast.Expr, error) {
	switch node.Kind() {
	case "array_type":
		return ParseArrayType(node)
	case "void_type":
	case "integral_type":
	case "floating_point_type":
	case "boolean_type":
	case "identifier", "type_identifier":
		return ast.NewIdent(node.Utf8Text(source)), nil
	case "scoped_type_identifier":
	case "generic_type":
	}

	return ast.NewIdent(node.Kind()), nil
}

func ParseFormalParameter(node sitter.Node) (*ast.Field, error) {
	mods, err := HandleModifiers(node)
	if err != nil {
		return nil, err
	}

	// TODO: Handle modifiers
	_ = mods

	paramType := *node.ChildByFieldName("type")
	parsedType, err := ParseUnannotatedType(paramType)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%v\n", paramType.Utf8Text(source))
	fmt.Printf("%v\n", paramType.NextSibling().Kind())

	// TODO: Debug this declarator
	// id, err := ParseVariableDeclaratorId(*paramType.NextSibling())
	// if err != nil {
	// 	return nil, err
	// }
	// _ = id

	return &ast.Field{
		Names: []*ast.Ident{ast.NewIdent("test")},
		Type:  parsedType,
	}, nil
}

// TODO: Determine the return type
func ParseVariableDeclaratorId(node sitter.Node) (*ast.Ident, error) {
	name := node.ChildByFieldName("name")
	dimensions := node.ChildByFieldName("dimensions")
	if dimensions != nil {
		return nil, fmt.Errorf("TODO: Handle dimensions")
	}

	return ast.NewIdent(name.Utf8Text(source)), nil
}

func ParseReceiverParameter(node sitter.Node) (*ast.FieldList, error) {
	cursor := node.Walk()
	for _, child := range node.NamedChildren(cursor) {
		switch child.Kind() {
		case "marker_annotation", "annotation":
			panic("TODO: Handle annotations")
		default:
			return nil, fmt.Errorf("Unknown receiver parameter type: %s", child.Kind())
		}
	}
	panic("TODO: Handle receiver parameter")
}

// constructor_body: $ => seq(
//
//	'{',
//	optional($.explicit_constructor_invocation),
//	repeat($.statement),
//	'}',
//
// ),
func ParseConstructorBody(node sitter.Node) (*ast.BlockStmt, error) {
	UnimplementedField(node, "explicit_constructor_invocation")
	// TODO: Handle constructor body
	return &ast.BlockStmt{}, nil
}

func ParseBlock(node sitter.Node) (*ast.BlockStmt, error) {
	// TODO: Handle blocks
	return &ast.BlockStmt{}, nil
}

// TODO: Determine the return type
func ParseDeclaration(node sitter.Node) ([]ast.Decl, error) {
	switch node.Kind() {
	case "module_declaration":
	case "package_declaration":
	case "import_declaration":
	case "class_declaration":
		return ParseClassDeclaration(node)
	case "record_declaration":
	case "interface_declaration":
	case "annotation_type_declaration":
	case "enum_declaration":
	default:
		return nil, fmt.Errorf(errUnknownNodeText, node.Kind())
	}

	fmt.Printf("declaration: %s\n", node.Kind())

	return nil, nil
}

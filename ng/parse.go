package ng

import (
	"fmt"
	"go/ast"

	"github.com/NickyBoy89/java2go/ng/codegen"
	"github.com/NickyBoy89/java2go/parsing"
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
	fmt.Printf("program { kind: %s }\n", p.Ast.Kind())
	cursor := p.Ast.Walk()

	source = p.Source

	for _, child := range p.Ast.Children(cursor) {
		if IsStatement(child) {
			if parsed, err := ParseStatement(child); err != nil {
				return nil, err
			} else {
				return parsed, nil
			}
		} else if child.Kind() == "method_declaration" {
			fmt.Println("Method declaration")
		} else {
			return nil, fmt.Errorf(errUnknownNodeText, child.Kind())
		}
	}

	return nil, nil
}

const errUnknownNodeText = "unhandled node kind: %s"

// From: https://github.com/tree-sitter/tree-sitter-java/blob/master/grammar.js#L539
func ParseStatement(node sitter.Node) (ast.Node, error) {

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

func ParseModifiers(node sitter.Node) error {
	cursor := node.Walk()

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
			return fmt.Errorf(errUnknownNodeText, node.Kind())
		}
	}

	return nil
}

// A class declaration is converted to a struct with some additional changes
// 1. Yes
func ParseClassDeclaration(node sitter.Node) (ast.Node, error) {
	if HasModifiers(node) {
		ParseModifiers(*node.NamedChild(0))
	}

	var s ast.Node = codegen.NewStruct("test", &ast.FieldList{List: []*ast.Field{}})

	// name
	ParseIdentifier(*node.ChildByFieldName("name"))

	UnimplementedField(node, "type_parameters")
	UnimplementedField(node, "superclass")
	UnimplementedField(node, "interfaces")
	UnimplementedField(node, "permits")

	// body
	ParseClassBody(*node.ChildByFieldName("body"))

	return s, nil
}

func ParseIdentifier(node sitter.Node) error {
	// TODO: Handle identifiers
	return nil
}

func ParseClassBody(node sitter.Node) error {
	// TODO: Handle class body
	return nil
}

func ParseDeclaration(node sitter.Node) (ast.Node, error) {
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

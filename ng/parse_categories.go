package ng

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
)

func IsStatement(node sitter.Node) bool {

	if IsDeclaration(node) {
		return true
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
		return false
	}

	return true
}

func IsDeclaration(node sitter.Node) bool {
	switch node.Kind() {
	case "module_declaration":
		fallthrough
	case "package_declaration":
		fallthrough
	case "import_declaration":
		fallthrough
	case "class_declaration":
		fallthrough
	case "record_declaration":
		fallthrough
	case "interface_declaration":
		fallthrough
	case "annotation_type_declaration":
		fallthrough
	case "enum_declaration":
		return true
	}

	return false
}

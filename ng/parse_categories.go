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
		fallthrough
	case "if_statement":
		fallthrough
	case "while_statement":
		fallthrough
	case "for_statement":
		fallthrough
	case "enhanced_for_statement":
		fallthrough
	case "block":
		fallthrough
	case "assert_statement":
		fallthrough
	case "do_statement":
		fallthrough
	case "break_statement":
		fallthrough
	case "continue_statement":
		fallthrough
	case "return_statement":
		fallthrough
	case "yield_statement":
		fallthrough
	case "switch_expression": // switch statements and expressions are identical
		fallthrough
	case "synchronized_statement":
		fallthrough
	case "local_variable_declaration":
		fallthrough
	case "throw_statement":
		fallthrough
	case "try_statement":
		fallthrough
	case "try_with_resources_statement":
		fallthrough
	case "expression_statement":
		return true
	}

	return false
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

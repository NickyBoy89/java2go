package symbol

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// Go reserved keywords that are not Java keywords, and create invalid code
var reservedKeywords = []string{"chan", "defer", "fallthrough", "func", "go", "map", "range", "select", "struct", "type"}

// IsReserved tests if a given identifier conflicts with a Go reserved keyword
func IsReserved(name string) bool {
	for _, keyword := range reservedKeywords {
		if keyword == name {
			return true
		}
	}
	return false
}

// TypeOfLiteral returns the corresponding type for a Java literal
func TypeOfLiteral(node *sitter.Node, source []byte) string {
	var originalType string

	switch node.Type() {
	case "decimal_integer_literal":
		switch node.Content(source)[len(node.Content(source))-1] {
		case 'L':
			originalType = "long"
		default:
			originalType = "int"
		}
	case "hex_integer_literal":
		panic("here")
	case "decimal_floating_point_literal":
		switch node.Content(source)[len(node.Content(source))-1] {
		case 'D':
			originalType = "double"
		default:
			originalType = "float"
		}
	case "string_literal":
		originalType = "String"
	case "character_literal":
		originalType = "char"
	}

	return originalType
}

// ResolveDefinition resolves a given definition for its type
// given its class scope, as well as the global scope
// It returns true if the definition was successfully resolved, and false otherwise
func ResolveDefinition(definition *Definition, fileScope *FileScope, globalScope *GlobalSymbols) bool {
	// Look in the class scope first
	if localClassDef := fileScope.BaseClass.FindClass(definition.Type); localClassDef != nil {
		// Every type in the local scope is a reference type, so prefix it with a pointer
		definition.Type = "*" + localClassDef.Name
		return true

	} else if globalDef, in := fileScope.Imports[definition.Type]; in { // Look through the imports
		// Find what package the type is in
		if packageDef := globalScope.FindPackage(globalDef); packageDef != nil {
			definition.Type = packageDef.FindClass(definition.Type).FindClass(definition.Type).Type
		}
		return true
	}

	// Unresolved
	return false
}

// ResolveChildren recursively resolves a definition and all of its children
// It returns true if all definitions were resolved correctly, and false otherwise
func ResolveChildren(definition *Definition, fileScope *FileScope, globalScope *GlobalSymbols) bool {
	result := ResolveDefinition(definition, fileScope, globalScope)
	for _, child := range definition.Children {
		result = ResolveChildren(child, fileScope, globalScope) && result
	}
	return result
}

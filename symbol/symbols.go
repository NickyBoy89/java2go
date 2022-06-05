package symbol

import (
	"fmt"

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
func ResolveDefinition(definition *Definition, classScope *ClassScope, globalScope *GlobalScope) bool {
	// Look in the class scope first
	if localClassDef := classScope.FindClass(definition.Type()); localClassDef != nil {
		// Every type in the local scope is a reference type, so prefix it with a pointer
		definition.typ = "*" + localClassDef.Name()
		return true

	} else if globalDef, in := classScope.Imports[definition.Type()]; in { // Look through the imports
		// Find what package the type is in
		if packageDef := globalScope.FindPackage(globalDef); packageDef != nil {
			definition.typ = packageDef.FindClass(definition.Type()).FindClass(definition.Type()).Type()
		}
		return true
	}

	// Unresolved
	return false
}

// ResolveChildren recursively resolves a definition and all of its children
// It returns true if all definitions were resolved correctly, and false otherwise
func ResolveChildren(definition *Definition, classScope *ClassScope, globalScope *GlobalScope) bool {
	result := ResolveDefinition(definition, classScope, globalScope)
	for _, child := range definition.children {
		result = ResolveChildren(child, classScope, globalScope) && result
	}
	return result
}

// A Scope represents a generic view into some number of definitions
type Scope interface {
	FindMethodByName(name string, ignoredParameterTypes []string) *Definition
}

// A GlobalScope represents a global view of all the packages in the parsed source
type GlobalScope struct {
	// Every package's path associatedd with its definition
	packages map[string]*PackageScope
}

func (gs GlobalScope) String() string {
	return fmt.Sprintf("Global: [%v]", gs.packages)
}

// FindPackage looks up a package's path in the global scope, and returns it
func (gs *GlobalScope) FindPackage(name string) *PackageScope {
	return gs.packages[name]
}

// PackageScope represents a single package, which can contain one or more files
type PackageScope struct {
	// Maps the file's name to its definitions
	files map[string]*FileScope
}

func (ps PackageScope) String() string {
	return fmt.Sprintf("Package: [%v]", ps.files)
}

// FindClass searches for a class in the given package and returns a scope for it
// the class may be the subclass of another class
func (ps *PackageScope) FindClass(name string) *ClassScope {
	for _, fileScope := range ps.files {
		if fileScope.Class.originalName == name {
			return fileScope
		}
		for _, subclass := range fileScope.Subclasses {
			class := subclass.FindClass(name)
			if class != nil {
				return fileScope
			}
		}
	}
	return nil
}

// FileScope represents the scope in a single source file, that can contain one
// or more source classes
type FileScope struct {
	// The global package that the file is located in
	Package string
	// Every external package that is imported into the file
	// Formatted as map[ImportedType: full.package.path]
	Imports map[string]string
	// The base class that is in the file
	BaseClass *ClassScope
}

// ClassScope represents a single defined class, and the declarations in it
type ClassScope struct {
	// The definition for the class defined within the class
	Class *Definition
	// Every class that is nested within the base class
	Subclasses []*ClassScope
	// Any normal and static fields associated with the class
	Fields []*Definition
	// Methods and constructors
	Methods []*Definition
}

// FindMethodByDisplayName searches for a given method by its display name
// If some ignored parameter types are specified as non-nil, it will skip over
// any function that matches these ignored parameter types exactly
func (cs *ClassScope) FindMethodByName(name string, ignoredParameterTypes []string) *Definition {
	return findMethodWithComparison(func(method *Definition) { return method.originalName == name }, ignoredParameterTypes)
}

// FindMethodByDisplayName searches for a given method by its display name
// If some ignored parameter types are specified as non-nil, it will skip over
// any function that matches these ignored parameter types exactly
func (cs *ClassScope) FindMethodByDisplayName(name string, ignoredParameterTypes []string) *Definition {
	return findMethodWithComparison(func(method *Definition) { return method.Name() == name }, ignoredParameterTypes)
}

func (cs *ClassScope) findMethodWithComparison(comparison func(method *Definition) bool, ignoredParameterTypes []string) *Definition {
	for _, method := range cs.Methods {
		if comparison(method) {
			// If no parameters were specified to ignore, then return the first match
			if ignoredParameterTypes == nil {
				return method
			} else if len(method.parameters) != len(ignoredParameterTypes) { // Size of parameters were not equal, instantly not equal
				return method
			}

			// Check the remaining paramters one-by-one
			for index, parameter := range method.parameters {
				if param.originalType != ignoredParameterTypes[index] {
					return method
				}
			}
		}
	}

	// Not found
	return nil
}

// FindClass searches through a class file and returns the definition for the
// found class, or nil if none was found
func (cs *ClassScope) FindClass(name string) *Definition {
	if cs.Class.originalName == name {
		return cs.Class
	}
	for _, subclass := range cs.Subclasses {
		class := subclass.FindClass(name)
		if class != nil {
			return class
		}
	}
	return nil
}

// FindFieldByName searches for a field by its original name, and returns its definition
// or nil if none was found
func (cs *ClassScope) FindFieldByName(name string) *Definition {
	for _, field := range cs.Fields {
		if field.originalName == name {
			return field
		}
	}
	return nil
}

func (cs *ClassScope) FindFieldByDisplayName(name string) *Definition {
	for _, field := range cs.Fields {
		if field.name == name {
			return field
		}
	}
	return nil
}

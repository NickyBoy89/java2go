package main

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type ClassFile struct {
	Name    string
	Package *PackageScope
	Imports []*PackageScope
}

// PackageScope is the package declaration for a single file
// it contains the name of the package (ex: "util"), as well as a pointer to
// another scope, which contains the rest of the scope (ex: "com.example")
// When the scope variable is `nil`, then you are at the root of the parent
type PackageScope struct {
	Scope []string
}

func (ps *PackageScope) String() string {
	var total strings.Builder
	for ind, item := range ps.Scope {
		total.WriteString(item)
		if ind < len(ps.Scope)-1 {
			total.WriteRune('.')
		}
	}
	return total.String()
}

// ParseScope takes a identifier from an import node, and the source code, and
// parses it as a `PackageScope` type
func ParseScope(node *sitter.Node, source []byte) *PackageScope {
	pack := &PackageScope{}
	// A `scoped_identifier` contains two items, one for the scope, and the other
	// for the name of the current package

	if node.Type() != "package_declaration" {
		pack.Scope = []string{node.NamedChild(1).Content(source)}
	}

	scope := node.NamedChild(0)
	for scope.Type() == "scoped_identifier" {
		pack.Scope = append(pack.Scope, scope.NamedChild(1).Content(source))
		scope = scope.NamedChild(0)
	}

	pack.Scope = append(pack.Scope, scope.Content(source))

	// Flip the order of the scope, because it is the wrong direction
	for ind := 0; ind < len(pack.Scope)/2; ind++ {
		rev := len(pack.Scope) - 1 - ind
		pack.Scope[ind], pack.Scope[rev] = pack.Scope[rev], pack.Scope[ind]
	}

	return pack
}

// ExtractImports takes in a tree-sitter node and the source code, returning the
// parsed source file's imports, and other package-related data
func ExtractImports(node *sitter.Node, source []byte) *ClassFile {
	class := &ClassFile{}

	// If this node has a package declaration, add it to the current class
	if node.Type() == "package_declaration" {
		class.Package = ParseScope(node, source)
	}

	// If the node is an import node, return that as a single import
	if node.Type() == "import_declaration" {
		class.Imports = append(class.Imports, ParseScope(node.NamedChild(0), source))
		return class
	}

	// Go through the children of the node to find everything else
	for _, child := range Children(node) {
		// Extract the node of the class being parsed
		if node.Type() == "class_declaration" && child.Type() == "identifier" {
			class.Name = child.Content(source)
		}

		// Go through the children of the current node to find the imports
		other := ExtractImports(child, source)
		if len(other.Imports) > 0 {
			class.Imports = append(class.Imports, other.Imports...)
		}
		// If the class name is unknown, and it has been found in one of the
		// children, populate it
		if class.Name == "" {
			class.Name = other.Name
		}

		if class.Package == nil {
			class.Package = other.Package
		}
	}

	return class
}

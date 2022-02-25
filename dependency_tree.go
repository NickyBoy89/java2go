package main

import (
	sitter "github.com/smacker/go-tree-sitter"
)

type ClassFile struct {
	Dependencies []string
}

type JavaPackage struct {
}

func ExtractImports(node *sitter.Node, source []byte, className string) (string, []string) {
	imports := []string{}

	if node.Type() == "import_declaration" {
		return className, []string{node.NamedChild(0).Content(source)}
	}

	for _, child := range Children(node) {
		if node.Type() == "class_declaration" && child.Type() == "identifier" {
			className = child.Content(source)
		}
		cn, imp := ExtractImports(child, source, className)
		if len(imp) > 0 {
			imports = append(imports, imp...)
		}
		if className == "" {
			className = cn
		}
	}
	return className, imports
}

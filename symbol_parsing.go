package main

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"unicode"

	"github.com/NickyBoy89/java2go/symbol"
	sitter "github.com/smacker/go-tree-sitter"
)

// Uppercase uppercases the first character of the given string
func Uppercase(name string) string {
	return string(unicode.ToUpper(rune(name[0]))) + name[1:]
}

// Lowercase lowercases the first character of the given string
func Lowercase(name string) string {
	return string(unicode.ToLower(rune(name[0]))) + name[1:]
}

// HandleExportStatus is a convenience method for renaming methods that may be
// either public or private, and need to be renamed
func HandleExportStatus(exported bool, name string) string {
	if exported {
		return Uppercase(name)
	}
	return Lowercase(name)
}

func nodeToStr(node any) string {
	var s bytes.Buffer
	err := printer.Fprint(&s, token.NewFileSet(), node)
	if err != nil {
		panic(err)
	}
	return s.String()
}

// ExtractDefinitions generates a symbol table for a single class file.
func ExtractDefinitions(root *sitter.Node, source []byte) *symbol.FileScope {
	var pack string
	imports := make(map[string]string)
	for i := 0; i < int(root.NamedChildCount()); i++ {
		switch root.NamedChild(i).Type() {
		case "package_declaration":
			pack = root.NamedChild(i).NamedChild(0).Content(source)
		case "import_declaration":
			imports[root.NamedChild(i).NamedChild(0).ChildByFieldName("name").Content(source)] = root.NamedChild(i).NamedChild(0).ChildByFieldName("scope").Content(source)
		}
	}

	return &symbol.FileScope{
		Imports:   imports,
		Package:   pack,
		BaseClass: parseClassScope(root.NamedChild(int(root.NamedChildCount())-1), source),
	}
}

func parseClassScope(root *sitter.Node, source []byte) *symbol.ClassScope {
	var public bool
	// Rename the type based on the public/static rules
	if root.NamedChild(0).Type() == "modifiers" {
		for cursor := sitter.NewTreeCursor(root.NamedChild(0).Child(0)); !cursor.CurrentNode().IsNull(); cursor.GoToNextSibling() {
			if cursor.CurrentNode().Type() == "public" {
				public = true
			}
		}
	}

	// DEBUG
	if root.ChildByFieldName("name").Type() != "identifier" {
		panic("Assertion failed")
	}
	className := root.ChildByFieldName("name").Content(source)
	scope := &symbol.ClassScope{
		Class: &symbol.Definition{
			OriginalName: className,
			Name:         HandleExportStatus(public, className),
		},
	}

	for _, node := range Children(root.ChildByFieldName("body")) {
		switch node.Type() {
		case "field_declaration":
			var public bool
			// Rename the type based on the public/static rules
			if node.NamedChild(0).Type() == "modifiers" {
				for _, modifier := range UnnamedChildren(node.NamedChild(0)) {
					if modifier.Type() == "public" {
						public = true
					}
				}
			}

			if node.ChildByFieldName("declarator").ChildByFieldName("name").Type() != "identifier" {
				panic("Assertion!")
			}
			name := node.ChildByFieldName("declarator").ChildByFieldName("name").Content(source)
			typeNode := node.ChildByFieldName("type")

			var fieldType string
			// Scoped type identifiers are in a format such as RemotePackage.ClassName
			// To handle this, we remove the RemotePackage part, and depend on the later
			// type resolution to figure things out
			// TODO: Fix this to allow partial lookups, instead of throwing out this information
			if typeNode.Type() == "scoped_type_identifier" {
				if typeNode.NamedChild(int(typeNode.NamedChildCount())-1).Type() != "identifier" {
					panic("assertion")
				}
				fieldType = typeNode.NamedChild(int(typeNode.NamedChildCount()) - 1).Content(source)
			} else {
				if typeNode.Type() != "identifier" {
					panic("Assertion")
				}
				fieldType = typeNode.Content(source)
			}
			scope.Fields = append(scope.Fields, &symbol.Definition{
				Name:         HandleExportStatus(public, name),
				OriginalName: name,
				Type:         fieldType,
				OriginalType: node.ChildByFieldName("type").Content(source),
			})
		case "method_declaration", "constructor_declaration":
			var public bool
			// Rename the type based on the public/static rules
			if node.NamedChild(0).Type() == "modifiers" {
				for _, modifier := range UnnamedChildren(node.NamedChild(0)) {
					if modifier.Type() == "public" {
						public = true
					}
				}
			}

			if node.ChildByFieldName("name").Type() != "identifier" {
				panic("Assertion")
			}
			name := node.ChildByFieldName("name").Content(source)
			declaration := &symbol.Definition{
				Name:         HandleExportStatus(public, name),
				OriginalName: name,
			}

			if node.Type() == "method_declaration" {
				declaration.OriginalType = node.ChildByFieldName("type").Content(source)
				if node.ChildByFieldName("type").Type() != "identifier" {
					panic("Assertion")
				}
				declaration.Type = node.ChildByFieldName("type").Content(source)
			} else {
				declaration.Rename(HandleExportStatus(public, "New") + name)
				// A constructor returns itself
				declaration.Constructor = true
				declaration.Type = name
			}

			for _, parameter := range Children(node.ChildByFieldName("parameters")) {
				parsed := ParseNode(parameter, source, Ctx{}).(*ast.Field)
				name := nodeToStr(parsed.Names[0])
				if parameter.Type() == "spread_parameter" {
					declaration.Parameters = append(declaration.Parameters, &symbol.Definition{
						OriginalName: name,
						OriginalType: parameter.NamedChild(0).Content(source),
						Type:         nodeToStr(parsed.Type),
						Name:         name,
					})
				} else {
					declaration.Parameters = append(declaration.Parameters, &symbol.Definition{
						OriginalName: name,
						OriginalType: parameter.ChildByFieldName("type").Content(source),
						Type:         nodeToStr(parsed.Type),
						Name:         name,
					})
				}
			}

			if node.ChildByFieldName("body") != nil {
				methodScope := parseScope(node.ChildByFieldName("body"), source)
				if !methodScope.IsEmpty() {
					declaration.Children = append(declaration.Children, methodScope.Children...)
				}
			}

			scope.Methods = append(scope.Methods, declaration)
		case "class_declaration", "interface_declaration", "enum_declaration":
			other := parseClassScope(node, source)
			// Any subclasses will be renamed to part of their parent class
			other.Class.Rename(scope.Class.Name + other.Class.Name)
			scope.Subclasses = append(scope.Subclasses, other)
		}
	}

	return scope
}

func parseScope(root *sitter.Node, source []byte) *symbol.Definition {
	def := &symbol.Definition{}
	var node *sitter.Node
	for i := 0; i < int(root.NamedChildCount()); i++ {
		node = root.NamedChild(i)
		switch node.Type() {
		case "local_variable_declaration":
			name := nodeToStr(ParseExpr(node.ChildByFieldName("declarator").ChildByFieldName("name"), source, Ctx{}))
			def.Children = append(def.Children, &symbol.Definition{
				OriginalName: name,
				OriginalType: node.ChildByFieldName("type").Content(source),
				Type:         nodeToStr(ParseExpr(node.ChildByFieldName("type"), source, Ctx{})),
				Name:         name,
			})
		case "for_statement", "enhanced_for_statement", "while_statement", "if_statement":
			def.Children = append(def.Children, parseScope(node, source))
		}
	}
	return def
}

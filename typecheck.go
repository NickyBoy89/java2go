package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"

	sitter "github.com/smacker/go-tree-sitter"
)

// A GlobalScope represents a global look of all the packages that make up the
// entirety of the parsed input
type GlobalScope struct {
	// Packages is the full path of the package associated with its definition
	packages map[string]*PackageScope
}

func (gs GlobalScope) String() string {
	return fmt.Sprintf("Global: [%v]", gs.packages)
}

// PackageScope represents a single package, which can contain one or more files
type PackageScope struct {
	// Maps the file's name to its definitions
	files map[string]*ClassScope
}

func (ps PackageScope) String() string {
	return fmt.Sprintf("Package: [%v]", ps.files)
}

// ClassScope contains the global and local scopes for a single file
// if a file contains multiply classes, all the definitions are folded into
// one ClassScope
type ClassScope struct {
	Package string
	Imports map[string]string
	Classes []*Definition
	Fields  []*Definition
	Methods []*Definition
}

func (cs ClassScope) String() string {
	return fmt.Sprintf("Classes: %v Fields: %v Methods: %v", cs.Classes, cs.Fields, cs.Methods)
}

// FindMethod looks for a given method by its name in a class definition
func (cs *ClassScope) FindMethod(name string) *Definition {
	for _, method := range cs.Methods {
		if method.originalName == name {
			return method
		}
	}
	return nil
}

func (cs *ClassScope) FindClass(name string) *Definition {
	for _, class := range cs.Classes {
		if class.originalName == name {
			return class
		}
	}
	return nil
}

// A Definition contains information about a single entry
type Definition struct {
	// The original java name
	originalName string
	// The display name, may be different from the original name
	name string
	// Type of the object
	definitionType string
	// If the definition is a constructor
	constructor bool
	// If the object is a function, it has parameters
	parameters []*Definition
	// Children of the declaration, if the declaration is a scope
	children []*Definition
}

func (d Definition) Name() string {
	return d.name
}

func (d Definition) Type() string {
	return d.definitionType
}

func (d Definition) String() string {
	if d.originalName != d.name {
		return fmt.Sprintf("Name: %s (Was %s) Type: %s", d.name, d.originalName, d.definitionType)

	}
	return fmt.Sprintf("Name: %s Type: %s", d.name, d.definitionType)
}

// Rename renames a definition for a type so that it can be referenced later with
// the correct name
func (d *Definition) Rename(name string) {
	d.name = name
}

func (d Definition) isEmpty() bool {
	return d.originalName == "" && len(d.children) == 0
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
func ExtractDefinitions(root *sitter.Node, source []byte) *ClassScope {
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
	scope := parseClassScope(root.NamedChild(int(root.NamedChildCount())-1), source)
	scope.Imports = imports
	scope.Package = pack
	return scope
}

func parseClassScope(root *sitter.Node, source []byte) *ClassScope {
	if root.Type() != "class_declaration" {
		return &ClassScope{}
	}
	className := nodeToStr(ParseExpr(root.ChildByFieldName("name"), source, Ctx{}))
	scope := &ClassScope{
		Classes: []*Definition{
			&Definition{
				originalName: className,
				name:         className,
			},
		},
	}

	var node *sitter.Node
	for i := 0; i < int(root.ChildByFieldName("body").NamedChildCount()); i++ {
		node = root.ChildByFieldName("body").NamedChild(i)
		switch node.Type() {
		case "field_declaration":
			name := nodeToStr(ParseExpr(node.ChildByFieldName("declarator").ChildByFieldName("name"), source, Ctx{}))
			scope.Fields = append(scope.Fields, &Definition{
				originalName:   name,
				definitionType: nodeToStr(ParseExpr(node.ChildByFieldName("type"), source, Ctx{})),
				name:           nodeToStr(ParseExpr(node.ChildByFieldName("declarator").ChildByFieldName("name"), source, Ctx{})),
			})
		case "method_declaration", "constructor_declaration":
			name := nodeToStr(ParseExpr(node.ChildByFieldName("name"), source, Ctx{}))
			declaration := &Definition{
				originalName: name,
				name:         name,
			}

			if node.Type() == "method_declaration" {
				declaration.definitionType = nodeToStr(ParseExpr(node.ChildByFieldName("type"), source, Ctx{}))
			} else {
				// A constructor returns itself, so it does not have a type
				declaration.constructor = true
			}

			for _, param := range ParseNode(node.ChildByFieldName("parameters"), source, Ctx{}).(*ast.FieldList).List {
				name := nodeToStr(param.Names[0])
				declaration.parameters = append(declaration.parameters, &Definition{
					originalName:   name,
					definitionType: nodeToStr(param.Type),
					name:           name,
				})
			}

			if node.ChildByFieldName("body") != nil {
				methodScope := parseScope(node.ChildByFieldName("body"), source)
				if !methodScope.isEmpty() {
					declaration.children = append(declaration.children, methodScope)
				}
			}

			scope.Methods = append(scope.Methods, declaration)
		case "class_declaration", "interface_declaration", "enum_declaration":
			other := parseClassScope(node, source)
			scope.Classes = append(scope.Classes, other.Classes...)
			scope.Fields = append(scope.Fields, other.Fields...)
			scope.Methods = append(scope.Methods, other.Methods...)
		}
	}

	return scope
}

func parseScope(root *sitter.Node, source []byte) *Definition {
	def := &Definition{}
	var node *sitter.Node
	for i := 0; i < int(root.NamedChildCount()); i++ {
		node = root.NamedChild(i)
		switch node.Type() {
		case "local_variable_declaration":
			name := nodeToStr(ParseExpr(node.ChildByFieldName("declarator").ChildByFieldName("name"), source, Ctx{}))
			def.children = append(def.children, &Definition{
				originalName:   name,
				definitionType: nodeToStr(ParseExpr(node.ChildByFieldName("type"), source, Ctx{})),
				name:           name,
			})
		case "for_statement", "enhanced_for_statement", "while_statement", "if_statement":
			def.children = append(def.children, parseScope(node, source))
		}
	}
	return def
}

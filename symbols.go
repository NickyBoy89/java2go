package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"

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

// ResolveDefinition resolves a given definition's type, given its class's scope
// as well as its global scope
// It returns true if the definition was successfully resolved, and false otherwise
func ResolveDefinition(definition *Definition, classScope *ClassScope, globalScope *GlobalScope) bool {
	// Look in the class scope first
	if localClassDef := classScope.FindClass(definition.Type()); localClassDef != nil {
		// Every type in the local scope is a reference type, so prefix it with a pointer
		definition.typ = "*" + localClassDef.Name()
		return true

		// Look in the global Scope
	} else if globalDef, in := classScope.Imports[definition.Type()]; in {
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

// A GlobalScope represents a global look of all the packages that make up the
// entirety of the parsed input
type GlobalScope struct {
	// Packages is the full path of the package associated with its definition
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
	files map[string]*ClassScope
}

func (ps PackageScope) String() string {
	return fmt.Sprintf("Package: [%v]", ps.files)
}

// FindClass searches for a class in the given package and returns a scope for it
// the class may be the subclass of another class
func (ps *PackageScope) FindClass(name string) *ClassScope {
	// If the type is a pointer type, look it up without the asterisk
	if name[0] == '*' {
		name = name[1:]
	}
	for _, fileScope := range ps.files {
		for _, className := range fileScope.Classes {
			if className.originalName == name {
				return fileScope
			}
		}
	}
	return nil
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
	return fmt.Sprintf("Classes: %v Fields: %v Methods: %v",
		cs.Classes,
		cs.Fields,
		cs.Methods)
}

// FindMethod looks for a given method by its name in a class definition
func (cs *ClassScope) FindMethod(name string, parameters []string) *Definition {
	for _, method := range cs.Methods {
		if method.originalName == name {
			// If the number of parameters and supplied parameters does not match,
			// reject it immediately
			if len(method.parameters) != len(parameters) {
				break
			}

			var invalid bool
			// Go through all the parameters and check to see if they are valid
			for ind, param := range method.parameters {
				if param.originalType != parameters[ind] {
					invalid = true
					break
				}
			}

			if !invalid {
				return method
			}
		}
	}
	return nil
}

func (cs *ClassScope) FindClass(name string) *Definition {
	if len(name) > 0 && name[0] == '*' {
		name = name[1:]
	}

	for _, class := range cs.Classes {
		if class.originalName == name {
			return class
		}
	}
	return nil
}

// FindField searches for a field by its original name, and returns its definition
// or nil if none was found
func (cs *ClassScope) FindField(name string) *Definition {
	for _, field := range cs.Fields {
		if field.originalName == name {
			return field
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
	// Display type of the object
	typ string
	// Original type of the object
	originalType string
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
	return d.typ
}

func (d Definition) String() string {
	if d.originalName != d.name {
		return fmt.Sprintf("Name: %s (Was %s) Type: %s", d.name, d.originalName, d.typ)

	}
	return fmt.Sprintf("Name: %s Type: %s", d.name, d.typ)
}

// Rename renames a definition for a type so that it can be referenced later with
// the correct name
func (d *Definition) Rename(name string) {
	d.name = name
}

func (d *Definition) ParameterByName(name string) *Definition {
	for _, param := range d.parameters {
		if param.originalName == name {
			return param
		}
	}
	return nil
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

	var public bool
	// Rename the type based on the public/static rules
	if root.NamedChild(0).Type() == "modifiers" {
		for _, modifier := range UnnamedChildren(root.NamedChild(0)) {
			if modifier.Type() == "public" {
				public = true
			}
		}
	}

	className := nodeToStr(ParseExpr(root.ChildByFieldName("name"), source, Ctx{}))
	scope := &ClassScope{
		Classes: []*Definition{
			&Definition{
				originalName: className,
				name:         HandleExportStatus(public, className),
			},
		},
	}

	var node *sitter.Node
	for i := 0; i < int(root.ChildByFieldName("body").NamedChildCount()); i++ {
		node = root.ChildByFieldName("body").NamedChild(i)
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

			name := nodeToStr(ParseExpr(node.ChildByFieldName("declarator").ChildByFieldName("name"), source, Ctx{}))
			scope.Fields = append(scope.Fields, &Definition{
				originalName: name,
				originalType: node.ChildByFieldName("type").Content(source),
				typ:          nodeToStr(ParseExpr(node.ChildByFieldName("type"), source, Ctx{})),
				name:         HandleExportStatus(public, name),
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

			name := nodeToStr(ParseExpr(node.ChildByFieldName("name"), source, Ctx{}))
			declaration := &Definition{
				originalName: name,
				name:         HandleExportStatus(public, name),
			}

			if node.Type() == "method_declaration" {
				declaration.originalType = node.ChildByFieldName("type").Content(source)
				declaration.typ = nodeToStr(ParseExpr(node.ChildByFieldName("type"), source, Ctx{}))
			} else {
				// A constructor returns itself
				declaration.constructor = true
				declaration.typ = name
			}

			for _, parameter := range Children(node.ChildByFieldName("parameters")) {
				parsed := ParseNode(parameter, source, Ctx{}).(*ast.Field)
				name := nodeToStr(parsed.Names[0])
				declaration.parameters = append(declaration.parameters, &Definition{
					originalName: name,
					originalType: parameter.ChildByFieldName("type").Content(source),
					typ:          nodeToStr(parsed.Type),
					name:         name,
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
				originalName: name,
				originalType: node.ChildByFieldName("type").Content(source),
				typ:          nodeToStr(ParseExpr(node.ChildByFieldName("type"), source, Ctx{})),
				name:         name,
			})
		case "for_statement", "enhanced_for_statement", "while_statement", "if_statement":
			def.children = append(def.children, parseScope(node, source))
		}
	}
	return def
}

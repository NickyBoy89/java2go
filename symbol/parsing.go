package symbol

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"

	sitter "github.com/smacker/go-tree-sitter"
)

func nodeToStr(node any) string {
	var s bytes.Buffer
	err := printer.Fprint(&s, token.NewFileSet(), node)
	if err != nil {
		panic(err)
	}
	return s.String()
}

// ExtractDefinitions generates a symbol table for a single class file.
func ExtractDefinitions(root *sitter.Node, source []byte) *FileScope {
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

	return &FileScope{
		Imports:   imports,
		Package:   pack,
		BaseClass: parseClassScope(root.NamedChild(int(root.NamedChildCount())-1), source),
	}
}

func parseClassScope(root *sitter.Node, source []byte) *ClassScope {
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
	scope := &ClassScope{
		Class: &Definition{
			originalName: className,
			name:         HandleExportStatus(public, className),
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

			name := nodeToStr(ParseExpr(node.ChildByFieldName("declarator").ChildByFieldName("name"), source, Ctx{}))
			typeNode := node.ChildByFieldName("type")

			var fieldType string
			// Scoped type identifiers are in a format such as RemotePackage.ClassName
			// To handle this, we remove the RemotePackage part, and depend on the later
			// type resolution to figure things out
			// TODO: Fix this to allow partial lookups, instead of throwing out this information
			if typeNode.Type() == "scoped_type_identifier" {
				fieldType = nodeToStr(ParseExpr(typeNode.NamedChild(int(typeNode.NamedChildCount())-1), source, Ctx{}))
			} else {
				fieldType = nodeToStr(ParseExpr(typeNode, source, Ctx{}))
			}
			scope.Fields = append(scope.Fields, &Definition{
				originalName: name,
				originalType: node.ChildByFieldName("type").Content(source),
				typ:          fieldType,
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
				declaration.Rename(HandleExportStatus(public, "New") + name)
				// A constructor returns itself
				declaration.constructor = true
				declaration.typ = name
			}

			for _, parameter := range Children(node.ChildByFieldName("parameters")) {
				parsed := ParseNode(parameter, source, Ctx{}).(*ast.Field)
				name := nodeToStr(parsed.Names[0])
				if parameter.Type() == "spread_parameter" {
					declaration.parameters = append(declaration.parameters, &Definition{
						originalName: name,
						originalType: parameter.NamedChild(0).Content(source),
						typ:          nodeToStr(parsed.Type),
						name:         name,
					})
				} else {
					declaration.parameters = append(declaration.parameters, &Definition{
						originalName: name,
						originalType: parameter.ChildByFieldName("type").Content(source),
						typ:          nodeToStr(parsed.Type),
						name:         name,
					})
				}
			}

			if node.ChildByFieldName("body") != nil {
				methodScope := parseScope(node.ChildByFieldName("body"), source)
				if !methodScope.isEmpty() {
					declaration.children = append(declaration.children, methodScope.children...)
				}
			}

			scope.Methods = append(scope.Methods, declaration)
		case "class_declaration", "interface_declaration", "enum_declaration":
			other := parseClassScope(node, source)
			// Any subclasses will be renamed to part of their parent class
			other.Class.Rename(scope.Class.Name() + other.Class.Name())
			scope.Subclasses = append(scope.Subclasses, other)
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

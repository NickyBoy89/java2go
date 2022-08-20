package symbol

import (
	"github.com/NickyBoy89/java2go/astutil"
	"github.com/NickyBoy89/java2go/nodeutil"
	sitter "github.com/smacker/go-tree-sitter"
)

// ParseSymbols generates a symbol table for a single class file.
func ParseSymbols(root *sitter.Node, source []byte) *FileScope {
	var filePackage string

	var baseClass *sitter.Node

	imports := make(map[string]string)
	for _, node := range nodeutil.NamedChildrenOf(root) {
		switch node.Type() {
		case "package_declaration":
			filePackage = node.NamedChild(0).Content(source)
		case "import_declaration":
			importedItem := node.NamedChild(0).ChildByFieldName("name").Content(source)
			importPath := node.NamedChild(0).ChildByFieldName("scope").Content(source)

			imports[importedItem] = importPath
		case "class_declaration", "interface_declaration", "enum_declaration":
			baseClass = node
		}
	}

	return &FileScope{
		Imports:   imports,
		Package:   filePackage,
		BaseClass: parseClassScope(baseClass, source),
	}
}

func parseClassScope(root *sitter.Node, source []byte) *ClassScope {
	var public bool
	// Rename the type based on the public/static rules
	if root.NamedChild(0).Type() == "modifiers" {
		for _, node := range nodeutil.UnnamedChildrenOf(root.NamedChild(0)) {
			if node.Type() == "public" {
				public = true
			}
		}
	}

	nodeutil.AssertTypeIs(root.ChildByFieldName("name"), "identifier")

	// Parse the main class in the file

	className := root.ChildByFieldName("name").Content(source)
	scope := &ClassScope{
		Class: &Definition{
			OriginalName: className,
			Name:         HandleExportStatus(public, className),
		},
	}

	// Parse the body of the class

	for _, node := range nodeutil.NamedChildrenOf(root.ChildByFieldName("body")) {

		switch node.Type() {
		case "field_declaration":
			var public bool
			// Rename the type based on the public/static rules
			if node.NamedChild(0).Type() == "modifiers" {
				for _, modifier := range nodeutil.UnnamedChildrenOf(node.NamedChild(0)) {
					if modifier.Type() == "public" {
						public = true
					}
				}
			}

			fieldNameNode := node.ChildByFieldName("declarator").ChildByFieldName("name")

			nodeutil.AssertTypeIs(fieldNameNode, "identifier")

			// TODO: Scoped type identifiers are in a format such as RemotePackage.ClassName
			// To handle this, we remove the RemotePackage part, and depend on the later
			// type resolution to figure things out

			// The node that the field's type comes from
			typeNode := node.ChildByFieldName("type")

			// If the field is being assigned to a value
			if typeNode.Type() == "scoped_type_identifier" {
				typeNode = typeNode.NamedChild(int(typeNode.NamedChildCount()) - 1)
			}

			// The converted name and type of the field
			fieldName := fieldNameNode.Content(source)
			fieldType := nodeToStr(astutil.ParseType(typeNode, source))

			scope.Fields = append(scope.Fields, &Definition{
				Name:         HandleExportStatus(public, fieldName),
				OriginalName: fieldName,
				Type:         fieldType,
				OriginalType: typeNode.Content(source),
			})
		case "method_declaration", "constructor_declaration":
			var public bool
			// Rename the type based on the public/static rules
			if node.NamedChild(0).Type() == "modifiers" {
				for _, modifier := range nodeutil.UnnamedChildrenOf(node.NamedChild(0)) {
					if modifier.Type() == "public" {
						public = true
					}
				}
			}

			nodeutil.AssertTypeIs(node.ChildByFieldName("name"), "identifier")

			name := node.ChildByFieldName("name").Content(source)
			declaration := &Definition{
				Name:         HandleExportStatus(public, name),
				OriginalName: name,
			}

			if node.Type() == "method_declaration" {
				declaration.Type = nodeToStr(astutil.ParseType(node.ChildByFieldName("type"), source))
				declaration.OriginalType = node.ChildByFieldName("type").Content(source)
			} else {
				// A constructor declaration returns the type being constructed

				// Rename the constructor with "New" + name of type
				declaration.Rename(HandleExportStatus(public, "New") + name)
				declaration.Constructor = true

				// There is no original type, and the constructor returns the name of
				// the new type
				declaration.Type = name
			}

			// Parse the parameters

			for _, parameter := range nodeutil.NamedChildrenOf(node.ChildByFieldName("parameters")) {

				var paramName string
				var paramType *sitter.Node

				// If this is a spread parameter, then it will be in the format:
				// (type) (variable_declarator name: (name))
				if parameter.Type() == "spread_parameter" {
					paramName = parameter.NamedChild(1).ChildByFieldName("name").Content(source)
					paramType = parameter.NamedChild(0)
				} else {
					paramName = parameter.ChildByFieldName("name").Content(source)
					paramType = parameter.ChildByFieldName("type")
				}

				declaration.Parameters = append(declaration.Parameters, &Definition{
					Name:         paramName,
					OriginalName: paramName,
					Type:         nodeToStr(astutil.ParseType(paramType, source)),
					OriginalType: paramType.Content(source),
				})
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

func parseScope(root *sitter.Node, source []byte) *Definition {
	def := &Definition{}
	for _, node := range nodeutil.NamedChildrenOf(root) {
		switch node.Type() {
		case "local_variable_declaration":
			/*
				name := nodeToStr(ParseExpr(node.ChildByFieldName("declarator").ChildByFieldName("name"), source, Ctx{}))
				def.Children = append(def.Children, &symbol.Definition{
					OriginalName: name,
					OriginalType: node.ChildByFieldName("type").Content(source),
					Type:         nodeToStr(ParseExpr(node.ChildByFieldName("type"), source, Ctx{})),
					Name:         name,
				})
			*/
		case "for_statement", "enhanced_for_statement", "while_statement", "if_statement":
			def.Children = append(def.Children, parseScope(node, source))
		}
	}
	return def
}

package main

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"

	sitter "github.com/smacker/go-tree-sitter"
)

// ClassScope contains the global and local scopes for a single file
// if a file contains multiply classes, all the definitions are folded into
// one ClassScope
type ClassScope struct {
	Classes []*Definition
	Fields  []*Definition
	Methods []*Definition
}

// A Definition contains information about a single entry
type Definition struct {
	// The original java name
	OriginalName string
	// The display name, may be different from the original name
	Name string
	// Type of the object
	Type string
	// If the definition is a constructor
	Constructor bool
	// If the object is a function, it has parameters
	Parameters []*Definition
	// Children of the declaration, if the declaration is a scope
	Children []*Definition
}

// Rename renames a definition for a type so that it can be referenced later with
// the correct name
func (d *Definition) Rename(name string) {
	d.Name = name
}

func (d Definition) isEmpty() bool {
	return d.OriginalName == "" && len(d.Children) == 0
}

func nodeToStr(node any) string {
	var s bytes.Buffer
	err := printer.Fprint(&s, token.NewFileSet(), node)
	if err != nil {
		panic(err)
	}
	return s.String()
}

// ExtractDefinitions generates a symbol table containing all the definitions
// for a single input file
func ExtractDefinitions(root *sitter.Node, source []byte) *ClassScope {
	return parseClassScope(root.NamedChild(0), source)
}

func parseClassScope(root *sitter.Node, source []byte) *ClassScope {
	className := nodeToStr(ParseExpr(root.ChildByFieldName("name"), source, Ctx{}))
	scope := &ClassScope{
		Classes: []*Definition{
			&Definition{
				OriginalName: className,
				Name:         className,
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
				OriginalName: name,
				Type:         nodeToStr(ParseExpr(node.ChildByFieldName("type"), source, Ctx{})),
				Name:         nodeToStr(ParseExpr(node.ChildByFieldName("declarator").ChildByFieldName("name"), source, Ctx{})),
			})
		case "method_declaration", "constructor_declaration":
			name := nodeToStr(ParseExpr(node.ChildByFieldName("name"), source, Ctx{}))
			declaration := &Definition{
				OriginalName: name,
				Name:         name,
			}

			if node.Type() == "method_declaration" {
				declaration.Type = nodeToStr(ParseExpr(node.ChildByFieldName("type"), source, Ctx{}))
			} else {
				// A constructor returns itself, so it does not have a type
				declaration.Constructor = true
			}

			for _, param := range ParseNode(node.ChildByFieldName("parameters"), source, Ctx{}).(*ast.FieldList).List {
				name := nodeToStr(param.Names[0])
				declaration.Parameters = append(declaration.Parameters, &Definition{
					OriginalName: name,
					Type:         nodeToStr(param.Type),
					Name:         name,
				})
			}

			methodScope := parseScope(node.ChildByFieldName("body"), source)
			if !methodScope.isEmpty() {
				declaration.Children = append(declaration.Children, methodScope)
			}

			scope.Methods = append(scope.Methods, declaration)
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
			def.Children = append(def.Children, &Definition{
				OriginalName: name,
				Type:         nodeToStr(ParseExpr(node.ChildByFieldName("type"), source, Ctx{})),
				Name:         name,
			})
		case "for_statement", "enhanced_for_statement", "while_statement", "if_statement":
			def.Children = append(def.Children, parseScope(node, source))
		}
	}
	return def
}

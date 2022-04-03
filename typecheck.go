package main

import (
	"bytes"
	"errors"
	"go/ast"
	"go/printer"
	"go/token"
	"io"

	sitter "github.com/smacker/go-tree-sitter"
)

// TypeInformation stores the type information of all known variables in the
// current declaration. This includes, but is not limited to: Local variables,
// return types, parameter types
type TypeInformation struct {
	types map[string]string
}

type TypeDecl struct {
	Type string
	Name string
}

func nodeToStr(node any) string {
	var s bytes.Buffer
	err := printer.Fprint(&s, token.NewFileSet(), node)
	if err != nil {
		panic(err)
	}
	return s.String()
}

func walkTypeInfo(node *sitter.Node, source []byte) []TypeDecl {
	types := []TypeDecl{}
	switch node.Type() {
	case "field_declaration", "local_variable_declaration":
		types = append(types, TypeDecl{
			Type: nodeToStr(ParseExpr(node.ChildByFieldName("type"), source, Ctx{})),
			Name: nodeToStr(ParseExpr(node.ChildByFieldName("declarator").ChildByFieldName("name"), source, Ctx{})),
		})
	case "method_declaration", "constructor_declaration":
		if node.Type() == "method_declaration" {
			// Declare the method, and its return type
			types = append(types, TypeDecl{
				Type: nodeToStr(ParseExpr(node.ChildByFieldName("type"), source, Ctx{})),
				Name: nodeToStr(ParseExpr(node.ChildByFieldName("name"), source, Ctx{})),
			})
		}

		// Declare all the parameters within a method
		for _, param := range ParseNode(node.ChildByFieldName("parameters"), source, Ctx{}).(*ast.FieldList).List {
			types = append(types, TypeDecl{
				Type: nodeToStr(param.Type),
				Name: nodeToStr(param.Names[0]),
			})
		}
	}
	return types
}

func ExtractTypeInformation(root *sitter.Node, source []byte) (TypeInformation, error) {
	info := TypeInformation{types: make(map[string]string)}

	iter := sitter.NewNamedIterator(root, sitter.DFSMode)

	var curNode *sitter.Node
	var err error

	for {
		curNode, err = iter.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return info, err
		}

		for _, decl := range walkTypeInfo(curNode, source) {
			info.types[decl.Name] = decl.Type
		}
	}

	return info, nil
}

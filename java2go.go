package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	toParse := os.Args[1:]
	for _, fileName := range toParse {
		contents, err := os.ReadFile(fileName)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatalf("Error reading file")
		}
		out := make(map[string]interface{})
		err = json.Unmarshal(contents, &out)
		if err != nil {
			log.WithFields(log.Fields{
				"file":  fileName,
				"error": err,
			}).Fatalf("Error unmarshaling JSON")
		}

		err = printer.Fprint(os.Stdout, token.NewFileSet(), ParseType(out))
		if err != nil {
			panic(err)
		}
	}
}

func ParseType(input map[string]interface{}) ast.Node {
	// If the JSON indicates a custom type
	if _, in := input["Name"]; in {
		// Contents is a convenience variable for accessing the contents of custom
		// types
		contents := input["Contents"].(map[string]interface{})
		switch input["Name"] {
		case "<class 'javalang.tree.CompilationUnit'>":
			astFile := &ast.File{
				Name:  &ast.Ident{Name: "main"},
				Decls: []ast.Decl{},
			}
			if contents["package"] != nil {
				astFile.Name = &ast.Ident{Name: contents["package"].(string)}
			}
			for _, item := range contents["types"].([]interface{}) {
				astFile.Decls = append(astFile.Decls, ParseType(item.(map[string]interface{})).(ast.Decl))
			}
			return astFile
		case "<class 'javalang.tree.ClassDeclaration'>":
			createdStruct := GenStruct(contents["name"].(string), &ast.FieldList{List: []*ast.Field{}})
			for _, bodyItem := range contents["body"].([]interface{}) {
				item := bodyItem.(map[string]interface{})
				if _, in := item["Name"]; in && item["Name"] == "<class 'javalang.tree.FieldDeclaration'>" {
					createdStruct.(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields.List = append(createdStruct.(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields.List, ParseType(item).(*ast.Field))
				}
			}
			return createdStruct
		case "<class 'javalang.tree.FieldDeclaration'>":
			createdField := &ast.Field{
				Names: []*ast.Ident{},
				Type:  ParseType(contents["type"].(map[string]interface{})).(*ast.Ident),
			}
			for _, decl := range contents["declarators"].([]interface{}) {
				createdField.Names = append(createdField.Names, ParseType(decl.(map[string]interface{})).(*ast.Ident))
			}
			return createdField
		case "<class 'javalang.tree.BasicType'>":
			return &ast.Ident{Name: contents["name"].(string)}
		case "<class 'javalang.tree.VariableDeclarator'>":
			return &ast.Ident{Name: contents["name"].(string)}
		default:
			panic(fmt.Sprintf("Unknown type: %v", input["Name"]))
		}
	}
	return nil
}

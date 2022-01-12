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
		switch input["Name"] {
		case "<class 'javalang.tree.CompilationUnit'>":
			return &ast.File{
				Name: &ast.Ident{Name: "main"},
				Decls: []ast.Decl{
					GenStruct("Test", WithFields(Fields{
						"value": "int",
					})),
					GenFunction("NewTest", nil, WithFields(Fields{
						"val": "int",
					}), &ast.FieldList{
						List: []*ast.Field{
							&ast.Field{
								Type: &ast.StarExpr{
									X: &ast.Ident{Name: "Test"},
								},
							},
						},
					}, &ast.BlockStmt{
						List: []ast.Stmt{},
					}),
				},
			}
			return ParseType(input["Contents"].(map[string]interface{}))
		default:
			panic(fmt.Sprintf("Unknown type: %v", input["Name"]))
		}
	}
	return nil
}

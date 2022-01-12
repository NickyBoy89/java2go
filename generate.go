package main

import (
	"go/ast"
	"go/token"
)

func GenStruct(structName string, structFields *ast.FieldList) ast.Decl {
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{
					Name: structName,
				},
				Type: &ast.StructType{
					Fields: structFields,
				},
			},
		},
	}
}

type Fields map[string]string

func WithFields(f Fields) *ast.FieldList {
	result := &ast.FieldList{
		List: []*ast.Field{},
	}
	for fieldName, fieldType := range f {
		result.List = append(result.List, &ast.Field{
			Names: []*ast.Ident{
				&ast.Ident{
					Name: fieldName,
				},
			},
			Type: &ast.Ident{
				Name: fieldType,
			},
		})
	}
	return result
}

func GenFunction(functionName string, receiver *ast.FieldList, parameters, results *ast.FieldList, body *ast.BlockStmt) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: &ast.Ident{Name: functionName},
		Recv: receiver,
		Type: &ast.FuncType{
			Params:  parameters,
			Results: results,
		},
		Body: body,
	}
}

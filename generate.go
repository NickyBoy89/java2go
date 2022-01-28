package main

import (
	"go/ast"
	"go/token"
	"unicode"
)

// ShortName returns the short-name representation of a class's name for use
// in methods and construtors
// Ex: Test -> ts
func ShortName(longName string) string {
	if len(longName) == 0 {
		return ""
	}
	return string(unicode.ToLower(rune(longName[0]))) + string(unicode.ToLower(rune(longName[len(longName)-1])))
}

// GenStruct is a utility method for generating the ast representation of
// a struct, given its name and fields
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

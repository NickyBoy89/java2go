package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"unicode"
)

var tokens = map[string]token.Token{
	"+": token.ADD,
	"-": token.SUB,
	"*": token.MUL,
	"/": token.QUO,
	"%": token.REM,

	"&": token.AND,
	"|": token.OR,
	"^": token.XOR,
	// Java bitwise complement (~)
	"~":  token.XOR,
	"<<": token.SHL,
	">>": token.SHR,
	"&^": token.AND_NOT,

	"+=": token.ADD_ASSIGN,
	"-=": token.SUB_ASSIGN,
	"*=": token.MUL_ASSIGN,
	"/=": token.QUO_ASSIGN,
	"%=": token.REM_ASSIGN,

	"&=":  token.AND_ASSIGN,
	"|=":  token.OR_ASSIGN,
	"^=":  token.XOR_ASSIGN,
	"<<=": token.SHL_ASSIGN,
	">>=": token.SHR_ASSIGN,
	"&^=": token.AND_NOT_ASSIGN,

	"&&": token.LAND,
	"||": token.LOR,
	"++": token.INC,
	"--": token.DEC,

	"==": token.EQL,
	"<":  token.LSS,
	">":  token.GTR,
	"=":  token.ASSIGN,
	"!":  token.NOT,

	"!=":  token.NEQ,
	"<=":  token.LEQ,
	">=":  token.GEQ,
	":=":  token.DEFINE,
	"...": token.ELLIPSIS,
}

// Maps a token's representation to its token, e.g. "+" -> token.ADD
func StrToToken(input string) token.Token {
	if outToken, known := tokens[input]; known {
		return outToken
	}
	panic(fmt.Errorf("Unknown token for [%v]", input))
}

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

func genType(remaining []string) ast.Expr {
	if len(remaining) == 1 {
		return &ast.UnaryExpr{
			Op: token.TILDE,
			X:  &ast.Ident{Name: remaining[0]},
		}
	}
	return &ast.BinaryExpr{
		X:  genType(remaining[1:]),
		Op: token.OR,
		Y:  genType(remaining[:1]),
	}
}

func GenTypeInterface(name string, types []string) ast.Decl {
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{Name: name},
				Type: &ast.InterfaceType{
					Methods: &ast.FieldList{
						List: []*ast.Field{
							&ast.Field{
								Type: genType(types),
							},
						},
					},
				},
			},
		},
	}
}

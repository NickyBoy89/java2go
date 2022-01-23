package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"unicode"

	log "github.com/sirupsen/logrus"
)

func oldmain() {
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

		err = printer.Fprint(os.Stdout, token.NewFileSet(), ParseType(out, ""))
		if err != nil {
			panic(err)
		}
	}
}

func StringToToken(str string) token.Token {
	switch str {
	case "==":
		return token.EQL
	case "<=":
		return token.LEQ
	case "++":
		return token.INC
	case "--":
		return token.DEC
	case ">":
		return token.GTR
	case "<":
		return token.LSS
	case "!=":
		return token.NEQ
	case "+":
		return token.ADD
	case "*":
		return token.MUL
	case "-":
		return token.SUB
	default:
		panic(fmt.Sprintf("Unknown token conversion from %v", str))
	}
}

func ShortName(longName string) string {
	if len(longName) == 0 {
		return ""
	}
	return string(unicode.ToLower(rune(longName[0]))) + string(unicode.ToLower(rune(longName[len(longName)-1])))
}

func AsStringSlice(in interface{}) []string {
	result := make([]string, len(in.([]interface{})))
	for ind, item := range in.([]interface{}) {
		result[ind] = item.(string)
	}
	return result
}

func AssertMapArray(ary interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, len(ary.([]interface{})))
	for ind, item := range ary.([]interface{}) {
		result[ind] = item.(map[string]interface{})
	}
	return result
}

func Contains(ary []string, target string) bool {
	for _, item := range ary {
		if item == target {
			return true
		}
	}
	return false
}

func ParseType(input map[string]interface{}, className string) ast.Node {
	// If the JSON indicates a custom type
	if _, in := input["Name"]; in {
		// Contents is a convenience variable for accessing the contents of custom
		// types
		contents := input["Contents"].(map[string]interface{})

		postFix := contents["postfix_operators"]
		if postFix != nil && len(postFix.([]interface{})) > 0 {
			for _, op := range postFix.([]interface{}) {
				switch op.(string) {
				case "++", "--":
					delete(input["Contents"].(map[string]interface{}), "postfix_operators")
					return &ast.IncDecStmt{X: ParseType(input, className).(ast.Expr), Tok: StringToToken(op.(string))}
				}
			}
		}

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
				if item.(map[string]interface{})["Name"] == "<class 'javalang.tree.ClassDeclaration'>" {
					astFile.Decls = append(astFile.Decls, ParseClass(item.(map[string]interface{}))...)
					continue
				}
				astFile.Decls = append(astFile.Decls, ParseType(item.(map[string]interface{}), className).(ast.Decl))
			}
			return astFile
		case "<class 'javalang.tree.ClassDeclaration'>":
			return &ast.File{Decls: ParseClass(input)}
		case "<class 'javalang.tree.ConstructorDeclaration'>", "<class 'javalang.tree.MethodDeclaration'>": // Constructors, Methods, and Functions all use the same underlying structure
			functionParams := &ast.FieldList{}
			for _, param := range contents["parameters"].([]interface{}) {
				functionParams.List = append(functionParams.List, ParseType(param.(map[string]interface{}), className).(*ast.Field))
			}
			functionBody := &ast.BlockStmt{}
			for _, body := range contents["body"].([]interface{}) {
				functionBody.List = append(functionBody.List, ParseType(body.(map[string]interface{}), className).(ast.Stmt))
			}
			switch input["Name"] {
			case "<class 'javalang.tree.MethodDeclaration'>":
				methodName := contents["name"].(string)
				if Contains(AsStringSlice(contents["modifiers"]), "public") {
					methodName = string(unicode.ToUpper(rune(methodName[0]))) + methodName[1:]
				}
				return &ast.FuncDecl{
					Name: &ast.Ident{Name: methodName},
					Recv: &ast.FieldList{List: []*ast.Field{&ast.Field{Names: []*ast.Ident{&ast.Ident{Name: ShortName(className)}}, Type: &ast.StarExpr{X: &ast.Ident{Name: className}}}}},
					Type: &ast.FuncType{
						Params:  functionParams,
						Results: &ast.FieldList{},
					},
					Body: functionBody,
				}
			case "<class 'javalang.tree.ConstructorDeclaration'>":
				functionBody.List = append([]ast.Stmt{&ast.AssignStmt{Lhs: []ast.Expr{&ast.Ident{Name: ShortName(className)}}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.Ident{Name: "new"}, Args: []ast.Expr{&ast.Ident{Name: className}}}}}}, functionBody.List...)
				functionBody.List = append(functionBody.List, &ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: ShortName(className)}}})
				return GenFunction("New"+contents["name"].(string), nil, functionParams, &ast.FieldList{}, functionBody)
			}
		case "<class 'javalang.tree.FieldDeclaration'>":
			createdField := &ast.Field{
				Names: []*ast.Ident{},
				Type:  ParseType(contents["type"].(map[string]interface{}), className).(ast.Expr),
			}
			for _, decl := range contents["declarators"].([]interface{}) {
				createdField.Names = append(createdField.Names, ParseType(decl.(map[string]interface{}), className).(*ast.Ident))
			}
			return createdField
		case "<class 'javalang.tree.MethodInvocation'>":
			params := []ast.Expr{}
			for _, expr := range contents["arguments"].([]interface{}) {
				params = append(params, ParseType(expr.(map[string]interface{}), className).(ast.Expr))
			}
			return &ast.CallExpr{Fun: &ast.Ident{Name: contents["member"].(string)}, Args: params}
		case "<class 'javalang.tree.BasicType'>":
			return &ast.Ident{Name: contents["name"].(string)}
		case "<class 'javalang.tree.ReferenceType'>":
			return &ast.StarExpr{X: &ast.Ident{Name: contents["name"].(string)}}
		case "<class 'javalang.tree.FormalParameter'>":
			return &ast.Field{
				Names: []*ast.Ident{&ast.Ident{Name: contents["name"].(string)}},
				Type:  ParseType(contents["type"].(map[string]interface{}), className).(ast.Expr),
			}
		case "<class 'javalang.tree.VariableDeclarator'>":
			return &ast.Ident{Name: contents["name"].(string)}
		case "<class 'javalang.tree.StatementExpression'>":
			returnStmt := ParseType(contents["expression"].(map[string]interface{}), className)
			if stmt, ok := returnStmt.(ast.Stmt); ok {
				return stmt
			}
			return &ast.ExprStmt{X: returnStmt.(ast.Expr)}
		case "<class 'javalang.tree.Assignment'>":
			return &ast.AssignStmt{Lhs: []ast.Expr{ParseType(contents["expressionl"].(map[string]interface{}), className).(ast.Expr)}, Tok: token.ASSIGN, Rhs: []ast.Expr{&ast.Ident{Name: "val"}}}
		case "<class 'javalang.tree.LocalVariableDeclaration'>", "<class 'javalang.tree.VariableDeclaration'>":
			var varNames, varValues []ast.Expr
			for _, dec := range contents["declarators"].([]interface{}) {
				name, value := ParseVariableDeclaration(dec.(map[string]interface{}))
				varNames = append(varNames, name)
				varValues = append(varValues, value)
			}
			return &ast.AssignStmt{Lhs: varNames, Tok: token.DEFINE, Rhs: varValues}
		case "<class 'javalang.tree.This'>":
			sel := ParseType(contents["selectors"].([]interface{})[0].(map[string]interface{}), className)

			if _, ok := sel.(*ast.CallExpr); ok {
				return &ast.SelectorExpr{
					X:   sel.(ast.Expr),
					Sel: &ast.Ident{Name: "asdsd"},
				}
			}

			selector := &ast.SelectorExpr{
				X:   &ast.Ident{Name: ShortName(className)},
				Sel: sel.(*ast.Ident),
			}
			return selector
		case "<class 'javalang.tree.MemberReference'>":
			return &ast.Ident{Name: contents["member"].(string)}
		case "<class 'javalang.tree.ReturnStatement'>":
			return &ast.ReturnStmt{Results: []ast.Expr{ParseType(contents["expression"].(map[string]interface{}), className).(ast.Expr)}}
		case "<class 'javalang.tree.IfStatement'>":
			ifStmt := &ast.IfStmt{
				Cond: ParseType(contents["condition"].(map[string]interface{}), className).(ast.Expr),
				Body: ParseType(contents["then_statement"].(map[string]interface{}), className).(*ast.BlockStmt),
			}

			if contents["else_statement"] != nil {
				ifStmt.Else = ParseType(contents["else_statement"].(map[string]interface{}), className).(ast.Stmt)
			}

			return ifStmt
		case "<class 'javalang.tree.ForStatement'>":
			control := ParseType(contents["control"].(map[string]interface{}), className).(*ast.ForStmt)
			return &ast.ForStmt{
				Init: control.Init,
				Cond: control.Cond,
				Post: control.Post,
				Body: ParseType(contents["body"].(map[string]interface{}), className).(*ast.BlockStmt),
			}
		case "<class 'javalang.tree.ForControl'>":
			return &ast.ForStmt{
				Init: ParseType(contents["init"].(map[string]interface{}), className).(ast.Stmt),
				Cond: ParseType(contents["condition"].(map[string]interface{}), className).(ast.Expr),
				Post: ParseType(contents["update"].([]interface{})[0].(map[string]interface{}), className).(ast.Stmt),
			}
		case "<class 'javalang.tree.BlockStatement'>":
			block := &ast.BlockStmt{List: []ast.Stmt{}}
			for _, stmt := range contents["statements"].([]interface{}) {
				block.List = append(block.List, ParseType(stmt.(map[string]interface{}), className).(ast.Stmt))
			}
			return block
		case "<class 'javalang.tree.ThrowStatement'>":
			return &ast.ExprStmt{X: &ast.CallExpr{Fun: &ast.Ident{Name: "panic"}, Args: []ast.Expr{ParseType(contents["expression"].(map[string]interface{}), className).(ast.Expr)}}}
		case "<class 'javalang.tree.ClassCreator'>":
			args := []ast.Expr{}
			for _, arg := range contents["arguments"].([]interface{}) {
				args = append(args, ParseType(arg.(map[string]interface{}), className).(ast.Expr))
			}
			return &ast.CallExpr{Fun: ParseType(contents["type"].(map[string]interface{}), className).(ast.Expr), Args: args}
		case "<class 'javalang.tree.ArrayCreator'>":
			return &ast.CompositeLit{Type: &ast.ArrayType{Elt: &ast.Ident{Name: "int"}}, Elts: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "99"}}}
		case "<class 'javalang.tree.ArrayInitializer'>":
			lits := []ast.Expr{}
			for _, init := range contents["initializers"].([]interface{}) {
				lits = append(lits, ParseType(init.(map[string]interface{}), className).(ast.Expr))
			}
			return &ast.CompositeLit{Type: &ast.ArrayType{Elt: &ast.Ident{Name: "int"}}, Elts: lits}
		case "<class 'javalang.tree.BinaryOperation'>":
			return &ast.BinaryExpr{X: ParseType(contents["operandl"].(map[string]interface{}), className).(ast.Expr), Op: StringToToken(contents["operator"].(string)), Y: ParseType(contents["operandr"].(map[string]interface{}), className).(ast.Expr)}
		case "<class 'javalang.tree.Literal'>":
			return &ast.Ident{Name: contents["value"].(string)}
		default:
			panic(fmt.Sprintf("Unknown type: %v", input["Name"]))
		}
	}
	return nil
}

// ParseClass is separate because all the parsed methods and functions are
// declared at the same level as the struct declaration, as opposed to inside
// the score of the original class
func ParseClass(input map[string]interface{}) []ast.Decl {
	nodes := []ast.Decl{}
	contents := input["Contents"].(map[string]interface{})
	className := contents["name"].(string)
	createdStruct := GenStruct(contents["name"].(string), &ast.FieldList{List: []*ast.Field{}})
	for _, bodyItem := range contents["body"].([]interface{}) {
		item := bodyItem.(map[string]interface{})
		if _, in := item["Name"]; in {
			switch item["Name"] {
			case "<class 'javalang.tree.FieldDeclaration'>":
				createdStruct.(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields.List = append(createdStruct.(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields.List, ParseType(item, className).(*ast.Field))
			default:
				parsed := ParseType(item, className)
				switch parsed.(type) {
				case *ast.File:
					nodes = append(nodes, parsed.(*ast.File).Decls...)
				default:
					nodes = append(nodes, parsed.(ast.Decl))
				}
			}
		}
	}
	return append([]ast.Decl{createdStruct}, nodes...)
}

func ParseVariableDeclaration(input map[string]interface{}) (ast.Expr, ast.Expr) {
	contents := input["Contents"].(map[string]interface{})
	switch input["Name"].(string) {
	case "<class 'javalang.tree.VariableDeclarator'>":
		return &ast.Ident{Name: contents["name"].(string)}, ParseType(contents["initializer"].(map[string]interface{}), "").(ast.Expr)
	default:
		panic(fmt.Sprintf("Unknown declaration type: %v", contents))
	}
}

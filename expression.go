package main

import (
	"fmt"
	"go/ast"
	"go/token"

	sitter "github.com/smacker/go-tree-sitter"
)

func ParseExpr(node *sitter.Node, source []byte, ctx Ctx) ast.Expr {
	switch node.Type() {
	case "assignment_expression":
		// A simple variable assignment, ex: `name = value`

		// Stores all the assignments if the statement is a multiple-expression
		assignments := []ast.Stmt{}

		names := []ast.Expr{}
		values := []ast.Expr{}
		for i := 0; i < int(node.NamedChildCount())-1; i++ {
			// Rewrite double assignments, e.g. `variable1 = variable2 = 1` to
			// `variable2 = 1`
			// `variable1 = variable2`
			if node.NamedChild(i+1).Type() == "assignment_expression" {
				// If a value is a multiple assignment, add that assignment before the
				// current one, and add the left side of the value to the current line
				otherAssign := ParseNode(node.NamedChild(i+1), source, ctx)

				if otherStmts, ok := otherAssign.([]ast.Stmt); ok {
					assignments = append(assignments, otherStmts...)
				} else {
					assignments = append(assignments, otherAssign.(ast.Stmt))
				}

				// Assign the value to the latest Lhs expression
				assignments = append(assignments, &ast.AssignStmt{
					Lhs: []ast.Expr{ParseExpr(node.NamedChild(i), source, ctx)},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{assignments[len(assignments)-1].(*ast.AssignStmt).Lhs[0]},
				})
			} else {
				names = append(names, ParseExpr(node.NamedChild(i), source, ctx))
				values = append(values, ParseExpr(node.NamedChild(i+1), source, ctx))
			}
		}

		/*
			if len(assignments) > 0 {
				return assignments
			}
		*/

		return &ast.AssignStmt{Lhs: names, Tok: token.ASSIGN, Rhs: values}
	case "method_invocation":
		// Class methods are called with three nodes, the selector, the identifier,
		// and the list of arguments, so that they form the shape
		// `selector.identifier(list of arguments)`
		// Static methods are only called with the identifier and list of args
		// They look like: `identifier(args)`

		switch node.NamedChildCount() {
		case 3: // Invoking a normal class method
			// This is of the form X.Sel(Args)
			return &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ParseExpr(node.NamedChild(0), source, ctx),
					Sel: ParseExpr(node.NamedChild(1), source, ctx).(*ast.Ident),
				},
				Args: ParseNode(node.NamedChild(2), source, ctx).([]ast.Expr),
			}
		case 2: // Invoking a static method
			return &ast.CallExpr{
				// The name is in the wrong place, so use the selector as the name
				Fun:  ParseExpr(node.NamedChild(0), source, ctx),
				Args: ParseNode(node.NamedChild(1), source, ctx).([]ast.Expr),
			}
		case 4: // Calling a method from the parent function
			// NOTE: Fix this to add the entire logic of having an outer function
			return &ast.BadExpr{}
		default:
			panic(fmt.Sprintf("Calling method with unknown number of args: %v", node.NamedChildCount()))
		}
	case "object_creation_expression":
		return &ast.CallExpr{
			// All object creations are usually done by calling the constructor
			//function, which is generated as `"New" + className`
			Fun:  &ast.Ident{Name: "New" + ParseExpr(node.NamedChild(0), source, ctx).(*ast.StarExpr).X.(*ast.Ident).Name},
			Args: ParseNode(node.NamedChild(1), source, ctx).([]ast.Expr),
		}
	case "array_creation_expression":
		// The type of the array
		arrayType := ParseExpr(node.NamedChild(0), source, ctx)
		// The dimensions of the array, which Golang only supports defining one at
		// a time with the use of the builtin `make`
		dimensions := []ast.Expr{&ast.ArrayType{Elt: arrayType}}
		for _, c := range Children(node)[1:] {
			if c.Type() == "dimensions_expr" {
				dimensions = append(dimensions, ParseExpr(c, source, ctx))
			}
		}

		return &ast.CallExpr{
			Fun:  &ast.Ident{Name: "make"},
			Args: dimensions,
		}
	case "instanceof_expression":
		return &ast.BadExpr{}
	case "dimensions_expr":
		return &ast.Ident{Name: node.NamedChild(0).Content(source)}
	case "binary_expression":
		return &ast.BinaryExpr{
			X:  ParseExpr(node.Child(0), source, ctx),
			Op: StrToToken(node.Child(1).Content(source)),
			Y:  ParseExpr(node.Child(2), source, ctx),
		}
	case "unary_expression":
		return &ast.UnaryExpr{
			Op: StrToToken(node.Child(0).Content(source)),
			X:  ParseExpr(node.Child(1), source, ctx),
		}
	case "parenthesized_expression":
		return &ast.ParenExpr{
			X: ParseExpr(node.NamedChild(0), source, ctx),
		}
	case "ternary_expression":
		// Ternary expressions are replaced with a function that takes in the
		// condition, and returns one of the two values, depending on the condition

		args := []ast.Expr{}
		for _, c := range Children(node) {
			args = append(args, ParseExpr(c, source, ctx))
		}
		return &ast.CallExpr{
			Fun:  &ast.Ident{Name: "ternary"},
			Args: args,
		}
	case "cast_expression":
		return &ast.TypeAssertExpr{
			X:    ParseExpr(node.NamedChild(1), source, ctx),
			Type: ParseExpr(node.NamedChild(0), source, ctx),
		}
	case "field_access":
		return &ast.SelectorExpr{
			X:   ParseExpr(node.NamedChild(0), source, ctx),
			Sel: ParseExpr(node.NamedChild(1), source, ctx).(*ast.Ident),
		}
	case "array_access":
		return &ast.IndexExpr{
			X:     ParseExpr(node.NamedChild(0), source, ctx),
			Index: ParseExpr(node.NamedChild(1), source, ctx),
		}
	case "scoped_identifier":
		return ParseExpr(node.NamedChild(0), source, ctx)
	case "this":
		return &ast.Ident{Name: ShortName(ctx.className)}
	case "identifier":
		return &ast.Ident{Name: node.Content(source)}
	case "integral_type":
		return &ast.Ident{Name: node.Content(source)}
	case "floating_point_type": // Can be either `float` or `double`
		return &ast.Ident{Name: node.Content(source)}
	case "void_type":
		return &ast.Ident{}
	case "boolean_type":
		return &ast.Ident{Name: node.Content(source)}
	case "generic_type":
		// A generic type is any type that is of the form GenericType<T>
		return &ast.Ident{Name: node.NamedChild(0).Content(source)}
	case "array_type":
		return &ast.ArrayType{Elt: ParseExpr(node.NamedChild(0), source, ctx)}
	case "type_identifier": // Any reference type
		return &ast.StarExpr{
			X: &ast.Ident{Name: node.Content(source)},
		}
	case "null_literal":
		return &ast.Ident{Name: "nil"}
	case "decimal_integer_literal":
		return &ast.Ident{Name: node.Content(source)}
	case "decimal_floating_point_literal":
		// This is something like 1.3D
		return &ast.Ident{Name: node.Content(source)}
	case "string_literal":
		return &ast.Ident{Name: node.Content(source)}
	case "character_literal":
		return &ast.Ident{Name: node.Content(source)}
	case "true", "false":
		return &ast.Ident{Name: node.Content(source)}
	}
	panic("Unknown node to expr conversion: " + node.Type())
}

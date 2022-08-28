package main

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/NickyBoy89/java2go/nodeutil"
	"github.com/NickyBoy89/java2go/symbol"
	log "github.com/sirupsen/logrus"
	sitter "github.com/smacker/go-tree-sitter"
)

func ParseExpr(node *sitter.Node, source []byte, ctx Ctx) ast.Expr {
	switch node.Type() {
	case "ERROR":
		log.WithFields(log.Fields{
			"parsed":    node.Content(source),
			"className": ctx.className,
		}).Warn("Expression parse error")
		return &ast.BadExpr{}
	case "comment":
		return &ast.BadExpr{}
	case "update_expression":
		// This can either be a pre or post expression
		// a pre expression has the identifier second, while the post expression
		// has the identifier first

		// Post-update expression, e.g. `i++`
		if node.Child(0).IsNamed() {
			return &ast.CallExpr{
				Fun:  &ast.Ident{Name: "PostUpdate"},
				Args: []ast.Expr{ParseExpr(node.Child(0), source, ctx)},
			}
		}

		// Otherwise, pre-update expression
		return &ast.CallExpr{
			Fun:  &ast.Ident{Name: "PreUpdate"},
			Args: []ast.Expr{ParseExpr(node.Child(1), source, ctx)},
		}
	case "class_literal":
		// Class literals refer to the class directly, such as
		// Object.class
		return &ast.BadExpr{}
	case "assignment_expression":
		return &ast.CallExpr{
			Fun: &ast.Ident{Name: "AssignmentExpression"},
			Args: []ast.Expr{
				ParseExpr(node.Child(0), source, ctx),
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("\"%s\"", node.Child(1).Content(source))},
				ParseExpr(node.Child(2), source, ctx),
			},
		}
	case "super":
		return &ast.BadExpr{}
	case "lambda_expression":
		// Lambdas can either be called with a list of expressions
		// (ex: (n1, n1) -> {}), or with a single expression
		// (ex: n1 -> {})

		var lambdaBody *ast.BlockStmt

		var lambdaParameters *ast.FieldList

		bodyNode := node.ChildByFieldName("body")

		switch bodyNode.Type() {
		case "block":
			lambdaBody = ParseStmt(bodyNode, source, ctx).(*ast.BlockStmt)
		default:
			// Lambdas can be called inline without a block expression
			lambdaBody = &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{
						X: ParseExpr(bodyNode, source, ctx),
					},
				},
			}
		}

		paramNode := node.ChildByFieldName("parameters")

		switch paramNode.Type() {
		case "inferred_parameters", "formal_parameters":
			lambdaParameters = ParseNode(paramNode, source, ctx).(*ast.FieldList)
		default:
			// If we can't identify the types of the parameters, then just set their
			// types to any
			lambdaParameters = &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{ParseExpr(paramNode, source, ctx).(*ast.Ident)},
						Type:  &ast.Ident{Name: "any"},
					},
				},
			}
		}

		return &ast.FuncLit{
			Type: &ast.FuncType{
				Params: lambdaParameters,
			},
			Body: lambdaBody,
		}
	case "method_reference":
		// This refers to manually selecting a function from a specific class and
		// passing it in as an argument in the `func(className::methodName)` style

		// For class constructors such as `Class::new`, you only get one node
		if node.NamedChildCount() < 2 {
			return &ast.SelectorExpr{
				X:   ParseExpr(node.NamedChild(0), source, ctx),
				Sel: &ast.Ident{Name: "new"},
			}
		}

		return &ast.SelectorExpr{
			X:   ParseExpr(node.NamedChild(0), source, ctx),
			Sel: ParseExpr(node.NamedChild(1), source, ctx).(*ast.Ident),
		}
	case "array_initializer":
		// A literal that initilzes an array, such as `{1, 2, 3}`
		items := []ast.Expr{}
		for _, c := range nodeutil.NamedChildrenOf(node) {
			items = append(items, ParseExpr(c, source, ctx))
		}

		// If there wasn't a type for the array specified, then use the one that has been defined
		if _, ok := ctx.lastType.(*ast.ArrayType); ctx.lastType != nil && ok {
			return &ast.CompositeLit{
				Type: ctx.lastType.(*ast.ArrayType),
				Elts: items,
			}
		}
		return &ast.CompositeLit{
			Elts: items,
		}
	case "method_invocation":
		// Methods with a selector are called as X.Sel(Args)
		// Otherwise, they are called as Fun(Args)
		if node.ChildByFieldName("object") != nil {
			return &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ParseExpr(node.ChildByFieldName("object"), source, ctx),
					Sel: ParseExpr(node.ChildByFieldName("name"), source, ctx).(*ast.Ident),
				},
				Args: ParseNode(node.ChildByFieldName("arguments"), source, ctx).([]ast.Expr),
			}
		}
		return &ast.CallExpr{
			Fun:  ParseExpr(node.ChildByFieldName("name"), source, ctx),
			Args: ParseNode(node.ChildByFieldName("arguments"), source, ctx).([]ast.Expr),
		}
	case "object_creation_expression":
		// This is called when anything is created with a constructor

		objectType := node.ChildByFieldName("type")

		// A object can also be created with this format:
		// parentClass.new NestedClass()
		if !node.NamedChild(0).Equal(objectType) {
		}

		// Get all the arguments, and look up their types
		objectArguments := node.ChildByFieldName("arguments")
		arguments := make([]ast.Expr, objectArguments.NamedChildCount())
		argumentTypes := make([]string, objectArguments.NamedChildCount())
		for ind, argument := range nodeutil.NamedChildrenOf(objectArguments) {
			arguments[ind] = ParseExpr(argument, source, ctx)
			// Look up each argument and find its type
			if argument.Type() != "identifier" {
				argumentTypes[ind] = symbol.TypeOfLiteral(argument, source)
			} else {
				if localDef := ctx.localScope.FindVariable(argument.Content(source)); localDef != nil {
					argumentTypes[ind] = localDef.OriginalType
					// Otherwise, a variable may exist as a global variable
				} else if def := ctx.currentClass.FindField().ByOriginalName(argument.Content(source))[0]; def != nil {
					argumentTypes[ind] = def.OriginalType
				}
			}
		}

		var constructor *symbol.Definition
		// Find the respective constructor, and call it
		if objectType.Type() == "generic_type" {
			constructor = ctx.currentClass.FindMethodByName(objectType.NamedChild(0).Content(source), argumentTypes)
		} else {
			constructor = ctx.currentClass.FindMethodByName(objectType.Content(source), argumentTypes)
		}

		if constructor != nil {
			return &ast.CallExpr{
				Fun:  &ast.Ident{Name: constructor.Name},
				Args: arguments,
			}
		}

		// It is also possible that a constructor could be unresolved, so we handle
		// this by calling the type of the type + "Construct" at the beginning
		return &ast.CallExpr{
			Fun:  &ast.Ident{Name: "Construct" + objectType.Content(source)},
			Args: arguments,
		}
	case "array_creation_expression":
		arguments := []ast.Expr{&ast.ArrayType{Elt: ParseExpr(node.ChildByFieldName("type"), source, ctx)}}

		for _, child := range nodeutil.NamedChildrenOf(node) {
			if child.Type() == "dimensions_expr" {
				arguments = append(arguments, ParseExpr(child, source, ctx))
			}
		}

		var methodName string
		switch len(arguments) - 1 {
		case 0:
			expr := ParseExpr(node.ChildByFieldName("value"), source, ctx).(*ast.CompositeLit)
			expr.Type = &ast.ArrayType{
				Elt: ParseExpr(node.ChildByFieldName("type"), source, ctx),
			}
			return expr
		case 1:
			methodName = "make"
		case 2:
			methodName = "MultiDimensionArray"
		case 3:
			methodName = "MultiDimensionArray3"
		default:
			panic("Unimplemented number of dimensions in array initializer")
		}

		return &ast.CallExpr{
			Fun:  &ast.Ident{Name: methodName},
			Args: arguments,
		}
	case "instanceof_expression":
		return &ast.BadExpr{}
	case "dimensions_expr":
		return ParseExpr(node.NamedChild(0), source, ctx)
	case "binary_expression":
		if node.Child(1).Content(source) == ">>>" {
			return &ast.CallExpr{
				Fun:  &ast.Ident{Name: "UnsignedRightShift"},
				Args: []ast.Expr{ParseExpr(node.Child(0), source, ctx), ParseExpr(node.Child(2), source, ctx)},
			}
		}
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
		for _, c := range nodeutil.NamedChildrenOf(node) {
			args = append(args, ParseExpr(c, source, ctx))
		}
		return &ast.CallExpr{
			Fun:  &ast.Ident{Name: "ternary"},
			Args: args,
		}
	case "cast_expression":
		// TODO: This probably should be a cast function, instead of an assertion
		return &ast.TypeAssertExpr{
			X:    ParseExpr(node.NamedChild(1), source, ctx),
			Type: ParseExpr(node.NamedChild(0), source, ctx),
		}
	case "field_access":
		// X.Sel
		obj := node.ChildByFieldName("object")

		if obj.Type() == "this" {
			def := ctx.currentClass.FindField().ByOriginalName(node.ChildByFieldName("field").Content(source))[0]
			if def == nil {
				// TODO: This field could not be found in the current class, because it exists in the superclass
				// definition for the class
				def = &symbol.Definition{
					Name: node.ChildByFieldName("field").Content(source),
				}
			}

			return &ast.SelectorExpr{
				X:   ParseExpr(node.ChildByFieldName("object"), source, ctx),
				Sel: &ast.Ident{Name: def.Name},
			}
		}
		return &ast.SelectorExpr{
			X:   ParseExpr(obj, source, ctx),
			Sel: ParseExpr(node.ChildByFieldName("field"), source, ctx).(*ast.Ident),
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
	case "type_identifier": // Any reference type
		switch node.Content(source) {
		// Special case for strings, because in Go, these are primitive types
		case "String":
			return &ast.Ident{Name: "string"}
		}

		if ctx.currentFile != nil {
			// Look for the class locally first
			if localClass := ctx.currentFile.FindClass(node.Content(source)); localClass != nil {
				return &ast.StarExpr{
					X: &ast.Ident{Name: localClass.Name},
				}
			}
		}

		return &ast.StarExpr{
			X: &ast.Ident{Name: node.Content(source)},
		}
	case "null_literal":
		return &ast.Ident{Name: "nil"}
	case "decimal_integer_literal":
		literal := node.Content(source)
		switch literal[len(literal)-1] {
		case 'L':
			return &ast.CallExpr{Fun: &ast.Ident{Name: "int64"}, Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: literal[:len(literal)-1]}}}
		}
		return &ast.Ident{Name: literal}
	case "hex_integer_literal":
		return &ast.Ident{Name: node.Content(source)}
	case "decimal_floating_point_literal":
		// This is something like 1.3D or 1.3F
		literal := node.Content(source)
		switch literal[len(literal)-1] {
		case 'D':
			return &ast.CallExpr{Fun: &ast.Ident{Name: "float64"}, Args: []ast.Expr{&ast.BasicLit{Kind: token.FLOAT, Value: literal[:len(literal)-1]}}}
		case 'F':
			return &ast.CallExpr{Fun: &ast.Ident{Name: "float32"}, Args: []ast.Expr{&ast.BasicLit{Kind: token.FLOAT, Value: literal[:len(literal)-1]}}}
		}
		return &ast.Ident{Name: literal}
	case "string_literal":
		return &ast.Ident{Name: node.Content(source)}
	case "character_literal":
		return &ast.Ident{Name: node.Content(source)}
	case "true", "false":
		return &ast.Ident{Name: node.Content(source)}
	}
	return nil
}

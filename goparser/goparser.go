package goparser

import (
	"fmt"
	"unicode"
	"strings"

	"gitlab.nicholasnovak.io/snapdragon/java2go/parsing"
	"gitlab.nicholasnovak.io/snapdragon/java2go/parsetools"
	"gitlab.nicholasnovak.io/snapdragon/java2go/codeparser"
)

const indentNum = 2

// A list of in-scope variables, for type-checking reasons
var inScopeVariables = make(map[string]string)

func ClearInScopeVariables() {
	inScopeVariables = make(map[string]string)
}

// NewClass is set to true if the class is a nested class
func ParseFile(sourceFile parsing.ParsedClasses, newClass bool) string {
	var generated string
	if newClass {
		generated += fmt.Sprintf("package main\n\n")
		generated += "func NewAssertionError(err string) error {\n  return errors.New(err)\n}\n\n"
	}
	fmt.Printf("Generated %s\n", sourceFile.GetType())
	switch sourceFile.GetType() {
	case "class":
		generated += ParseClass(sourceFile.(parsing.ParsedClass)) // Parse the class into one struct
	case "interface":
		panic("Parsing interface not implemented")
	case "enum":
		panic("Parsing enum not implemented")
	default:
		panic("Unknown class type: " + sourceFile.GetType())
	}

	// Replace generated methods
	for found, newFunc := range specializedFunctions {
		generated = strings.ReplaceAll(generated, found, newFunc)
	}

	return generated
}

// Parse a given class
func ParseClass(source parsing.ParsedClass) string {
	var generated string

	// Create a context for the class, so that the methods have some frame of reference
	classContext := new(ClassContext)
	// Set the name of the class context
	classContext.Name = ToPublic(source.Name)
	// Extract the method names from the class itself, before it has been parsed
	classContext.Methods = source.MethodContext()

	// Register a NewAssertionError for error-handling
	classContext.Methods["AssertionError"] = []string{"string"}

	// If the line below is commented out, then every struct will be declared as public

	// if !IsPublic(source.Modifiers) {
	// 	classContext.Name = ToPrivate(source.Name)
	// }

	// Parse the class itself as a struct
	// If the class is static, don't generate a struct for it
	if !parsetools.Contains("static", source.Modifiers) {
		generated += CreateStruct(classContext, source.ClassVariables)
	}

	generated += "\n\n" // Add some spacing after the initial struct

	// Parse the methods of the class
	for _, method := range source.Methods {
		generated += CreateMethod(classContext, method)
		generated += "\n\n" // Add some spacing in between the methods
	}

	// Parse the nested classes
	for _, nested := range source.NestedClasses {
		generated += ParseFile(nested, false)
	}

	return generated
}

func CreateStruct(classContext *ClassContext, fields []parsing.ParsedVariable) string {
	result := fmt.Sprintf("type %s struct {", classContext.Name)
	for _, field := range fields { // Struct fields
		dataType := field.DataType
		// Treats a capitalized data type of a field as a custom (non-primitive) data type, and makes it into a pointer to that data
		// This is done because any non-primitive type is always a reference type in java
		if unicode.IsUpper(rune(field.DataType[0])) {
			dataType = "*" + dataType
		}

		// Write out a field (ex: value int)
		if IsPublic(field.Modifiers) {
			result += fmt.Sprintf("\n%s%s %s", strings.Repeat(" ", indentNum), ToPublic(field.Name), dataType)
		} else {
			result += fmt.Sprintf("\n%s%s %s", strings.Repeat(" ", indentNum), field.Name, dataType)
		}
	}
	return result + "\n}"
}

func ToPublic(name string) string {
	return string(unicode.ToUpper(rune(name[0]))) + name[1:]
}

func ToPrivate(name string) string {
	return string(unicode.ToLower(rune(name[0]))) + name[1:]
}

// Tests if an object is public, given its modifiers
func IsPublic(modifiers []string) bool {
	if parsetools.Contains("public", modifiers) || parsetools.Contains("protected", modifiers) {
		return true
	}
	return false
}

// Generates a two-character shorthand of a struct's name for a method (ex: IntLinkedList -> it)
func AsShorthand(name string) string {
	return string(unicode.ToLower(rune(name[0]))) + string(unicode.ToLower(rune(name[len(name) - 1])))
}

// Creates a string representation of a method from the parsed method and the name of the class that it came from
// For a static method (standalone class) pass in an empty class name
func CreateMethod(classContext *ClassContext, methodSource parsing.ParsedMethod) string {
	var result string
	if parsetools.Contains("static", methodSource.Modifiers) { // Method is static, so not associated with any class
		// Special methods, ex: main, init
		switch methodSource.Name {
		case "main":
			result += fmt.Sprintf("func %s(", methodSource.Name)
		default:
			if IsPublic(methodSource.Modifiers) {
				result += fmt.Sprintf("func %s(", ToPublic(methodSource.Name))
			} else {
				result += fmt.Sprintf("func %s(", ToPrivate(methodSource.Name))
			}
		}
	} else if methodSource.ReturnType == "constructor" { // Constructor methods just get handled as generator functions
		// if IsPublic(methodSource.Modifiers) {
			result += fmt.Sprintf("func New%s(", classContext.Name) // If public, the constructor function is public as well
		// } else {
			// result += fmt.Sprintf("func new%s(", classContext.Name) // Private constructor
		// }
	} else {
		if IsPublic(methodSource.Modifiers) {
			result += fmt.Sprintf(
				"func (%s *%s) %s(",
				AsShorthand(classContext.Name),
				classContext.Name,
				ToPublic(methodSource.Name),
			)
		} else {
			result += fmt.Sprintf(
				"func (%s *%s) %s(",
				AsShorthand(classContext.Name),
				classContext.Name,
				ToPrivate(methodSource.Name),
			)
		}
	}

	if methodSource.Name != "main" {
		for pi, param := range methodSource.Parameters { // Parameters
			result += param.Name + " " + ToReferenceType(JavaToGoArray(param.DataType))
			if pi < len(methodSource.Parameters) - 1 {
				result += ", "
			}
		}
	}

	if methodSource.ReturnType == "constructor" {
		result += fmt.Sprintf(
			") *%s {\n%s%s := new(%s)\n%s\n%sreturn %s\n}",
			classContext.Name,
			strings.Repeat(" ", indentNum),
			AsShorthand(classContext.Name),
			classContext.Name,
			CreateBody(methodSource.Body, classContext, 2),
			strings.Repeat(" ", indentNum),
			AsShorthand(classContext.Name),
		)
		return result
	}
	result += fmt.Sprintf(") %v {\n%s\n}", ReplaceWord(methodSource.ReturnType), CreateBody(methodSource.Body, classContext, 2))
	ClearInScopeVariables()
	return result
}

// Parses the lines of the body
func CreateBody(body []codeparser.LineTyper, classContext *ClassContext, indentation int) string {
	var result string
	for _, line := range body {
		// fmt.Printf("Going through line of type: %s\n", line.GetName())
		result += CreateLine(line, classContext, indentation, true)
	}
	return result
}

func CreateLine(line codeparser.LineTyper, classContext *ClassContext, indentation int, indent bool) string {
	var result string
	if indent {
		result += "\n"
	}
	// result += fmt.Sprintf("//%s\n", line.GetName())
	switch line.GetName() {
	case "GenericLine":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Statement"].([]codeparser.LineType) {
			body += CreateLine(line, classContext, 0, false)
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s", body)
	case "StringLiteral":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s", line.(codeparser.LineType).Words["String"])
	case "CreateAndAssignVariable":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += CreateLine(line, classContext	, 0, false)
		}
		// Commented-out block is for variables to be declared explicitly
		// result += strings.Repeat(" ", indentation) + fmt.Sprintf(
		// 	"var %s %s = %s",
		// 	line.(codeparser.LineType).Words["VariableName"],
		// 	JavaToGoArray(ReplaceWord(line.(codeparser.LineType).Words["VariableType"].(string))),
		// 	body,
		// )
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"%s := %s",
			CreateLine(line.(codeparser.LineType).Words["VariableName"].([]codeparser.LineType)[0], classContext, 0, false),
			body,
		)
		inScopeVariables[line.(codeparser.LineType).Words["VariableName"].([]codeparser.LineType)[0].Words["Expression"].(string)] = JavaToGoArray(ReplaceWord(line.(codeparser.LineType).Words["VariableType"].(string)))
	case "AssignVariable":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += ReplaceWord(CreateLine(line, classContext, 0, false))
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"%s = %s",
			CreateLine(line.(codeparser.LineType).Words["VariableName"].([]codeparser.LineType)[0], classContext, 0, false),
			body,
		)
	case "CompoundAssignment":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += ReplaceWord(CreateLine(line, classContext, 0, false))
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"%s %s= %s",
			line.(codeparser.LineType).Words["VariableName"],
			line.(codeparser.LineType).Words["Operator"],
			body,
		)
	case "FunctionCall":
		functionName := line.(codeparser.LineType).Words["FunctionName"].(string)
		if classContext.ContainsMethod(ToPrivate(functionName)) {
			functionName = ToPrivate(functionName)
		} else if classContext.ContainsMethod(ToPublic(functionName)) {
			functionName = ToPublic(functionName)
		} else {
			panic("Unknown non-package function " + functionName + "")
		}

		// Populate the parameters of the function
		var body string
		for li, expressionLine := range line.(codeparser.LineType).Words["Parameters"].([][]codeparser.LineType) {
			for _, expLine := range expressionLine {
				body += CreateLine(expLine, classContext, 0, false)
			}
			if li != len(line.(codeparser.LineType).Words["Parameters"].([][]codeparser.LineType)) - 1 { // For the commas, don't add one to the last element
				body += ", "
			}
		}

		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"%s(%s)",
			functionName,
			body,
		)
	case "ConstructArray":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("make([]%s, %s)", line.(codeparser.LineType).Words["ArrayType"], line.(codeparser.LineType).Words["InitialSize"])
	case "AccessArrayElement":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s[%s]", line.(codeparser.LineType).Words["ArrayName"], line.(codeparser.LineType).Words["Index"])
	case "ReturnStatement":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += CreateLine(line, classContext, 0, false)
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("return %s", body)
	case "IfStatement":
		var body string
		for _, line := range line.(codeparser.LineBlock).Words["Condition"].([]codeparser.LineType) {
			body += CreateLine(line, classContext, 0, false) + " "
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"if %s {%s\n%s}",
			body,
			CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2),
			strings.Repeat(" ", indentation),
		)
	case "ElseLoop":
		// This is an equals only, to cut out the newline
		result = fmt.Sprintf(" else {%s\n%s}\n", CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2), strings.Repeat(" ", indentation))
	case "ForLoop":
		var body string
		for _, line := range line.(codeparser.LineBlock).Words["Conditional"].([]codeparser.LineType) {
			body += CreateLine(line, classContext, 0, false) + " "
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"for %s; %s; %s {%s\n%s}",
			CreateLine(line.(codeparser.LineBlock).Words["Initializer"].(codeparser.LineTyper), classContext, 0, false),
			body,
			CreateLine(line.(codeparser.LineBlock).Words["Incrementer"].(codeparser.LineTyper), classContext, 0, false),
			CreateBody(line.(codeparser.LineBlock).Lines, classContext, indentation + 2),
			strings.Repeat(" ", indentation),
		)
	case "NewConstructor":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("New%s", CreateLine(line.(codeparser.LineType).Words["Expression"].(codeparser.LineType), classContext, 0, false))
	case "ThrowException":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("panic(%s)\n", CreateLine(line.(codeparser.LineType).Words["Expression"].(codeparser.LineType), classContext, 0, false))
	case "ImplicitArrayAssignment":
		var body string
		for li, expressionLine := range line.(codeparser.LineType).Words["Elements"].([][]codeparser.LineType) {
			for _, expLine := range expressionLine {
				body += CreateLine(expLine, classContext, 0, false)
			}
			if li != len(line.(codeparser.LineType).Words["Elements"].([][]codeparser.LineType)) - 1 { // For the commas, don't add one to the last element
				body += ", "
			}
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s{%s}", JavaToGoArray(ReplaceWord(line.(codeparser.LineType).Words["ArrayType"].(string))), body)
	// The expression types, don't have a newline associated with them
	case "LocalVariableOrExpression":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s", ReplaceWord(line.(codeparser.LineType).Words["Expression"].(string)))
	case "RemoteVariableOrExpression":
		packageName := line.(codeparser.LineType).Words["RemotePackage"]

		// The expression field will only ever contain one entry (ex: value in this.value)
		// Let's just assert that
		if len(line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType)) != 1 {
			panic("Remote expression does not have one expression, this should not be the case")
		}

		expression := CreateLine(line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType)[0], classContext, 0, false)

		switch packageName {
		case "this": // If package name is reserved word "this", then treat it as referring to the struct method's shorthand
			packageName = AsShorthand(classContext.Name)
		case classContext.Name: // If the package name is the current package
			result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s", expression)
			expression = ""
		}

		switch expression {
		case "":
		case "length": // Getting the "length" field of a variable will instead call the len() go builtin function
			result += strings.Repeat(" ", indentation) + fmt.Sprintf(
				"len(%s)",
				packageName,
			)
		default:
			result += strings.Repeat(" ", indentation) + fmt.Sprintf(
				"%s.%s",
				packageName,
				expression,
			)
		}
	default:
		panic("Unknown line type: " + line.GetName())
	}

	return result
}

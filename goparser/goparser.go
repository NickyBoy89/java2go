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

// NewClass is set to true if the class is a nested class
func ParseFile(sourceFile parsing.ParsedClasses, newClass bool) string {
	var generated string
	if newClass {
		generated += fmt.Sprintf("package main\n\n")
	}
	fmt.Println(sourceFile.GetType())
	switch sourceFile.GetType() {
	case "class":
		generated += ParseClass(sourceFile.(parsing.ParsedClass)) // Parse the class into one struct
	case "interface":
	case "enum":
	default:
		panic("Unknown class type: " + sourceFile.GetType())
	}
	return generated
}

// Parse a given class
func ParseClass(source parsing.ParsedClass) string {
	var generated string

	className := ToPublic(source.Name)
	if !IsPublic(source.Modifiers) {
		className = ToPrivate(source.Name)
	}

	// Parse the class itself as a struct
	// If the class is static, don't generate a struct for it
	if !parsetools.Contains("static", source.Modifiers) {
		generated += CreateStruct(className, source.ClassVariables)
	}

	generated += "\n\n" // Add some spacing after the initial struct

	// Parse the methods of the class
	for _, method := range source.Methods {
		generated += CreateMethod(className, method)
		generated += "\n\n" // Add some spacing in between the methods
	}

	return generated
}

func CreateStruct(name string, fields []parsing.ParsedVariable) string {
	result := fmt.Sprintf("type %s struct {", name)
	for _, field := range fields { // Struct fields
		dataType := field.DataType
		// Treats a capitalized data type of a field as a custom (non-primitive) data type, and makes it into a pointer to that data
		// This is done because any non-primitive type is always a reference type
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
func CreateMethod(className string, methodSource parsing.ParsedMethod) string {
	var result string
	if className == "" { // Class name is blank, the method is just a plain function
		if IsPublic(methodSource.Modifiers) {
			result += fmt.Sprintf("func %s(", ToPublic(methodSource.Name))
		} else {
			result += fmt.Sprintf("func %s(", ToPrivate(methodSource.Name))
		}
	} else if methodSource.ReturnType == "constructor" { // Constructor methods just get handled as generator functions
		if IsPublic(methodSource.Modifiers) {
			result += fmt.Sprintf("func New%s(", className) // If public, the constructor function is public as well
		} else {
			result += fmt.Sprintf("func new%s(", className) // Private constructor
		}
	} else {
		if IsPublic(methodSource.Modifiers) {
			result += fmt.Sprintf("func (%s %s) %s(", AsShorthand(className), className, ToPublic(methodSource.Name))
		} else {
			result += fmt.Sprintf("func (%s %s) %s(", AsShorthand(className), className, ToPrivate(methodSource.Name))
		}
	}

	for pi, param := range methodSource.Parameters { // Parameters
		result += param.Name + " " + JavaToGoArray(param.DataType)
		if pi < len(methodSource.Parameters) - 1 {
			result += ", "
		}
	}

	if methodSource.ReturnType == "constructor" {
		result += fmt.Sprintf(") *%s {\n%s%s := new(%s)\n%s\n%sreturn %s\n}", className, strings.Repeat(" ", indentNum), AsShorthand(className), className, CreateBody(methodSource.Body, className, 2), strings.Repeat(" ", indentNum), AsShorthand(className))
		return result
	}
	result += fmt.Sprintf(") %v {\n%s\n}", ReplaceWord(methodSource.ReturnType), CreateBody(methodSource.Body, className, 2))
	return result
}

// Parses the lines of the body
func CreateBody(body []codeparser.LineTyper, className string, indentation int) string {
	var result string
	for _, line := range body {
		fmt.Printf("Going through line of type: %s\n", line.GetName())
		result += CreateLine(line, className, indentation, true)
	}
	return result
}

func CreateLine(line codeparser.LineTyper, className string, indentation int, indent bool) string {
	var result string
	if indent {
		result += "\n"
	}
	// result += fmt.Sprintf("//%s\n", line.GetName())
	switch line.GetName() {
	case "GenericLine":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Statement"].([]codeparser.LineType) {
			body += CreateLine(line, className, 0, false)
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s", body)
	case "CreateAndAssignVariable":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += CreateLine(line, className, 0, false)
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
			CreateLine(line.(codeparser.LineType).Words["VariableName"].([]codeparser.LineType)[0], className, 0, false),
			body,
		)
	case "AssignVariable":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += ReplaceWord(CreateLine(line, className, 0, false))
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"%s = %s",
			CreateLine(line.(codeparser.LineType).Words["VariableName"].([]codeparser.LineType)[0], className, 0, false),
			body,
		)
	case "CompoundAssignment":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += ReplaceWord(CreateLine(line, className, 0, false))
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"%s %s= %s",
			line.(codeparser.LineType).Words["VariableName"],
			line.(codeparser.LineType).Words["Operator"],
			body,
		)
	case "FunctionCall":
		var body string
		for li, expressionLine := range line.(codeparser.LineType).Words["Parameters"].([]codeparser.LineType) {
			body += CreateLine(expressionLine, className, 0, false)
			if li != len(line.(codeparser.LineType).Words["Parameters"].([]codeparser.LineType)) - 1 { // For the commas, don't add one to the last element
				body += ", "
			}
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"%s(%s)",
			line.(codeparser.LineType).Words["FunctionName"],
			body,
		)
	case "ReturnStatement":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += CreateLine(line, className, 0, false)
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("return %s", body)
	case "IfStatement":
		var body string
		for _, line := range line.(codeparser.LineBlock).Words["Condition"].([]codeparser.LineType) {
			body += CreateLine(line, className, 0, false) + " "
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"if %s {%s\n%s}",
			body,
			CreateBody(line.(codeparser.LineBlock).Lines, className, indentation + 2),
			strings.Repeat(" ", indentation),
		)
	case "ElseLoop":
		// This is an equals only, to cut out the newline
		result = fmt.Sprintf(" else {%s\n%s}\n", CreateBody(line.(codeparser.LineBlock).Lines, className, indentation + 2), strings.Repeat(" ", indentation))
	case "ForLoop":
		var body string
		for _, line := range line.(codeparser.LineBlock).Words["Conditional"].([]codeparser.LineType) {
			body += CreateLine(line, className, 0, false) + " "
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf(
			"for %s; %s; %s {%s\n%s}",
			CreateLine(line.(codeparser.LineBlock).Words["Initializer"].(codeparser.LineTyper), className, 0, false),
			body,
			CreateLine(line.(codeparser.LineBlock).Words["Incrementer"].(codeparser.LineTyper), className, 0, false),
			CreateBody(line.(codeparser.LineBlock).Lines, className, indentation + 2),
			strings.Repeat(" ", indentation),
		)
	case "NewConstructor":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("New%s", CreateLine(line.(codeparser.LineType).Words["Expression"].(codeparser.LineType), className, 0, false))
	case "ThrowException":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("panic(%s)\n", CreateLine(line.(codeparser.LineType).Words["Expression"].(codeparser.LineType), className, 0, false))
	case "ImplicitArrayAssignment":
		var body string
		for li, expressionLine := range line.(codeparser.LineType).Words["Elements"].([]codeparser.LineType) {
			body += CreateLine(expressionLine, className, 0, false)
			if li != len(line.(codeparser.LineType).Words["Elements"].([]codeparser.LineType)) - 1 { // For the commas, don't add one to the last element
				body += ", "
			}
		}
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s{%s}", JavaToGoArray(ReplaceWord(line.(codeparser.LineType).Words["ArrayType"].(string))), body)
	// The expression types, don't have a newline associated with them
	case "LocalVariableOrExpression":
		result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s", ReplaceWord(line.(codeparser.LineType).Words["Expression"].(string)))
	case "RemoteVariableOrExpression":
		if line.(codeparser.LineType).Words["RemotePackage"] == "this" {
			result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s.%s", AsShorthand(className), line.(codeparser.LineType).Words["Expression"])
		} else {
			result += strings.Repeat(" ", indentation) + fmt.Sprintf("%s.%s", line.(codeparser.LineType).Words["RemotePackage"], line.(codeparser.LineType).Words["Expression"])
		}
	default:
		panic("Unknown line type: " + line.GetName())
	}

	return result
}

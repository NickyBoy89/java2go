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
		result += fmt.Sprintf(") *%s {\n%sresult := new(%s)\n%s\n%sreturn result\n}", className, strings.Repeat(" ", indentNum), className, CreateBody(methodSource.Body, className), strings.Repeat(" ", indentNum))
		return result
	}
	result += fmt.Sprintf(") %v {\n%s\n}", ReplaceWord(methodSource.ReturnType), CreateBody(methodSource.Body, className))
	return result
}

// Parses the lines of the body
func CreateBody(body []codeparser.LineTyper, className string) string {
	var result string
	for _, line := range body {
		// fmt.Printf("Going through line of type: %s\n", line.GetName())
		result += CreateLine(line, className)
	}
	return result
}

func CreateLine(line codeparser.LineTyper, className string) string {
	switch line.GetName() {
	case "GenericLine":
		return fmt.Sprintf("%s\n", line.(codeparser.LineType).Words["Statement"])
	case "CreateAndAssignVariable":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += CreateLine(line, className)
		}
		return fmt.Sprintf("var %s %s = %s\n", line.(codeparser.LineType).Words["VariableName"], line.(codeparser.LineType).Words["VariableType"], body)
	case "AssignVariable":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += CreateLine(line, className)
		}
		return fmt.Sprintf("%s = %s\n", line.(codeparser.LineType).Words["VariableName"], body)
	case "ReturnStatement":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += CreateLine(line, className)
		}
		return fmt.Sprintf("return %s\n", body)
	case "IfStatement":
		return fmt.Sprintf("if %s {\n%s\n}\n", line.(codeparser.LineBlock).Words["Condition"], CreateBody(line.(codeparser.LineBlock).Lines, className))
	case "ElseLoop":
		return fmt.Sprintf(" else {\n%s\n}\n", CreateBody(line.(codeparser.LineBlock).Lines, className))
	case "ForLoop":
		return fmt.Sprintf("for %s; %s; %s {\n%s\n}\n", line.(codeparser.LineBlock).Words["Initializer"], line.(codeparser.LineBlock).Words["Conditional"], line.(codeparser.LineBlock).Words["Incrementer"], CreateBody(line.(codeparser.LineBlock).Lines, className))
	case "ThrowException":
		var body string
		for _, line := range line.(codeparser.LineType).Words["Expression"].([]codeparser.LineType) {
			body += CreateLine(line, className)
		}
		return fmt.Sprintf("panic(%s)\n", body)
	case "LocalVariableOrExpression":
		return fmt.Sprintf("%s\n", line.(codeparser.LineType).Words["Expression"])
	case "RemoteVariableOrExpression":
		if line.(codeparser.LineType).Words["RemotePackage"] == "this" {
			return fmt.Sprintf("%s.%s\n", AsShorthand(className), line.(codeparser.LineType).Words["Expression"])
		}
		return fmt.Sprintf("%s.%s\n", line.(codeparser.LineType).Words["RemotePackage"], line.(codeparser.LineType).Words["Expression"])
	default:
		panic("Unknown line type: " + line.GetName())
	}
}

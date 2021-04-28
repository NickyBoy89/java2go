package goparser

import (
	"fmt"
	"unicode"
	"strings"

	"gitlab.nicholasnovak.io/snapdragon/java2go/parsing"
)

const indentNum = 2

func ParseFile(sourceFile parsing.ParsedClasses, newClass bool) string {
	switch sourceFile.GetType() {
	case "class":
		return ParseClass(sourceFile.(parsing.ParsedClass), newClass)
	case "interface":
		return ParseInterface(sourceFile.(parsing.ParsedInterface), newClass)
	case "enum":
		return ParseEnum(sourceFile.(parsing.ParsedEnum), newClass)
	default:
		panic("Incompatible object to parse, this should only happen for a custom object to parse")
	}
}

// Parses a class, if the newClass param is true, then a package declaration will be made as well
func ParseClass(sourceClass parsing.ParsedClass, newClass bool) string {
	var result string
	if newClass {
		result += "package main\n\n"
	}

	result += CreateStruct(sourceClass.Name, sourceClass.ClassVariables) + "\n"

	for _, method := range sourceClass.Methods {
		result += "\n" + CreateMethod(sourceClass.Name, method) + "\n"
	}

	// Nested classes
	for _, nestedClass := range sourceClass.NestedClasses {
		result += "\n" + ParseFile(nestedClass, false)
	}

	return result
}

func ParseInterface(sourceInterface parsing.ParsedInterface, newInterface bool) string {
	return ""
}

func ParseEnum(sourceEnum parsing.ParsedEnum, newEnum bool) string {
	return ""
}

// Parses the blocked content of the data
func ParseContent(content string) string {
	contentWords := strings.Split()
	var parsed string

	fmt.Println(strings.Split(content, " "))

	lastInterest := 0
	ci := 0
	for ; ci < len(content); ci++ {
		c := content[ci]
		switch c {
		case ';':
			parsed += strings.Repeat(" ", indentNum) + content[lastInterest:ci] + "\n"
			lastInterest = ci + 1 // Skip the semicolon
		}
	}

	return parsed
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

func IsPublic(modifiers []string) bool {
	if parsing.Contains("public", modifiers) || parsing.Contains("protected", modifiers) {
		return true
	}
	return false
}

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
		result += fmt.Sprintf(") *%s {\n%sresult := new(%s)\n%s\n%sreturn result\n}", className, strings.Repeat(" ", indentNum), className, ParseContent(methodSource.Body), strings.Repeat(" ", indentNum))
		return result
	}
	result += fmt.Sprintf(") %v {\n%s\n}", ReplaceWord(methodSource.ReturnType), ParseContent(methodSource.Body))
	return result
}
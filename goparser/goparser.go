package goparser

import (
	"fmt"
	"unicode"
	"strings"

	"gitlab.nicholasnovak.io/snapdragon/java2go/parsing"
	"gitlab.nicholasnovak.io/snapdragon/java2go/parsetools"
	// "gitlab.nicholasnovak.io/snapdragon/java2go/codeparser"
)

const indentNum = 2

func ParseFile(sourceFile parsing.ParsedClasses, newClass bool) string {
	fmt.Println(sourceFile.GetType())
	switch sourceFile.GetType() {
	case "class", "interface", "enum":

		return ""
	default:
		panic("Unknown class type: " + sourceFile.GetType())
	}
}

// Parse a given class
func ParseClass(source parsing.ParsedClass) string {
	return ""
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

// Tests if an object is public, given its modifiers
func IsPublic(modifiers []string) bool {
	if parsetools.Contains("public", modifiers) || parsetools.Contains("protected", modifiers) {
		return true
	}
	return false
}

func AsShorthand(name string) string {
	return string(unicode.ToLower(rune(name[0]))) + string(unicode.ToLower(rune(name[len(name) - 1])))
}

// Creates a string representation of a method from the parsed method and the name of the class that it came from
// For a static method (standalone class) pass in an empty class name
// func CreateMethod(className string, methodSource parsing.ParsedMethod) string {
// 	var result string
// 	if className == "" { // Class name is blank, the method is just a plain function
// 		if IsPublic(methodSource.Modifiers) {
// 			result += fmt.Sprintf("func %s(", ToPublic(methodSource.Name))
// 		} else {
// 			result += fmt.Sprintf("func %s(", ToPrivate(methodSource.Name))
// 		}
// 	} else if methodSource.ReturnType == "constructor" { // Constructor methods just get handled as generator functions
// 		if IsPublic(methodSource.Modifiers) {
// 			result += fmt.Sprintf("func New%s(", className) // If public, the constructor function is public as well
// 		} else {
// 			result += fmt.Sprintf("func new%s(", className) // Private constructor
// 		}
// 	} else {
// 		if IsPublic(methodSource.Modifiers) {
// 			result += fmt.Sprintf("func (%s %s) %s(", AsShorthand(className), className, ToPublic(methodSource.Name))
// 		} else {
// 			result += fmt.Sprintf("func (%s %s) %s(", AsShorthand(className), className, ToPrivate(methodSource.Name))
// 		}
// 	}
//
// 	for pi, param := range methodSource.Parameters { // Parameters
// 		result += param.Name + " " + JavaToGoArray(param.DataType)
// 		if pi < len(methodSource.Parameters) - 1 {
// 			result += ", "
// 		}
// 	}
//
// 	if methodSource.ReturnType == "constructor" {
// 		result += fmt.Sprintf(") *%s {\n%sresult := new(%s)\n%s\n%sreturn result\n}", className, strings.Repeat(" ", indentNum), className, codeparser.ParseContent(methodSource.Body), strings.Repeat(" ", indentNum))
// 		return result
// 	}
// 	result += fmt.Sprintf(") %v {\n%s\n}", ReplaceWord(methodSource.ReturnType), codeparser.ParseContent(methodSource.Body))
// 	return result
// }

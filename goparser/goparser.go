package goparser

import (
	"fmt"
	"unicode"

	"gitlab.nicholasnovak.io/snapdragon/java2go/parsing"
)

const indentation = "  "

func CreateStruct(name string, fields []parsing.ParsedVariable) string {
	result := fmt.Sprintf("type %v struct {", name)
	for _, field := range fields {
		if parsing.Contains("public", field.Modifiers) || parsing.Contains("protected", field.Modifiers) {
			result += fmt.Sprintf("\n%v%v %v", indentation, ToPublic(field.Name), field.DataType)
		} else {
			result += fmt.Sprintf("\n%v%v %v", indentation, field.Name, field.DataType)
		}
	}
	return result + "}"
}

func ToPublic(name string) string {
	return string(unicode.ToUpper(rune(name[0]))) + name[1:]
}

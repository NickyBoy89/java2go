package goparser

import (
	"strings"
	"gitlab.nicholasnovak.io/snapdragon/java2go/keywords"
)

var functionTable = []string{
	"AssertionError",
	"Println",
	"IndexOutOfBoundsException",
}

var specializedFunctions = map[string]string{
	"System.out.Println": "fmt.Println",
	"System.out.Printf": "fmt.Printf",
}

var replacementTable = map[string]string{
	"void": "",
	"null": "nil",
	"String": "string",
	"double": "float64",
	"float": "float32",
	"short": "int16",
	"char": "rune",
	"boolean": "bool",
	"long": "int64",
	"Integer": "int",
	"Short": "int16",
	"Byte": "byte",
	"Long": "int64",
	"Float": "float32",
	"Double": "float64",
	"Boolean": "bool",
	"Character": "rune",
}

func ReplaceWord(word string) string {
	if val, ok := replacementTable[word]; ok {
		return val
	}
	return word
}

func FormatVariable(in string) string {
	return ToReferenceType(JavaToGoArray(ReplaceWord(in)))
}

// Since java's syntax for an array is something like: int[], and go's is like []int
// we just switch around the brackets
func JavaToGoArray(arr string) string {
	if strings.ContainsRune(arr, '[') {
		// Already in go format
		if arr[0] == '[' {
			return arr
		}

		openingBracket := strings.IndexRune(arr, '[')
		closingBracket := strings.IndexRune(arr, ']')
		return arr[openingBracket:closingBracket + 1] + ReplaceWord(arr[:openingBracket])
	}
	return arr
}

// Note, assumes that the object passed in is in the go array format
func ToReferenceType(in string) string {
	for _, primitive := range keywords.PrimitiveTypes {
		if primitive == in[strings.IndexRune(in, ']') + 1:] { // If the type is an object type
			return in
		}
	}

	// Different asterisk placement for array types
	if lastBrace := strings.LastIndex(in, "]"); lastBrace != -1 {
		return in[:lastBrace + 1] + "*" + in[lastBrace + 1:]
	}

	// Void types return nothing
	if in == "" {
		return in
	}
	return "*" + in
}

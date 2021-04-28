package goparser

import (
	"strings"
)

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

// Since java's syntax for an array is something like: int[], and go's is like []int
// we just switch around the brackets
func JavaToGoArray(arr string) string {
	if strings.ContainsRune(arr, '[') {
		openingBracket := strings.IndexRune(arr, '[')
		closingBracket := strings.IndexRune(arr, ']')
		return arr[openingBracket:closingBracket + 1] + ReplaceWord(arr[:openingBracket])
	}
	return arr
}
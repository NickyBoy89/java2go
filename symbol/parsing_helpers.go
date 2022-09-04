package symbol

import (
	"bytes"
	"go/printer"
	"go/token"
	"unicode"
)

// Uppercase uppercases the first character of the given string
func Uppercase(name string) string {
	return string(unicode.ToUpper(rune(name[0]))) + name[1:]
}

// Lowercase lowercases the first character of the given string
func Lowercase(name string) string {
	return string(unicode.ToLower(rune(name[0]))) + name[1:]
}

// HandleExportStatus is a convenience method for renaming methods that may be
// either public or private, and need to be renamed
func HandleExportStatus(exported bool, name string) string {
	if exported {
		return Uppercase(name)
	}
	return Lowercase(name)
}

// nodeToStr converts any AST node to its string representation
func nodeToStr(node any) string {
	var s bytes.Buffer
	err := printer.Fprint(&s, token.NewFileSet(), node)
	if err != nil {
		panic(err)
	}
	return s.String()
}

func NodeToStr(node any) string {
	return nodeToStr(node)
}

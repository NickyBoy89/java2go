package codeparser

import (
	"gitlab.nicholasnovak.io/snapdragon/java2go/parsetools"
)

var currentReturn string

// Parses the content into separate lines and blocks of execution
func ParseContent(source string) []LineTyper {
	lines := []LineTyper{}

	// Start looping through all the characters in the source
	lastLine, ci := 0, 0
	for ; ci < len(source); ci++ {
		switch rune(source[ci]) {
		// Skip over some pairs of characters (ex: (), "")
		case '(':
			closingParenths := parsetools.IndexOfMatchingParenths(source, ci)
			ci = closingParenths
		case '"':
			closingQuotes := parsetools.FindNextIndexOfCharWithSkip(source[ci + 1:], '"', ``) + ci + 1
			ci = closingQuotes
		case '\'':
			closingSingleQuotes := parsetools.FindNextIndexOfCharWithSkip(source[ci + 1:], '\'', ``) + ci + 1
			ci = closingSingleQuotes
		// These signify that a lines has closed
		case ';':
			lines = append(lines, ParseLine(source[lastLine:ci]))
			lastLine = ci + 1
		case '{':
			closingBrace := parsetools.IndexOfMatchingBrace(source, ci)
			lines = append(lines, ParseLine(source[lastLine:closingBrace + 1]))
			ci = closingBrace
			lastLine = ci + 1
		}
	}

	return lines
}

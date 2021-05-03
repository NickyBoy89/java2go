package codeparser

import (
	// "fmt"
	"strings"

	"gitlab.nicholasnovak.io/snapdragon/java2go/parsing"
)

// Expressions evaluate to a value, statements do not

func ParseContent(sourceData string) []LineTyper {
	contentLines := []LineTyper{}

	lastLine := 0

	ci := 0
	for ; ci < len(sourceData); ci++ {
		switch rune(sourceData[ci]) {
		case ';': // An expression
			contentLines = append(contentLines, ParseLine(sourceData[lastLine:ci]))
		case '(':
			panic("Parenthesies not implemented")
		}
	}
	return contentLines
}

func ParseLine(sourceString string) LineType {
	if equalsIndex := strings.IndexRune(sourceString, '='); equalsIndex != -1 {
		switch parsing.CountRuneWithSkip(strings.Trim(sourceString[:equalsIndex], " \n"), ' ', "<") {
		case 0:
			return LineType{
				name: "AssignVariable",
				Words: map[string]interface{}{
					"VariableName": strings.Trim(sourceString[:equalsIndex], " \n"),
					"Expression": ParseExpression(strings.Trim(sourceString[equalsIndex+1:], " \n")),
				},
			}
		case 1:
			spaceIndex, _ := parsing.FindNextIndexOfCharWithSkip(sourceString, ' ', "<") // Skips generic types
			return LineType{
				name: "CreateAndAssignVariable",
				Words: map[string]interface{}{
					"VariableName": strings.Trim(sourceString[spaceIndex + 1:equalsIndex], " \n"),
					"VariableType": strings.Trim(sourceString[:spaceIndex], " \n"),
					"Expression": ParseExpression(strings.Trim(sourceString[equalsIndex + 1:], " \n")),
				},
			}

		}
	}
	return LineType{}
}

// Assumes that the input has already been stripped
func ParseExpression(source string) []LineType {
	words := []LineType{}

	// Start going through the characters
	ci := 0
	lastWord := 0
	for ; ci < len(source); ci++ {
		switch rune(source[ci]) {
		case ' ':
			words = append(words, LineType{
				name: "LocalVariableOrExpression",
				Words: map[string]interface{}{
					"Expression": source[lastWord:ci],
				},
			})
			lastWord = ci + 1
		case '.':
			words = append(words, LineType{
				name: "RemoteVariableOrExpression",
				Words: map[string]interface{}{
					"RemotePackage": source[lastWord:ci],
					"Expression": source[ci + 1:strings.IndexRune(source[ci:], ' ') + ci],
				},
			})
		case '(':
			panic("Parenthesies in expression not implemented")
		}
	}

	if len(words) == 0 {
		return []LineType{
			LineType{
				name: "LocalVariableOrExpression",
				Words: map[string]interface{}{
					"Expression": strings.Trim(source, " \n"),
				},
			},
		}
	}

	return words
}

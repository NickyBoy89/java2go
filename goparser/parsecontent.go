package goparser

import (
	"encoding/json"
	"fmt"
	"strings"

	"gitlab.nicholasnovak.io/snapdragon/java2go/parsing"
)

// Expressions evaluate to a value, statements do not

func ParseContent(sourceData string) string {
	var content string

	lastLine := 0

	ci := 0
	for ; ci < len(sourceData); ci++ {
		switch rune(sourceData[ci]) {
		case ';': // An expression
			parsedLine, err := json.MarshalIndent(ParseLine(sourceData[lastLine:ci]), "", "  ")
			if err != nil {
				panic(err)
			}
			content += string(parsedLine)
		case '(':
			panic("Parenthesies not implemented")
		}
	}
	return content
}

func ParseLine(sourceString string) LineTyper {
	if equalsIndex := strings.IndexRune(sourceString, '='); equalsIndex != -1 {
		fmt.Println(parsing.CountRuneWithSkip(strings.Trim(sourceString[:equalsIndex], " \n"), ' ', "<"))
		switch parsing.CountRuneWithSkip(strings.Trim(sourceString[:equalsIndex], " \n"), ' ', "<") {
		case 0:
			return LineType{
				name: "AssignVariable",
				Words: map[string]interface{}{
					"VariableName": strings.Trim(sourceString[:equalsIndex], " \n"),
					"Expression":   ParseExpression(strings.Trim(sourceString[equalsIndex+1:], " \n")),
				},
			}
		case 1:
			return LineType{
				name: "CreateAndAssignVariable",
				Words: map[string]interface{}{
					"VariableName": nil,
					"VariableType": nil,
					"Expression":   ParseExpression(),
				},
			}

		}
	}
	return LineType{}
}

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
		case '.':
			panic("Non-package things not implemented")
		case '(':
			panic("Parenthesies in expression not implemented")
		}
	}

	return words
}

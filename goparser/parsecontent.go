package goparser

import (
	"strings"
	"fmt"
	"encoding/json"

	"gitlab.nicholasnovak.io/snapdragon/java2go/parsing"
)

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
		switch parsing.CountRuneWithSkip(sourceString[:equalsIndex], ' ', "<") {
		case 1:
			return LineType{
				Name: "AssignVariable",
				Words: map[string]interface{}{
					"VariableName": strings.Trim(sourceString[], " ")
				}
				Words: strings.Split(sourceString)
			}
		case 2:

		}
	}
	return LineType{}
}

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
			ci = closingParenths + 1
		case '"':
			closingQuotes := parsetools.FindNextIndexOfCharWithSkip(source[ci + 1:], '"', ``) + ci + 1
			ci = closingQuotes + 1
		// These signify that a lines has closed
		case ';':
			lines = append(lines, ParseLine(source[lastLine:ci]))
			lastLine = ci + 1
		case '{':
			closingBrace := parsetools.IndexOfMatchingBrace(source, ci)
			lines = append(lines, ParseLine(source[lastLine:closingBrace + 1]))
			ci = closingBrace + 1
			lastLine = ci
		}
	}

	return lines
}

// func OldParseContent(sourceData string) []LineTyper {
// 	fmt.Printf("Content: %s\n", sourceData)
// 	contentLines := []LineTyper{}
//
// 	lastLine := 0
//
// 	ci := 0
// 	for ; ci < len(sourceData); ci++ { // Separate out the lines in the content
// 		switch rune(sourceData[ci]) {
// 		case ';': // An expression, the line ends with a semicolon
// 			contentLines = append(contentLines, ParseLine(sourceData[lastLine:ci]))
// 			lastLine = ci + 1
// 		case '=': // Expression, to separate out function calls
// 			semicolon := parsetools.FindNextIndexOfChar(sourceData[ci:], ';') + ci
// 			contentLines = append(contentLines, ParseLine(sourceData[lastLine:semicolon]))
// 			ci = semicolon + 1
// 			lastLine = ci
// 		case '"': // Skip over string literals
// 			closingQuotes := parsetools.FindNextIndexOfCharWithSkip(sourceData[ci + 1:], '"', ``) + ci + 1
// 			ci = closingQuotes + 1
// 			lastLine = ci
// 		case ':': // Switch case or loop label
// 			if ci < len(sourceData) - 1 {
// 				if sourceData[ci + 1] == ':' { // Double colon, specifies content of method
// 					// Method reference stops at the first "non-normal" character that is invalid for a method name
// 					nextInvalid := parsetools.IndexOfNextNonNormal(sourceData[ci + 2:]) + ci + 2
//
// 					contentLines = append(contentLines, LineType{
// 						Name: "MethodReference",
// 						Words: map[string]interface{}{
// 							"ClassName": sourceData[lastLine:ci],
// 							"ReferenceName": sourceData[ci + 2:nextInvalid],
// 						},
// 					})
// 				}
// 			}
// 			fmt.Printf("Looking for case in: %s\n", sourceData[lastLine:ci])
// 			// Test to see if there is a "case" in the string
// 			if caseInd := parsetools.IndexWithSkip(sourceData[lastLine:ci], "case", `'"{(`); caseInd != -1 {
// 				// If there is still another case left
// 				if nextCase := parsetools.IndexWithSkip(sourceData[ci:], "case", `'"{(`); nextCase != -1 {
// 					contentLines = append(contentLines, ParseControlFlow(
// 						"case",
// 						strings.Trim(sourceData[lastLine + caseInd + len("case"):ci], " "),
// 						strings.Trim(sourceData[ci + 1:nextCase + ci], " "),
// 					))
// 					ci = nextCase + ci
// 					lastLine = ci
// 				// Tests for a "default" case at the end of the case
// 				} else if parsetools.ContainsWithSkip(sourceData[ci:], "default", `'"{(`) {
// 					contentLines = append(contentLines, ParseControlFlow(
// 						"default",
// 						"", // Default case has no conditions
// 						strings.Trim(sourceData[ci + 1:], " "),
// 					))
// 					ci = len(sourceData)
// 					lastLine = ci
// 				} else {
// 					contentLines = append(contentLines, ParseControlFlow(
// 						"case",
// 						strings.Trim(sourceData[lastLine + caseInd + len("case"):ci], " "),
// 						strings.Trim(sourceData[ci + 1:], " "),
// 					))
// 					ci = len(sourceData)
// 					lastLine = ci
// 				}
// 			} else { // Loop label
// 				contentLines = append(contentLines, LineType{
// 					Name: "ControlLabel",
// 					Words: map[string]interface{}{
// 						"LabelName": strings.Trim(sourceData[lastLine:ci], " "),
// 					},
// 				})
// 				lastLine = ci + 1
// 			}
// 		case '(': // Some type of control flow mechanism (Ex: if, while, etc...) or a function call that does not assign anything
// 			closingParenths := parsetools.IndexOfMatchingParenths(sourceData, ci)
//
// 			fmt.Printf("Inside parenths: %s\n", sourceData[ci:closingParenths + 1])
//
// 			nextChar, ind := parsetools.FindNextNonBlankChar(sourceData[closingParenths + 1:]) // Get the next character after the closing parenths
// 			ind += closingParenths + 1 // Account for the index in relation to the entire string
// 			switch nextChar {
// 			// The next character should be an opening brace, anything else and it is likely an inline method
// 			case '{': // Some type of control flow
// 				closingBrace := parsetools.IndexOfMatchingBrace(sourceData, ind)
// 				contentLines = append(contentLines, ParseControlFlow(
// 					strings.Trim(sourceData[lastLine:ci], " \n"), // The name of the loop (ex: for, if)
// 					sourceData[ci + 1:closingParenths], // The parameters, does not include parenths
// 					sourceData[ind + 1:closingBrace]), // The content of the block
// 				)
// 				ci = closingBrace + 1
// 				lastLine = ci
// 			case ';': // Called function that does not assign anything, semicolon after the closing
// 				ci = ind - 1 // Keep the semicolon so that the expression will be parsed as normal
// 				continue // Should just get parsed out as an expression
// 			case '-':
// 				if sourceData[ind + 1] == '>' { // Found the arrow of a lambda expression (->)
// 					openingBrace := parsetools.FindNextIndexOfCharWithSkip(sourceData[ind:], '{', `'"`) + ind
// 					closingBrace := parsetools.IndexOfMatchingBrace(sourceData, openingBrace)
// 					contentLines = append(contentLines, ParseControlFlow(
// 						"lambdaExpression", // Name
// 						sourceData[ci + 1:closingParenths], // Parameters
// 						sourceData[openingBrace + 1:closingBrace], // Content
// 					))
// 					ci = closingBrace + 1
// 					lastLine = ci
// 				}
// 			default:
// 				contentLines = append(contentLines, LineType{
// 					Name: "ParenthesiedExpression",
// 					Words: map[string]interface{}{
// 						"Expression": ParseExpression(sourceData[ci + 1:closingParenths]),
// 					},
// 				})
// 				ci = closingParenths + 1
// 				lastLine = ci
// 			}
// 		case '{': // Certains other types of control flow (ex: do-while loop)
//
// 			lastBrace := parsetools.IndexOfMatchingBrace(sourceData, ci)
//
// 			fmt.Printf("Found control flow: %s\n", sourceData[lastLine:ci])
//
// 			// If the control flow has a bracket in it, treat it as an implicit creation of an array
// 			// ex: new int[]{1, 2, 3}
// 			if strings.ContainsRune(sourceData[lastLine:ci], '[') {
// 				continue
//
// 			// If the control flow has a parenths in it, treat it as specifying the methods for a class
// 			} else if strings.ContainsRune(sourceData[lastLine:ci], '(') {
// 				panic("method passed")
// 				contentLines = append(contentLines, ParseLine(sourceData[lastLine:lastBrace]))
// 				ci = lastBrace + 1
// 				lastLine = ci
// 				continue
// 			}
//
// 			switch strings.Trim(sourceData[lastLine:ci], " \n") {
// 			case "else":
// 				contentLines = append(contentLines, LineBlock{
// 					Name: "ElseLoop",
// 					Words: make(map[string]interface{}),
// 					Lines: ParseContent(sourceData[ci + 1:lastBrace]),
// 				})
// 				ci = lastBrace + 1
// 				lastLine = ci
// 			case "try":
// 				contentLines = append(contentLines, LineBlock{
// 					Name: "TryBlock",
// 					Words: make(map[string]interface{}),
// 					Lines: ParseContent(sourceData[ci + 1:lastBrace]),
// 				})
// 				ci = lastBrace + 1
// 				lastLine = ci
// 			case "finally": // Finally block always executes when the try block finishes
// 				contentLines = append(contentLines, LineBlock{
// 					Name: "FinallyBlock",
// 					Words: make(map[string]interface{}),
// 					Lines: ParseContent(sourceData[ci + 1:lastBrace]),
// 				})
// 				ci = lastBrace + 1
// 				lastLine = ci
// 			case "do":
// 				contentLines = append(contentLines, LineBlock{
// 					Name: "DoBlock",
// 					Words: make(map[string]interface{}),
// 					Lines: ParseContent(sourceData[ci + 1:lastBrace]),
// 				})
// 				endingSemicolon := parsetools.FindNextIndexOfCharWithSkip(sourceData[lastBrace:], ';', `'"{(`) + lastBrace
// 				// Time for the while expression, becase it doesn't have any lines of its own
// 				contentLines = append(contentLines, ParseControlFlow("do-while", sourceData[strings.IndexRune(sourceData[lastBrace:], '('):endingSemicolon], ""))
// 				ci = endingSemicolon + 1
// 				lastLine = ci
// 			default:
// 				panic("Other type of control flow detected, got [" + strings.Trim(sourceData[lastLine:ci], " \n") + "]")
// 			}
// 		}
// 	}
// 	return contentLines
// }

package codeparser

import (
	// "fmt"
	"strings"

	"gitlab.nicholasnovak.io/snapdragon/java2go/parsetools"
)

var currentReturn string

// Expressions evaluate to a value, statements do not

func ParseContent(sourceData string) []LineTyper {
	contentLines := []LineTyper{}

	lastLine := 0

	ci := 0
	for ; ci < len(sourceData); ci++ { // Separate out the lines in the content
		switch rune(sourceData[ci]) {
		case ';': // An expression, the line ends with a semicolon
			contentLines = append(contentLines, ParseLine(sourceData[lastLine:ci]))
			lastLine = ci + 1
		case '=': // Expression, to separate out function calls
			semicolon := parsetools.FindNextIndexOfChar(sourceData[ci:], ';') + ci
			contentLines = append(contentLines, ParseLine(sourceData[lastLine:semicolon]))
			ci = semicolon + 1
			lastLine = ci
		case '(': // Some type of control flow mechanism (Ex: if, while, etc...) or a function call that does not assign anything
			closingParenths := parsetools.IndexOfMatchingParenths(sourceData, ci)

			nextChar, ind := parsetools.FindNextNonBlankChar(sourceData[closingParenths + 1:]) // Get the next character after the closing parenths
			ind += closingParenths + 1 // Account for the index in relation to the entire string
			switch nextChar {
			case '{': // The next character should be an opening brace, anything else and it is likely an inline method
				closingBrace := parsetools.IndexOfMatchingBrace(sourceData, ind)
				contentLines = append(contentLines, ParseControlFlow(
					strings.Trim(sourceData[lastLine:ci], " \n"), // The name of the loop (ex: for, if)
					sourceData[ci + 1:closingParenths], // The parameters, does not include parenths
					sourceData[ind + 1:closingBrace]), // The content of the block
				)
				ci = closingBrace + 1
				lastLine = ci
			case ';': // Called function that does not assign anything, semicolon after the closing
				ci = ind - 1 // Keep the semicolon so that the expression will be parsed as normal
				continue // Should just get parsed out as an expression
			default:
				panic("Inline method detected found char: [" + string(nextChar) + "]")
			}
		case '{': // Certains other types of control flow (ex: do-while loop)
			lastBrace := parsetools.IndexOfMatchingBrace(sourceData, ci)
			switch strings.Trim(sourceData[lastLine:ci], " \n") {
			case "else":
				contentLines = append(contentLines, LineBlock{
					Name: "ElseLoop",
					Words: make(map[string]interface{}),
					Lines: ParseContent(sourceData[ci + 1:lastBrace]),
				})
				ci = lastBrace + 1
				lastLine = ci
			default:
				panic("Other type of control flow detected, got [" + sourceData[strings.LastIndex(sourceData[lastLine:ci - 1], " "):ci] + "]")
			}
		}
	}
	return contentLines
}

func ParseLine(sourceString string) LineType {
	if equalsIndex := parsetools.FindNextIndexOfCharWithSkip(sourceString, '=', `'"{(`); equalsIndex != -1 { // An equals means an expression
		// The character before the equals, detects things like "+=" and "*=", the compound assignment operators
		// https://www.geeksforgeeks.org/compound-assignment-operators-java/
		switch sourceString[equalsIndex - 1] {
		case '^', '+', '-', '*', '/', '%', '&', '|':
			return LineType{
				Name: "CompoundAssignment",
				Words: map[string]interface{}{
					"Operator": string(sourceString[equalsIndex - 1]),
					"VariableName": strings.Trim(sourceString[:equalsIndex - 1], " \n"), // Just strip one character earlier
					"Expression": ParseExpression(strings.Trim(sourceString[equalsIndex+1:], " \n")),
				},
			}
		}
		switch sourceString[equalsIndex - 2:equalsIndex] { // Three-character compound assignment operators
		case ">>", "<<":
			return LineType{
				Name: "CompoundAssignment",
				Words: map[string]interface{}{
					"Operator": string(sourceString[equalsIndex - 2:equalsIndex]),
					"VariableName": strings.Trim(sourceString[:equalsIndex - 2], " \n"), // Just strip two characters earlier
					"Expression": ParseExpression(strings.Trim(sourceString[equalsIndex+1:], " \n")),
				},
			}
		}
		if sourceString[equalsIndex - 3:equalsIndex] == ">>>" { // Compound right-shift filled 0 assignment operator
			return LineType{
				Name: "CompoundAssignment",
				Words: map[string]interface{}{
					"Operator": ">>>",
					"VariableName": strings.Trim(sourceString[:equalsIndex - 3], " \n"), // Just strip three characters earlier
					"Expression": ParseExpression(strings.Trim(sourceString[equalsIndex+1:], " \n")),
				},
			}
		}
		switch parsetools.CountRuneWithSkip(strings.Trim(sourceString[:equalsIndex], " \n"), ' ', "<") { // Counts the space between the type of variable and variable name (ex: int value)
		case 0: // 0 spaces means that there are no spaces, and the variable is just being re-assigned
			return LineType{
				Name: "AssignVariable",
				Words: map[string]interface{}{
					"VariableName": ParseExpression(strings.Trim(sourceString[:equalsIndex], " \n")),
					"Expression": []LineType{ParseLine(strings.Trim(sourceString[equalsIndex+1:], " \n"))},
				},
			}
		case 1: // 1 space means that the variable is being declared and set to a value
			spaceIndex := parsetools.FindNextIndexOfCharWithSkip(sourceString, ' ', "<") // Skips generic types
			currentReturn = strings.Trim(sourceString[:spaceIndex], " \n") // Global variable
			return LineType{
				Name: "CreateAndAssignVariable",
				Words: map[string]interface{}{
					"VariableName": ParseExpression(strings.Trim(sourceString[spaceIndex + 1:equalsIndex], " \n")),
					"VariableType": currentReturn,
					"Expression": ParseExpression(strings.Trim(sourceString[equalsIndex + 1:], " \n")),
				},
			}

		}
	}
	words := parsetools.DiscardBlankStrings(strings.Split(sourceString, " "))
	switch words[0] {
	case "return": // The return keyword for a line
		if len(words) == 1 { // Naked return, no other expressions after
			return LineType{
				Name: "ReturnStatement",
				Words: map[string]interface{}{
					"Expression": []LineType{},
				},
			}
		}
		return LineType{
			Name: "ReturnStatement",
			Words: map[string]interface{}{
				"Expression": ParseExpression(sourceString[len(words[0]) + 1:]),
			},
		}
	case "new":
		return LineType{
			Name: "NewConstructor",
			Words: map[string]interface{}{
				"Expression": ParseExpression(sourceString[len(words[0]) + 1:])[0],
			},
		}
	case "throw":
		return LineType{
			Name: "ThrowException",
			Words: map[string]interface{}{
				"Expression": ParseLine(sourceString[len(words[0]) + 1:]), // Re-parse the line
			},
		}
	}

	return LineType{
		Name: "GenericLine",
		Words: map[string]interface{}{
			"Statement": ParseExpression(sourceString),
		},
	}
}

// Really simply parses a variable of the form (Type Variable), and handles everything else
func ParseVariableAndType(source string) (variableType, variableName string) {
	spaceIndex := parsetools.FindNextIndexOfCharWithSkip(source, ' ', "<") // The space between the type and the variable, accounts for generics
	return strings.Trim(source[:spaceIndex], " \n"), strings.Trim(source[spaceIndex + 1:], " \n")
}

// Assumes that the input has already been stripped
func ParseExpression(source string) []LineType {
	words := []LineType{}

	firstSpace := parsetools.FindNextIndexOfCharWithSkip(source, ' ', `'"{(`)
	if firstSpace != -1 {
		switch source[:firstSpace] {
		case "new":
			// Array constructor
			if openingBracket := strings.IndexRune(source[firstSpace:], '[') + firstSpace; openingBracket - firstSpace != -1 { // "new" is only constructing an array, and does not create a generator function
				closingBracket := strings.IndexRune(source[openingBracket:], ']') + openingBracket
				return []LineType{
					LineType{
						Name: "ConstructArray",
						Words: map[string]interface{}{
							"ArrayType": strings.Trim(source[firstSpace:openingBracket], " "),
							"InitialSize": source[openingBracket + 1:closingBracket],
						},
					},
				}
			}
			return []LineType{
				LineType{
					Name: "NewConstructor",
					Words: map[string]interface{}{
						"Expression": ParseExpression(strings.Trim(source[firstSpace:], " "))[0],
					},
				},
			}
		}
	}

	switch source[0] {
	case '{':  // Means implicitly filling an array
		return []LineType{
			LineType{
				Name: "ImplicitArrayAssignment",
				Words: map[string]interface{}{
					"ArrayType": currentReturn, // Gets the current return type from a global, becuse I didn't think the best way for that would be passing it in as a parameter
					"Elements": ParseCommaSeparatedValues(source[1:len(source) - 1]),
				},
			},
		}
	}

	// Start going through the characters
	ci := 0
	lastWord := 0
	for ; ci < len(source); ci++ {
		switch rune(source[ci]) {
		case ' ': // If a divider between expressions is detected (NOTE: '+', '-', '/' are all valid, but I think not needed)
			words = append(words, LineType{
				Name: "LocalVariableOrExpression",
				Words: map[string]interface{}{
					"Expression": strings.Trim(source[lastWord:ci], " "),
				},
			})
			lastWord = ci + 1
		case '.': // A dot signals another package, should not interfere with the declaration of a float also, because the package should always come before
			if spaceInd := parsetools.FindNextIndexOfCharWithSkip(source[ci:], ' ', `"'{(`); spaceInd != -1 { // If a space exists (not the last expression)
				endSpace := strings.IndexRune(source[ci:], ' ') + ci
				words = append(words, LineType{
					Name: "RemoteVariableOrExpression",
					Words: map[string]interface{}{
						"RemotePackage": source[lastWord:ci],
						"Expression": ParseExpression(source[ci + 1:endSpace]),
					},
				})
				ci = endSpace + 1
				lastWord = ci
			} else { // The last expression in the expression
				words = append(words, LineType{
					Name: "RemoteVariableOrExpression",
					Words: map[string]interface{}{
						"RemotePackage": source[lastWord:ci],
						"Expression": ParseExpression(source[ci + 1:]),
					},
				})
				ci = len(source) // Should just break out
				lastWord = ci
			}
		case '(': // Function call
			closingParenths := parsetools.IndexOfMatchingParenths(source, ci)
			words = append(words, LineType{
				Name: "FunctionCall",
				Words: map[string]interface{}{
					"FunctionName": strings.Trim(source[lastWord:ci], " "),
					"Parameters": ParseCommaSeparatedValues(source[ci + 1:closingParenths]),
				},
			})
			ci = closingParenths + 1
			lastWord = ci
		case '[': // Access a specific element of an array
			closingBrace := strings.IndexRune(source[ci:], ']') + ci
			words = append(words, LineType{
				Name: "AccessArrayElement",
				Words: map[string]interface{}{
					"ArrayName": strings.Trim(source[lastWord:ci], " "),
					"Index": source[ci + 1:closingBrace],
				},
			})
			ci = closingBrace + 1
			lastWord = ci
		// Start getting into the literals (ex: "yes" is a string literal)
		case '"': // String literal
			closingQuotes := parsetools.FindNextIndexOfCharWithSkip(source[ci + 1:], '"', ``) + ci + 1
			words = append(words, LineType{
				Name: "StringLiteral",
				Words: map[string]interface{}{
					"String": source[ci:closingQuotes + 1],
				},
			})
			ci = closingQuotes + 1
			lastWord = ci
		case '\'': // Rune literal
			closingSingQuotes := strings.IndexRune(source[ci + 1:], '\'') + ci + 1
			words = append(words, LineType{
				Name: "RuneLiteral",
				Words: map[string]interface{}{
					"Rune": source[ci + 1:closingSingQuotes],
				},
			})
			ci = closingSingQuotes + 1
			lastWord = ci
		}
	}

	if lastWord != len(source) { // If there is still one more expression left
		words = append(words, LineType{
			Name: "LocalVariableOrExpression",
			Words: map[string]interface{}{
				"Expression": source[lastWord:],
			},
		})
	}

	if len(words) == 0 { // If no word has been detected
		return []LineType{
			LineType{
				Name: "LocalVariableOrExpression",
				Words: map[string]interface{}{
					"Expression": strings.Trim(source, " \n"),
				},
			},
		}
	}

	return words
}

func ParseControlFlow(controlBlockname, parameters, source string) LineTyper {
	switch controlBlockname {
	case "for": // For loop, can be a normal for loop, or a for-each loop
		statement := ParseStatements(parameters)
		return LineBlock{
			Name: "ForLoop",
			Words: map[string]interface{}{
				"Initializer": statement["Initializer"],
				"Conditional": statement["Conditional"],
				"Incrementer": statement["Incrementer"],
			},
			Lines: ParseContent(strings.Trim(source, " \n")),
		}
	case "if":
		return LineBlock{
			Name: "IfStatement",
			Words: map[string]interface{}{
				"Condition": ParseExpression(parameters),
			},
			Lines: ParseContent(strings.Trim(source, " \n")),
		}
	case "else if":
		return LineBlock{
			Name: "ElseIfStatement",
			Words: map[string]interface{}{
				"Condition": ParseExpression(parameters),
			},
			Lines: ParseContent(strings.Trim(source, " \n")),
		}
	default:
		panic("Unrecognized loop type, got [" + controlBlockname + "]")
	}
}

// Parses out the statements in a for loop (Initializer, Conditional, Incrementer)
func ParseStatements(source string) map[string]interface{} {
	if colonInd := parsetools.FindNextIndexOfCharWithSkip(source, ':', `"'{(`); colonInd == -1 { // No colon found, not a for-each loop
		semicolons := parsetools.FindAllIndexesOfChar(source, ';')
		if len(semicolons) != 2 { // Should just be 2 semicolons
			panic("Invalid number of semicolons in for loop")
		}
		init := strings.Trim(source[:semicolons[0]], " \n")
		cond := strings.Trim(source[semicolons[0] + 1:semicolons[1]], " \n")
		incr := strings.Trim(source[semicolons[1] + 1:], " \n")
		return map[string]interface{}{
			"Initializer": ParseLine(init),
			"Conditional": ParseExpression(cond),
			"Incrementer": ParseLine(incr),
		}
	} else { // An else block here just so that I can use the colonInd variable
		// For-each loop
		// Format: (type Counter : Iterable)
		return map[string]interface{}{
			"Counter": strings.Trim(source[:colonInd], " \n"),
			"Iterable": strings.Trim(source[colonInd + 1:], " \n"),
		}
	}
}

func ParseCommaSeparatedValues(source string) [][]LineType {
	elementSeparators := parsetools.FindAllIndexesOfChar(source, ',')

	arrayElements := [][]LineType{}

	carrier := 0
	for _, sep := range elementSeparators {
		arrayElements = append(arrayElements, ParseExpression(strings.Trim(source[carrier:sep], " ")))
		carrier = sep + 1 // Skip the comma
	}
	if carrier != len(source) { // If there is still one more element not found yet
		arrayElements = append(arrayElements, ParseExpression(strings.Trim(source[carrier:], " ")))
	}

	return arrayElements
}

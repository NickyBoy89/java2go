package codeparser

import (
	"fmt"
	"strings"
	"unicode"

	"gitlab.nicholasnovak.io/snapdragon/java2go/parsetools"
)

var currentReturn string

// Expressions evaluate to a value, statements do not

func ParseContent(sourceData string) []LineTyper {
	fmt.Printf("Content: %s\n", sourceData)
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
		case '"': // Skip over string literals
			closingQuotes := parsetools.FindNextIndexOfCharWithSkip(sourceData[ci + 1:], '"', ``) + ci + 1
			ci = closingQuotes + 1
			lastLine = ci
		case ':': // Switch case or loop label
			if ci < len(sourceData) - 1 {
				if sourceData[ci + 1] == ':' { // Double colon, specifies content of method
					// Method reference stops at the first "non-normal" character that is invalid for a method name
					nextInvalid := parsetools.IndexOfNextNonNormal(sourceData[ci + 2:]) + ci + 2

					contentLines = append(contentLines, LineType{
						Name: "MethodReference",
						Words: map[string]interface{}{
							"ClassName": sourceData[lastLine:ci],
							"ReferenceName": sourceData[ci + 2:nextInvalid],
						},
					})
				}
			}
			fmt.Printf("Looking for case in: %s\n", sourceData[lastLine:ci])
			// Test to see if there is a "case" in the string
			if caseInd := parsetools.IndexWithSkip(sourceData[lastLine:ci], "case", `'"{(`); caseInd != -1 {
				// If there is still another case left
				if nextCase := parsetools.IndexWithSkip(sourceData[ci:], "case", `'"{(`); nextCase != -1 {
					contentLines = append(contentLines, ParseControlFlow(
						"case",
						strings.Trim(sourceData[lastLine + caseInd + len("case"):ci], " "),
						strings.Trim(sourceData[ci + 1:nextCase + ci], " "),
					))
					ci = nextCase + ci
					lastLine = ci
				// Tests for a "default" case at the end of the case
				} else if parsetools.ContainsWithSkip(sourceData[ci:], "default", `'"{(`) {
					contentLines = append(contentLines, ParseControlFlow(
						"default",
						"", // Default case has no conditions
						strings.Trim(sourceData[ci + 1:], " "),
					))
					ci = len(sourceData)
					lastLine = ci
				} else {
					contentLines = append(contentLines, ParseControlFlow(
						"case",
						strings.Trim(sourceData[lastLine + caseInd + len("case"):ci], " "),
						strings.Trim(sourceData[ci + 1:], " "),
					))
					ci = len(sourceData)
					lastLine = ci
				}
			} else { // Loop label
				contentLines = append(contentLines, LineType{
					Name: "ControlLabel",
					Words: map[string]interface{}{
						"LabelName": strings.Trim(sourceData[lastLine:ci], " "),
					},
				})
				lastLine = ci + 1
			}
		case '(': // Some type of control flow mechanism (Ex: if, while, etc...) or a function call that does not assign anything
			closingParenths := parsetools.IndexOfMatchingParenths(sourceData, ci)

			fmt.Printf("Inside parenths: %s\n", sourceData[ci:closingParenths + 1])

			nextChar, ind := parsetools.FindNextNonBlankChar(sourceData[closingParenths + 1:]) // Get the next character after the closing parenths
			ind += closingParenths + 1 // Account for the index in relation to the entire string
			switch nextChar {
			// The next character should be an opening brace, anything else and it is likely an inline method
			case '{': // Some type of control flow
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
			case '-':
				if sourceData[ind + 1] == '>' { // Found the arrow of a lambda expression (->)
					openingBrace := parsetools.FindNextIndexOfCharWithSkip(sourceData[ind:], '{', `'"`) + ind
					closingBrace := parsetools.IndexOfMatchingBrace(sourceData, openingBrace)
					contentLines = append(contentLines, ParseControlFlow(
						"lambdaExpression", // Name
						sourceData[ci + 1:closingParenths], // Parameters
						sourceData[openingBrace + 1:closingBrace], // Content
					))
					ci = closingBrace + 1
					lastLine = ci
				}
			default:
				contentLines = append(contentLines, LineType{
					Name: "ParenthesiedExpression",
					Words: map[string]interface{}{
						"Expression": ParseExpression(sourceData[ci + 1:closingParenths]),
					},
				})
				ci = closingParenths + 1
				lastLine = ci
			}
		case '{': // Certains other types of control flow (ex: do-while loop)

			lastBrace := parsetools.IndexOfMatchingBrace(sourceData, ci)

			fmt.Printf("Found control flow: %s\n", sourceData[lastLine:ci])

			// If the control flow has a bracket in it, treat it as an implicit creation of an array
			// ex: new int[]{1, 2, 3}
			if strings.ContainsRune(sourceData[lastLine:ci], '[') {
				continue

			// If the control flow has a parenths in it, treat it as specifying the methods for a class
			} else if strings.ContainsRune(sourceData[lastLine:ci], '(') {
				panic("method passed")
				contentLines = append(contentLines, ParseLine(sourceData[lastLine:lastBrace]))
				ci = lastBrace + 1
				lastLine = ci
				continue
			}

			switch strings.Trim(sourceData[lastLine:ci], " \n") {
			case "else":
				contentLines = append(contentLines, LineBlock{
					Name: "ElseLoop",
					Words: make(map[string]interface{}),
					Lines: ParseContent(sourceData[ci + 1:lastBrace]),
				})
				ci = lastBrace + 1
				lastLine = ci
			case "try":
				contentLines = append(contentLines, LineBlock{
					Name: "TryBlock",
					Words: make(map[string]interface{}),
					Lines: ParseContent(sourceData[ci + 1:lastBrace]),
				})
				ci = lastBrace + 1
				lastLine = ci
			case "finally": // Finally block always executes when the try block finishes
				contentLines = append(contentLines, LineBlock{
					Name: "FinallyBlock",
					Words: make(map[string]interface{}),
					Lines: ParseContent(sourceData[ci + 1:lastBrace]),
				})
				ci = lastBrace + 1
				lastLine = ci
			case "do":
				contentLines = append(contentLines, LineBlock{
					Name: "DoBlock",
					Words: make(map[string]interface{}),
					Lines: ParseContent(sourceData[ci + 1:lastBrace]),
				})
				endingSemicolon := parsetools.FindNextIndexOfCharWithSkip(sourceData[lastBrace:], ';', `'"{(`) + lastBrace
				// Time for the while expression, becase it doesn't have any lines of its own
				contentLines = append(contentLines, ParseControlFlow("do-while", sourceData[strings.IndexRune(sourceData[lastBrace:], '('):endingSemicolon], ""))
				ci = endingSemicolon + 1
				lastLine = ci
			default:
				panic("Other type of control flow detected, got [" + strings.Trim(sourceData[lastLine:ci], " \n") + "]")
			}
		}
	}
	return contentLines
}

func ParseLine(sourceString string) LineType {
	// fmt.Println(sourceString)
	if equalsIndex := parsetools.FindNextIndexOfCharWithSkip(sourceString, '=', `'"{(`); equalsIndex != -1 && equalsIndex - 1 > 0 && sourceString[equalsIndex - 1] != '!' { // An equals means an expression
		// The character before the equals, detects things like "+=" and "*=", the compound assignment operators
		// https://www.geeksforgeeks.org/compound-assignment-operators-java/
		if equalsIndex - 3 >= 0 {
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
		}
		// Counts the space between the type of variable and variable name (ex: int value)
		switch parsetools.CountRuneWithSkip(strings.Trim(sourceString[:equalsIndex], " \n"), ' ', "<") {
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
	if source == "" { // No expression
		return []LineType{}
	}
	// fmt.Printf("Expression: %s\n", source)
	words := []LineType{}

	firstSpace := parsetools.FindNextIndexOfCharWithSkip(source, ' ', `'"{(`)
	if firstSpace != -1 {
		switch source[:firstSpace] {
		case "new":

			// If we are creating an array (new int[]{1, 2, 3} or new int[0]), this doesn't make sense to have a newConstructor for it
			// So it just outputs a ConstructArray type
			if openingBracket := strings.IndexRune(source[firstSpace:], '['); openingBracket != -1 {
				lastOpeningBracket := strings.LastIndex(source[firstSpace:], "[")
				closingBracket := strings.LastIndex(source[openingBracket:], "]") + openingBracket

				// If there is no implicit array assignment (ex: new int[0])
				if closingBracket == len(source) - 1 {
					return []LineType{
						LineType{
							Name: "ConstructArray",
							Words: map[string]interface{}{
								"ArrayType": strings.Trim(source[firstSpace:firstSpace + openingBracket], " "),
								"InitialSize": source[firstSpace + openingBracket + 1:closingBracket],
							},
						},
					}
				}

				// Some implicit array assignment
				currentReturn = strings.Trim(source[firstSpace:firstSpace + lastOpeningBracket], " ")
				return []LineType{
					LineType{
						Name: "ConstructArrayWithImplicit",
						Words: map[string]interface{}{
							"ArrayType": currentReturn,
							"Elements": ParseExpression(strings.Trim(source[closingBracket + 1:], " ")),
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

		// If the array is an array type, then remove one level of array from it
		tempReturn := currentReturn
		if strings.ContainsRune(currentReturn, '[') {
			currentReturn = strings.Replace(currentReturn, "[]", "", 1)
		}

		return []LineType{
			LineType{
				Name: "ImplicitArrayAssignment",
				Words: map[string]interface{}{
					"ArrayType": tempReturn, // Gets the current return type from a global, becuse I didn't think the best way for that would be passing it in as a parameter
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
				words = append(words, LineType{
					Name: "RemoteVariableOrExpression",
					Words: map[string]interface{}{
						"RemotePackage": source[lastWord:ci],
						"Expression": ParseExpression(source[ci + 1:spaceInd + ci]),
					},
				})
				ci += spaceInd + 1
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
		case '(': // Function call or type assertion
			closingParenths := parsetools.IndexOfMatchingParenths(source, ci)

			fmt.Printf("Inside parenths: %s\n", source[ci:closingParenths + 1])

			// Look at the character before the function
			//to test whether it is a function or type assertion
			if ci == 0 || !unicode.IsLetter(rune(source[ci - 1])) && !unicode.IsDigit(rune(source[ci - 1])){ // If function call ends with number or letter

				// Look at the character after the function to determine if it is just a statement in parenthesies
				var nextChar rune
				if len(source[closingParenths + 1:]) != 0 { // Bounds-check to see if there is a non-blank character after the parenths statement
					nextChar, _ = parsetools.FindNextNonBlankChar(source[closingParenths + 1:])
				}
				if unicode.IsLetter(rune(nextChar)) || unicode.IsDigit(rune(nextChar)) { // Letter after the parenthesies means a type assertion
					words = append(words, LineType{
						Name: "TypeAssertion",
						Words: map[string]interface{}{
							"AssertedType": strings.Trim(source[ci + 1:closingParenths], " "),
						},
					})
				} else { // Not a type assertion, some other form of operator means that this is a parenthesied statement
					words = append(words, LineType{
						Name: "ParenthesiedExpression",
						Words: map[string]interface{}{
							"Expression": ParseExpression(strings.Trim(source[ci + 1:closingParenths], " ")),
						},
					})
				}
			} else {
				words = append(words, LineType{
					Name: "FunctionCall",
					Words: map[string]interface{}{
						"FunctionName": strings.Trim(source[lastWord:ci], " "),
						"Parameters": ParseCommaSeparatedValues(source[ci + 1:closingParenths]),
					},
				})
			}
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
		case '=':
			switch source[ci + 1] {
			case '=':
				words = append(words, LineType{
					Name: "ComparisonOperator",
					Words: map[string]interface{}{
						"Operator": "==",
					},
				})
				ci += 2
				lastWord = ci
			}
		case '?': // If there is a bare question mark in the expression, then there likely is a ternary operator
			colonInd := parsetools.FindNextIndexOfCharWithSkip(source[ci:], ':', `'"{(`) + ci
			words = append(words, LineType{
				Name: "TernaryOperator",
				Words: map[string]interface{}{
					"TrueExpression": ParseExpression(strings.Trim(source[ci + 1:colonInd], " ")),
					"FalseExpression": ParseExpression(strings.Trim(source[colonInd + 1:], " ")),
				},
			})
			ci = len(source)
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
	case "for": // For loop, can be a normal for loop, or a for-each loop (enhanced for loop)
		if colonInd := strings.IndexRune(parameters, ':'); colonInd != -1 { // An enhanced for loop will have a colon in it
			declarationWords := parsetools.DiscardBlankStrings(strings.Split(parameters[:colonInd], " "))
			return LineBlock{
				Name: "EnhancedForLoop",
				Words: map[string]interface{}{
					"DeclarationType": declarationWords[0],
					"DeclarationName": declarationWords[1],
					"Iterable": ParseExpression(strings.Trim(parameters[colonInd + 1:], " ")),
				},
				Lines: ParseContent(strings.Trim(source, " \n")),
			}
		}
		statement := ParseForLoopStatements(parameters)
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
	case "while":
		return LineBlock{
			Name: "WhiteStatement",
			Words: map[string]interface{}{
				"Condition": ParseExpression(parameters),
			},
			Lines: ParseContent(strings.Trim(source, " \n")),
		}
	case "catch":
		return LineBlock{
			Name: "CatchBlock",
			Words: map[string]interface{}{
				"Exception": ParseExpression(parameters),
			},
			Lines: ParseContent(strings.Trim(source, " \n")),
		}
	case "switch":
		return LineBlock{
			Name: "SwitchExpression",
			Words: map[string]interface{}{
				"SwitchExpression": ParseExpression(parameters),
			},
			Lines: ParseContent(strings.Trim(source, " ")),
		}
	case "case":
		return LineBlock{
			Name: "SwitchCase",
			Words: map[string]interface{}{
				"Case": parameters,
			},
			Lines: ParseContent(strings.Trim(source, " ")),
		}
	case "default": // Default switch statement
		return LineBlock{
			Name: "DefaultCase",
			Words: make(map[string]interface{}),
			Lines: ParseContent(strings.Trim(source, " ")),
		}
	case "do-while":
		return LineBlock{
			Name: "DoWhileStatement",
			Words: map[string]interface{}{
				"Statement": ParseExpression(parameters),
			},
		}
	case "lambdaExpression":
		return LineBlock{
			Name: "LambdaExpression",
			Words: map[string]interface{}{
				"Parameters": ParseCommaSeparatedValues(parameters),
			},
			Lines: ParseContent(strings.Trim(source, " ")),
		}
	default:
		return LineBlock{
			Name: "ImplicitObjectCreation",
			Words: map[string]interface{}{
				"MethodLine": ParseLine(controlBlockname),
			},
			Lines: ParseContent(strings.Trim(source, " ")),
		}
	}
}

// Parses out the statements in a for loop (Initializer, Conditional, Incrementer)
func ParseForLoopStatements(source string) map[string]interface{} {
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

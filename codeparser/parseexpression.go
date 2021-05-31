package codeparser

import (
  "strings"
  "unicode"
  "fmt"

  "gitlab.nicholasnovak.io/snapdragon/java2go/parsetools"
)

// Parses an expression (anything that returns a value)
func ParseExpression(source string) []LineType {
	if source == "" { // No expression
		return []LineType{}
	}
	fmt.Printf("Expression: [%s]\n", source)
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
    fmt.Println("Brackets")
			closingBrace := parsetools.IndexOfMatchingBrackets(source, ci)
			words = append(words, LineType{
				Name: "AccessArrayElement",
				Words: map[string]interface{}{
					"ArrayName": strings.Trim(source[lastWord:ci], " "),
					"Index": source[ci + 1:closingBrace],
				},
			})
			ci = closingBrace
			lastWord = ci + 1
		case '=':
      if len(source) > 1 {
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

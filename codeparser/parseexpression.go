package codeparser

import (
  "strings"
  "unicode"

  "gitlab.nicholasnovak.io/snapdragon/java2go/parsetools"
)

// Parses an expression (anything that returns a value)
func ParseExpression(source string) []LineType {
  // fmt.Printf("Expression: %s\n", source)
	if source == "" { // No expression
		return []LineType{}
	}
	words := []LineType{}

	lastWord, ci := 0, 0
	for ; ci < len(source); ci++ {
		switch rune(source[ci]) {
		case ' ': // Words
      switch strings.Trim(source[lastWord:ci], " ") {
      case "new":
        // If the constructor creates an array, handle that as a custom constructor type
        // ex: new int[0]
        if openingBracket := parsetools.FindNextIndexOfCharWithSkip(source[ci:], '[', `"'`) + ci; openingBracket - ci != -1 {
          closingBracket := parsetools.IndexOfMatchingBrackets(source, openingBracket)
          bracketContents := ContentOfBrackets(source[ci:closingBracket + 1])
          return []LineType{
            LineType{
              Name: "ConstructArray",
              Words: map[string]interface{}{
                "ArrayType": strings.Trim(source[ci:closingBracket], " "),
                "InitialSize": bracketContents[len(bracketContents) - 1], // Last bracket in bracket contents
              },
            },
          }
        }

        // New constructor
        return []LineType{
          LineType{
            Name: "NewConstructor",
            Words: map[string]interface{}{
              "Expression": ParseExpression(strings.Trim(source[ci:], " "))[0], // Should only return one constructor as a function
            },
          },
        }
      }
			words = append(words, LineType{
				Name: "LocalVariableOrExpression",
				Words: map[string]interface{}{
					"Expression": strings.Trim(source[lastWord:ci], " "),
				},
			})
			lastWord = ci + 1
    case '{': // Implicit filling of array (ex: new int[] = {1, 2, 3})
      closingBrace := parsetools.IndexOfMatchingBrace(source, ci)
      words = append(words, LineType{
        Name: "ImplicitArrayAssignment",
        Words: map[string]interface{}{
          "ArrayType": currentReturn,
          "Elements": ParseCommaSeparatedValues(source[ci + 1:closingBrace]),
        },
      })
      ci = closingBrace
      lastWord = ci + 1
		case '.': // A dot means another package, or aspect of variable (method, class variable)
			if spaceInd := parsetools.FindNextIndexOfCharWithSkip(source[ci:], ' ', `"'{([`); spaceInd != -1 { // If a space exists (not the last expression)
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
		case '(': // Function call, type assertion, normal parenthesies, lambda parameters
			closingParenths := parsetools.IndexOfMatchingParenths(source, ci)

      // Start filtering out what one of these four options the parenthesies is

      // Start with the character before the parenthesies
      var charBeforeParenthesies rune
      if ci > 0 { // If the parenthesies is not at the start of the expression
        charBeforeParenthesies = rune(source[ci - 1])
      }
      // Look at the character after the parenthesies
      var charAfterParenthsInd int
      var charAfterParenths rune
      if closingParenths < len(source) - 1 { // If there is a character after the parenthesies
        charAfterParenths, charAfterParenthsInd = parsetools.FindNextNonBlankChar(source[closingParenths + 1:])
      }

      // A letter or number before will mean a function (ex: method_1234())
      if charBeforeParenthesies != 0 && unicode.IsLetter(rune(charBeforeParenthesies)) || unicode.IsDigit(rune(charBeforeParenthesies)) {
        words = append(words, LineType{
          Name: "FunctionCall",
          Words: map[string]interface{}{
            "FunctionName": strings.Trim(source[lastWord:ci], " "),
            "Parameters": ParseCommaSeparatedValues(source[ci + 1:closingParenths]),
          },
        })
        ci = closingParenths
        lastWord = ci + 1
      } else { // Determine between type assertion, parenthesies, and lambda parameters
        // If the first character after the parenthesies is a normal value, then it is a type assertion (ex: (int)1.0)
        if charAfterParenths != 0 && unicode.IsLetter(rune(charAfterParenths)) || unicode.IsLetter(rune(charAfterParenths)) {
          words = append(words, LineType{
            Name: "TypeAssertion",
            Words: map[string]interface{}{
              "AssertedType": strings.Trim(source[ci + 1:closingParenths], " "),
            },
          })
          ci = closingParenths
          lastWord = ci + 1
        // Detect a lambda
        } else if charAfterParenths == '-' && source[charAfterParenthsInd + closingParenths + 2] == '>' {
          openingBraces := strings.IndexRune(source[closingParenths:], '{') + closingParenths
          closingBraces := parsetools.IndexOfMatchingBrace(source, openingBraces)
          words = append(words, LineType{
            Name: "LambdaExpression",
            Words: map[string]interface{}{
              "Parameters": ParseCommaSeparatedValues(strings.Trim(source[ci + 1:closingParenths], " ")),
              "Lines": ParseContent(strings.Trim(source[openingBraces + 1:closingBraces], " ")),
            },
          })
          ci = closingBraces
          lastWord = ci + 1
        } else { // Normal parenthesies
          words = append(words, LineType{
            Name: "ParenthesiedExpression",
            Words: map[string]interface{}{
              "Expression": ParseExpression(strings.Trim(source[ci + 1:closingParenths], " ")),
            },
          })
          ci = closingParenths
          lastWord = ci + 1
        }
      }
    case ':': // Loop label, or method reference operator
      if ci < len(source) - 1 && source[ci + 1] == ':' { // Double colon
        referenceEnd := parsetools.IndexOfNextNonNormal(source[ci + 2:]) + ci + 2
        words = append(words, LineType{
          // A method reference calls a method by referring to its class directly
          Name: "MethodReference",
          Words: map[string]interface{}{
            "MethodClass": source[lastWord:ci],
            "MethodName": source[ci + 2:referenceEnd],
          },
        })
        ci = referenceEnd
        lastWord = ci + 1
      } else {
        // Ignore a "case" or "default", as well as a non-space character before the colon
        if !strings.Contains(source[lastWord:ci], "case") && !strings.Contains(source[lastWord:ci], "default") && ci > len(source) - 1 && source[ci - 1] != ' ' {
          words = append(words, LineType{
            Name: "ContentLabel",
            Words: map[string]interface{}{
              "LabelName": strings.Trim(source[lastWord:ci], " "),
            },
          })
          lastWord = ci + 1
        }
      }
		case '[': // Access a specific element of an array
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
      // If there is no colon (no ternary false statement)
      if colonInd - ci == -1 {
        words = append(words, LineType{
  				Name: "TernaryOperator",
  				Words: map[string]interface{}{
  					"TrueExpression": ParseExpression(strings.Trim(source[ci + 1:], " ")),
  					"FalseExpression": []LineType{},
  				},
  			})
      } else {
        words = append(words, LineType{
          Name: "TernaryOperator",
          Words: map[string]interface{}{
            "TrueExpression": ParseExpression(strings.Trim(source[ci + 1:colonInd], " ")),
            "FalseExpression": ParseExpression(strings.Trim(source[colonInd + 1:], " ")),
          },
        })
      }
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

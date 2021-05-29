package codeparser

import (
  "strings"
  "fmt"

  "gitlab.nicholasnovak.io/snapdragon/java2go/parsetools"
)

func ParseLine(source string) LineTyper {

  fmt.Printf("Line content: %s\n", source)

  currentWords := []LineType{}

  lastLine, ci := 0, 0
  for ; ci < len(source); ci++ {
    switch rune(source[ci]) {
    case '<': // Skip over angles
      closingAngle := parsetools.IndexOfMatchingAngles(source, ci)
      ci = closingAngle + 1
      lastLine = ci
    case ' ': // A word
      switch source[lastLine:ci] {
      case "": // Empty string, ignore
      case "return": // The return keyword for a line
    		return LineType{
    			Name: "ReturnStatement",
    			Words: map[string]interface{}{
    				"Expression": ParseExpression(source[lastLine:ci]),
    			},
    		}
    	case "throw":
    		return LineType{
    			Name: "ThrowException",
    			Words: map[string]interface{}{
    				"Expression": ParseExpression(source[lastLine:ci]),
    			},
    		}
      default:
        fmt.Printf("Word in front of brackets: %s\n", source[lastLine:ci])
        currentWords = append(currentWords, ParseExpression(source[lastLine:ci])...)
        lastLine = ci + 1
      }
    case '=':
      // Filters out compound assignment operator (*=, +=) and such
      if ci > 1 {
        switch source[ci - 1] {
        // Not equals operator
        case '!':
          return LineType{
            Name: "NotEquals",
            Words: map[string]interface{}{
              "LeftExpression": "",
              "RightExpression": "",
            },
          }
        // The character before the equals, detects things like "+=" and "*=", the compound assignment operators
    		// https://www.geeksforgeeks.org/compound-assignment-operators-java/
        case '^', '+', '-', '*', '/', '%', '&', '|':
          return LineType{
            Name: "CompoundAssignment",
            Words: map[string]interface{}{
              "Operator": string(source[ci - 1]),
              "VariableName": strings.Trim(source[:ci - 1], " \n"), // Just strip one character earlier
              "Expression": ParseExpression(strings.Trim(source[ci+1:], " \n")),
            },
          }
        }
      }

      if ci > 2 {
        switch source[ci - 2:ci] { // Three-character compound assignment operators
        case ">>", "<<":
          return LineType{
            Name: "CompoundAssignment",
            Words: map[string]interface{}{
              "Operator": string(source[ci - 2:ci]),
              "VariableName": strings.Trim(source[:ci - 2], " \n"), // Just strip two characters earlier
              "Expression": ParseExpression(strings.Trim(source[ci+1:], " \n")),
            },
          }
        }
      }

      if ci > 3 {
        if source[ci - 3:ci] == ">>>" { // Compound right-shift filled 0 assignment operator
          return LineType{
            Name: "CompoundAssignment",
            Words: map[string]interface{}{
              "Operator": ">>>",
              "VariableName": strings.Trim(source[:ci - 3], " \n"), // Just strip three characters earlier
              "Expression": ParseExpression(strings.Trim(source[ci+1:], " \n")),
            },
          }
        }
      }

      // If there is no assignment operator, then it is an equals
      switch len(currentWords) {
      case 1: // One word means assigning a variable
      return LineType{
        Name: "AssignVariable",
        Words: map[string]interface{}{
          "VariableName": ParseExpression(strings.Trim(source[:ci], " \n")),
          "Expression": ParseExpression(strings.Trim(source[ci + 1:], " \n")),
        },
      }
      case 2: // Assign and create a variable
        spaceIndex := parsetools.FindNextIndexOfCharWithSkip(source, ' ', "<") // Skips generic types
        currentReturn = strings.Trim(source[:spaceIndex], " \n") // Global variable
        return LineType{
          Name: "CreateAndAssignVariable",
          Words: map[string]interface{}{
            "VariableName": ParseExpression(strings.Trim(source[spaceIndex + 1:ci], " \n")),
            "VariableType": currentReturn,
            "Expression": ParseExpression(strings.Trim(source[ci + 1:], " \n")),
          },
        }
      }
    case ':': // Switch case or loop label
      // Test if this element is not at the end of the source
      if ci + 1 < len(source) {
        if source[ci + 1] == ':' { // Double colon, method reference operator
          // Method stops at non-letter or number
          nextInvalid := parsetools.IndexOfNextNonNormal(source[ci + 2:]) + ci + 2
          currentWords = append(currentWords, LineType{
						Name: "MethodReference",
						Words: map[string]interface{}{
							"ClassName": source[lastLine:ci],
							"ReferenceName": source[ci + 2:nextInvalid],
						},
					})
        }
      }

    }
  }

  panic("Unknown line")
}

/* func OldParseLine(source string) LineType {

  fmt.Printf("Line content: %s\n", source)

  // Test for lines with equals, to catch a variable being assigned
  equalsIndex := parsetools.FindNextIndexOfCharWithSkip(source, '=', `'"{(`)

  // Filters out compound assignment operator (*=, +=) and such
  if equalsIndex > 1 {
    switch source[equalsIndex - 1] {
    // Not equals operator
    case '!':
      return LineType{
        Name: "NotEquals",
        Words: map[string]interface{}{
          "LeftExpression": "",
          "RightExpression": "",
        },
      }
    // The character before the equals, detects things like "+=" and "*=", the compound assignment operators
		// https://www.geeksforgeeks.org/compound-assignment-operators-java/
    case '^', '+', '-', '*', '/', '%', '&', '|':
      return LineType{
        Name: "CompoundAssignment",
        Words: map[string]interface{}{
          "Operator": string(source[equalsIndex - 1]),
          "VariableName": strings.Trim(source[:equalsIndex - 1], " \n"), // Just strip one character earlier
          "Expression": ParseExpression(strings.Trim(source[equalsIndex+1:], " \n")),
        },
      }
    }
  }

  if equalsIndex > 2 {
    switch source[equalsIndex - 2:equalsIndex] { // Three-character compound assignment operators
    case ">>", "<<":
      return LineType{
        Name: "CompoundAssignment",
        Words: map[string]interface{}{
          "Operator": string(source[equalsIndex - 2:equalsIndex]),
          "VariableName": strings.Trim(source[:equalsIndex - 2], " \n"), // Just strip two characters earlier
          "Expression": ParseExpression(strings.Trim(source[equalsIndex+1:], " \n")),
        },
      }
    }
  }

  if equalsIndex > 3 {
    if source[equalsIndex - 3:equalsIndex] == ">>>" { // Compound right-shift filled 0 assignment operator
      return LineType{
        Name: "CompoundAssignment",
        Words: map[string]interface{}{
          "Operator": ">>>",
          "VariableName": strings.Trim(source[:equalsIndex - 3], " \n"), // Just strip three characters earlier
          "Expression": ParseExpression(strings.Trim(source[equalsIndex+1:], " \n")),
        },
      }
    }
  }

	if equalsIndex != -1 {
		// Counts the space between the type of variable and variable name (ex: int value)
		switch parsetools.CountRuneWithSkip(strings.Trim(source[:equalsIndex], " \n"), ' ', "<") {
		case 0: // 0 spaces means that there are no spaces, and the variable is just being re-assigned
			return LineType{
				Name: "AssignVariable",
				Words: map[string]interface{}{
					"VariableName": ParseExpression(strings.Trim(source[:equalsIndex], " \n")),
					"Expression": []LineType{ParseLine(strings.Trim(source[equalsIndex + 1:], " \n"))},
				},
			}
		case 1: // 1 space means that the variable is being declared and set to a value
			spaceIndex := parsetools.FindNextIndexOfCharWithSkip(source, ' ', "<") // Skips generic types
			currentReturn = strings.Trim(source[:spaceIndex], " \n") // Global variable
			return LineType{
				Name: "CreateAndAssignVariable",
				Words: map[string]interface{}{
					"VariableName": ParseExpression(strings.Trim(source[spaceIndex + 1:equalsIndex], " \n")),
					"VariableType": currentReturn,
					"Expression": ParseExpression(strings.Trim(source[equalsIndex + 1:], " \n")),
				},
			}
		}
	}
	words := parsetools.DiscardBlankStrings(strings.Split(source, " "))
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
				"Expression": ParseExpression(source[len(words[0]) + 1:]),
			},
		}
	case "new":
		return LineType{
			Name: "NewConstructor",
			Words: map[string]interface{}{
				"Expression": ParseExpression(source[len(words[0]) + 1:])[0],
			},
		}
	case "throw":
		return LineType{
			Name: "ThrowException",
			Words: map[string]interface{}{
				"Expression": ParseLine(source[len(words[0]) + 1:]), // Re-parse the line
			},
		}
	}

	return LineType{
		Name: "GenericLine",
		Words: map[string]interface{}{
			"Statement": ParseExpression(source),
		},
	}
}
*/

// Really simply parses a variable of the form (Type Variable), and handles everything else
func ParseVariableAndType(source string) (variableType, variableName string) {
	spaceIndex := parsetools.FindNextIndexOfCharWithSkip(source, ' ', "<") // The space between the type and the variable, accounts for generics
	return strings.Trim(source[:spaceIndex], " \n"), strings.Trim(source[spaceIndex + 1:], " \n")
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

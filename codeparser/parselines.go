package codeparser

import (
  "strings"

  "gitlab.nicholasnovak.io/snapdragon/java2go/parsetools"
)

func ParseLine(source string) LineTyper {
  source = strings.Trim(source, " \n")
  // fmt.Printf("Line: %s\n", source)
  currentWords := []LineType{}

  lastLine, ci := 0, 0
  for ; ci < len(source); ci++ {
    switch rune(source[ci]) {
    case '<': // Skip over angles
      // But not if it is a part of a bit-shift
      switch rune(source[ci + 1]) {
      case '<', '=':
      default:
        closingAngle := parsetools.IndexOfMatchingAngles(source, ci)
        ci = closingAngle
        lastLine = ci
      }
    case '[': // Skip over brackets
      closingBrace := parsetools.IndexOfMatchingBrackets(source, ci)
      ci = closingBrace
      lastLine = ci
    case ' ': // A word
      switch source[lastLine:ci] {
      case "": // Empty string, ignore
      case "return": // The return keyword for a line
    		return LineType{
    			Name: "ReturnStatement",
    			Words: map[string]interface{}{
    				"Expression": ParseExpression(strings.Trim(source[ci:], " ")),
    			},
    		}
    	case "throw":
    		return LineType{
    			Name: "ThrowException",
    			Words: map[string]interface{}{
    				"Expression": ParseExpression(strings.Trim(source[ci:], " ")),
    			},
    		}
      default:
        curWord := ParseExpression(strings.Trim(source[lastLine:ci], " "))
        if len(curWord) > 0 {
          currentWords = append(currentWords, curWord[0])
          lastLine = ci + 1
        }
      }
    case '(':
      closingParenths := parsetools.IndexOfMatchingParenths(source, ci)
      // Handle the first word before the parenthesies without a space (for() vs for ())
      if source[lastLine:ci] != "" {
        currentWords = append(currentWords, ParseExpression(source[lastLine:ci])[0])
      }

      // Assuming that there is some type of control flow

      startingBrace := strings.IndexRune(source[closingParenths:], '{') + closingParenths
      closingBrace := -1
      if startingBrace - closingParenths != -1 {
        closingBrace = parsetools.IndexOfMatchingBrace(source, startingBrace)
      }

      if len(currentWords) > 0 && currentWords[len(currentWords) - 1].Name == "LocalVariableOrExpression" {
        switch currentWords[len(currentWords) - 1].Words["Expression"].(string) {
        case "for":
          return ParseControlFlow("for", source[ci + 1:closingParenths], source[startingBrace + 1:closingBrace])
        case "if":
          // Else if
          if len(currentWords) > 1 && currentWords[len(currentWords) - 2].Words["Expression"].(string) == "else" {
            // Inline
            if closingBrace == -1 {
              return ParseControlFlow("else if", source[ci + 1:closingParenths], source[startingBrace + 1:])
            }
            return ParseControlFlow("else if", source[ci + 1:closingParenths], source[startingBrace + 1:closingBrace])
          }
          // Inline
          if closingBrace == -1 {
            return ParseControlFlow("if", source[ci + 1:closingParenths], strings.Trim(source[closingParenths + 1:], " "))
          }
          return ParseControlFlow("if", source[ci + 1:closingParenths], source[startingBrace + 1:closingBrace])
        case "while":
          // If the while is part of a do-while loop, then there won't be a closing brace
          if closingBrace == -1 {
            return ParseControlFlow("doWhileCondition", source[ci + 1:closingParenths], "")
          }
          return ParseControlFlow("while", source[ci + 1:closingParenths], source[startingBrace + 1:closingBrace])
        case "catch":
          return ParseControlFlow("catch", source[ci + 1:closingParenths], "")
        case "switch":
          return ParseControlFlow("switch", source[ci + 1:closingParenths], source[startingBrace + 1:closingBrace])
        case "synchronized":
          return ParseControlFlow("synchronized", source[ci + 1:closingParenths], source[startingBrace + 1:closingBrace])
        default:
          // fmt.Printf("--- Control flow: [%s] ---\n", currentWords[len(currentWords) - 1].Words["Expression"].(string))
        }
      }

      // No control flow detected, just a normal method that is parsed as an expression
      ci = closingParenths
      lastLine = ci + 1
    case '{':
      closingBrace := parsetools.IndexOfMatchingBrace(source, ci)
      // Handle the first word before the parenthesies without a space (for() vs for ())
      if source[lastLine:ci] != "" {
        currentWords = append(currentWords, ParseExpression(source[lastLine:ci])...)
      }

      // Assuming that there is some type of control flow
      if currentWords[len(currentWords) - 1].Name == "LocalVariableOrExpression" {
        curWord := currentWords[len(currentWords) - 1].Words["Expression"].(string)

        // Detect a labeled block
        if colonInd := strings.IndexRune(curWord, ':'); colonInd != -1 {
          // Colon is at end of string, label is the only word before the block
          if curWord[len(curWord) - 1] == ':' {
            return ParseControlFlow("labeledBlock", currentWords[len(currentWords) - 1].Words["Expression"].(string), source[ci + 1:closingBrace])
          }
          // If the labelled block has another word on it

          // Remove the old word
          currentWords = currentWords[:len(currentWords) - 1]
          // Add both words separately to the words
          currentWords = append(currentWords, ParseExpression(curWord[:colonInd + 1])...)
          currentWords = append(currentWords, ParseExpression(curWord[colonInd + 1:])...)
          curWord = currentWords[len(currentWords) - 1].Words["Expression"].(string)
        }

        switch currentWords[len(currentWords) - 1].Words["Expression"].(string) {
        case "else":
          return ParseControlFlow("else", "", source[ci + 1:closingBrace])
        case "try":
          return ParseControlFlow("try", "", source[ci + 1:closingBrace])
        case "finally":
          return ParseControlFlow("finally", "", source[ci + 1:closingBrace])
        case "do":
          return ParseControlFlow("do-while", "", source[ci + 1:closingBrace])
        case "":
          return ParseControlFlow("blankStatement", "", source[ci + 1:closingBrace])
        default:

          panic("Possible unknown control flow: [" + currentWords[len(currentWords) - 1].Words["Expression"].(string) + "]")
        }
      }
    case '=': // Any equals operator or a plain equals
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
      case '0': // Skip if zero words
      case 1: // One word means assigning a variable
      return LineType{
        Name: "AssignVariable",
        Words: map[string]interface{}{
          "VariableName": ParseExpression(strings.Trim(source[:ci], " \n")),
          "Expression": ParseExpression(strings.Trim(source[ci + 1:], " \n")),
        },
      }
      default: // Assign and create a variable
        spaceIndex := parsetools.FindNextIndexOfCharWithSkip(source, ' ', "<") // Skips generic types
        currentReturn = strings.Trim(source[:spaceIndex], " \n") // Global variable
        return LineType{
          Name: "CreateAndAssignVariable",
          Words: map[string]interface{}{
            "VariableName": currentWords[len(currentWords) - 1],
            "VariableType": currentReturn,
            "Expression": ParseExpression(strings.Trim(source[ci + 1:], " \n")),
          },
        }
      }
    }
  }

  // If no other line type is specified, then the line is just a plain expression
  return LineType{
    Name: "GenericExpression",
    Words: map[string]interface{}{
      "Expression": ParseExpression(source),
    },
  }
}

func ParseSwitch(source string) (cases []LineTyper) {
  cases = []LineTyper{}

  lastCase, ci := 0, 0
  for ; ci < len(source); ci++ {
    switch rune(source[ci]) {
      case '{':
        closingBrace := parsetools.IndexOfMatchingBrace(source, ci)
        ci = closingBrace
      case '(':
        closingParenths := parsetools.IndexOfMatchingParenths(source, ci)
        ci = closingParenths
      case '\'':
        closingSingQuotes := parsetools.FindNextIndexOfCharWithSkip(source[ci + 1:], '\'', ``) + ci + 1
        ci = closingSingQuotes
      case '"':
        closingQuotes := parsetools.FindNextIndexOfCharWithSkip(source[ci + 1:], '"', ``) + ci + 1
        ci = closingQuotes
      case ':': // Switch case
        if ci > 0 && source[ci - 1] == ' ' {
          continue
        }

        if ci < len(source) - 1 && source[ci + 1] == ':' {
          ci += 1
          continue
        }

      nextCase := parsetools.IndexWithSkip(source[ci:], "case", `'"{(`) + ci
      nextDefaultCase := parsetools.IndexWithSkip(source[ci:], "default", `'"{(`) + ci

      var caseName string
      if parsetools.ContainsWithSkip(source[lastCase:ci], "case", `'"{(`) {
        caseName = "case"
      } else if parsetools.ContainsWithSkip(source[lastCase:ci], "default", `'"{(`) {
        caseName = "default"
      }

      // No more cases or defaults
      if nextCase - ci == -1 && nextDefaultCase - ci == -1 {
        cases = append(cases, ParseControlFlow(
          caseName,
          source[lastCase + len(caseName):ci],
          source[ci + 1:],
        ))
        ci = len(source)
      // No default case or nextCase is nearer than the default
      } else if nextDefaultCase - ci == -1 || nextCase < nextDefaultCase { // Case is closer
        if ci >= nextCase { // No content for the current case
          cases = append(cases, ParseControlFlow(
            caseName,
            source[lastCase + len(caseName):ci],
            "",
          ))
        } else {
          cases = append(cases, ParseControlFlow(
            caseName,
            source[lastCase + len(caseName):ci],
            source[ci + 1:nextCase],
          ))
          ci = nextCase
        }
        lastCase = ci
      } else { // Default is closer
        if ci >= nextDefaultCase {
          cases = append(cases, ParseControlFlow(
            caseName,
            source[lastCase + len(caseName):ci],
            "",
          ))
        } else {
          cases = append(cases, ParseControlFlow(
            caseName,
            source[lastCase + len(caseName):ci],
            source[ci + 1:nextDefaultCase],
          ))
          ci = nextDefaultCase
        }
        lastCase = ci
      }
    }
  }

  return cases
}

// Really simply parses a variable of the form (Type Variable), and handles everything else
func ParseVariableAndType(source string) (variableType, variableName string) {
	spaceIndex := parsetools.FindNextIndexOfCharWithSkip(source, ' ', "<") // The space between the type and the variable, accounts for generics
	return strings.Trim(source[:spaceIndex], " \n"), strings.Trim(source[spaceIndex + 1:], " \n")
}

func ParseControlFlow(controlBlockname, parameters, source string) LineBlock {
	switch controlBlockname {
	case "for": // For loop, can be a normal for loop, or a for-each loop (enhanced for loop)
		if colonInd := parsetools.FindNextIndexOfCharWithSkip(parameters, ':', `"'{(`); colonInd != -1 { // An enhanced for loop will have a colon in it
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
			Name: "WhileStatement",
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
			Lines: ParseSwitch(strings.Trim(source, " ")),
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
			Words: make(map[string]interface{}),
      Lines: ParseContent(strings.Trim(source, " \n")),
		}
  case "synchronized":
    return LineBlock{
      Name: "SynchronizedBlock",
      Words: map[string]interface{}{
				"Condition": ParseExpression(parameters),
			},
			Lines: ParseContent(strings.Trim(source, " \n")),
    }
  case "else":
    return LineBlock{
      Name: "ElseBlock",
      Words: make(map[string]interface{}),
			Lines: ParseContent(strings.Trim(source, " \n")),
    }
  case "try":
    return LineBlock{
      Name: "TryBlock",
      Words: make(map[string]interface{}),
			Lines: ParseContent(strings.Trim(source, " \n")),
    }
  case "finally":
    return LineBlock{
      Name: "FinallyBlock",
      Words: make(map[string]interface{}),
      Lines: ParseContent(strings.Trim(source, " \n")),
    }
  case "doWhileCondition":
    return LineBlock{
      Name: "DoWhileCondition",
      Words: map[string]interface{}{
        "Condition": ParseExpression(parameters),
      },
    }
  case "blankStatement":
    return LineBlock{
      Name: "BlankStatement",
      Words: make(map[string]interface{}),
      Lines: ParseContent(strings.Trim(source, " \n")),
    }
  case "labeledBlock":
    return LineBlock{
      Name: "LabeledBlock",
      Words: map[string]interface{}{
        "Label": parameters,
      },
      Lines: ParseContent(strings.Trim(source, " \n")),
    }
	default:
    panic("Unknown control flow [" + controlBlockname + "]")
	}
}

func ContentOfBrackets(source string) []string {
  contents := []string{}

  for ci, char := range source {
    if char == '[' {
      closingBracket := parsetools.IndexOfMatchingBrackets(source, ci)
      contents = append(contents, source[ci + 1:closingBracket])
    }
  }

  return contents
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

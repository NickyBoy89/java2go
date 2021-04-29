package parsing

import (
  "strings"
  // "fmt"
)

func ParseInterface(sourceString string) ParsedInterface {
  sourceString = strings.ReplaceAll(sourceString, "\t", "")
  sourceString = strings.ReplaceAll(sourceString, "\r", "")

  sourceString = strings.Trim(sourceString, "\n")

  var result ParsedInterface

  bodyDivider := strings.IndexRune(sourceString, '{')

  words := DiscardBlankStrings(strings.Split(sourceString[:bodyDivider], " "))

  result.Name = words[len(words) - 1]
  result.Modifiers = words[:len(words) - 2]
  result.Methods = []ParsedMethod{}
  result.DefaultMethods = []ParsedMethod{}
  result.StaticFields = []ParsedVariable{}

  classBody := sourceString[bodyDivider + 1:IndexOfMatchingBrace(sourceString, bodyDivider)]

  var currentAnnotation string

  lastInterest := 0
  ci := 0
  for ; ci < len(classBody); ci++ {
    char := classBody[ci]
    if char == '@' { // Detected an annotation
      newlineIndex, err := FindNextIndexOfChar(classBody[ci:], '\n')
      if err != nil {
        panic(err)
      }

      if currentAnnotation == "" { // No annotation already there
        currentAnnotation = classBody[ci:newlineIndex + ci]
      } else {
        currentAnnotation += "\n" + classBody[ci:newlineIndex + ci]
      }
      ci += newlineIndex + 1
      lastInterest = ci
    } else if char == ';' || char == '=' { // Semicolon and equal detect class variables
      semicolonIndex := FindNextSemicolonIndex(classBody[ci:]) + ci
      result.StaticFields = append(result.StaticFields, ParseClassVariable(strings.Trim(classBody[lastInterest + 1:semicolonIndex], " \n"), currentAnnotation))
      ci = semicolonIndex
      currentAnnotation = ""
      lastInterest = ci + 1
    } else if char == '{' {
      closingBrace := IndexOfMatchingBrace(classBody, ci)
      if strings.Contains(classBody[lastInterest:ci], "class") { // Nested class
        result.NestedClasses = append(result.NestedClasses, ParseClass(strings.Trim(classBody[lastInterest + 1:IndexOfMatchingBrace(classBody, ci) + 1], " \n")))
      } else if strings.Contains(classBody[lastInterest:ci], "interface") { // Nested interface
        result.NestedClasses = append(result.NestedClasses, ParseInterface(strings.Trim(classBody[lastInterest + 1:IndexOfMatchingBrace(classBody, ci) + 1], " \n")))
      } else if strings.Contains(classBody[lastInterest:ci], "enum") { // Nested enum
        result.NestedClasses = append(result.NestedClasses, ParseEnum(strings.Trim(classBody[lastInterest + 1:IndexOfMatchingBrace(classBody, ci) + 1], " \n")))
      }
      ci = closingBrace
      lastInterest = ci + 1
    } else if char == '(' {
      methodWords := DiscardBlankStrings(strings.Split(classBody[lastInterest:ci], " "))
      switch methodWords[0] {
      case "default":
        openingBracket := strings.IndexRune(classBody[lastInterest:], '{') + lastInterest
        closingBracket := IndexOfMatchingBrace(classBody, openingBracket)
        result.Methods = append(result.Methods, ParseMethod(strings.Trim(classBody[lastInterest + 1:closingBracket + 1], " \n"), currentAnnotation))
        ci = closingBracket + 1
      case "static":
        openingBracket := strings.IndexRune(classBody[lastInterest:], '{') + lastInterest
        closingBracket := IndexOfMatchingBrace(classBody, openingBracket)
        result.Methods = append(result.Methods, ParseMethod(strings.Trim(classBody[lastInterest + 1:closingBracket + 1], " \n"), currentAnnotation))
        ci = closingBracket + 1
      default:
        closingSemicolon := FindNextSemicolonIndex(classBody[ci:]) + ci
        result.Methods = append(result.Methods, ParseMethod(strings.Trim(classBody[lastInterest + 1:closingSemicolon], " \n"), currentAnnotation))
        ci = closingSemicolon + 1
      }
      lastInterest = ci
      currentAnnotation = ""
    }
  }

  return result
}

package parsing

import (
  "strings"
  // "fmt"
)

func ParseEnum(sourceString string) ParsedEnum {
  // panic("Enum not implemented")
  sourceString = RemoveComments(sourceString)
  sourceString = RemoveImports(sourceString)
  sourceString = RemovePackage(sourceString)

  sourceString = strings.ReplaceAll(sourceString, "\t", "")
  sourceString = strings.ReplaceAll(sourceString, "\r", "")

  sourceString = strings.Trim(sourceString, "\n")

  var result ParsedEnum

  bodyDivider := strings.IndexRune(sourceString, '{')

  words := discardBlankStrings(strings.Split(sourceString[:bodyDivider], " "))

  result.Implements = []string{}

  classWordRange := len(words)
  for wi, testWord := range words {
    if testWord == "extends" { // Extends comes before implements
      panic("Enum already extends by default Enum, so it should not extend anything else")
    } else if testWord == "implements" {
      result.Implements = append(result.Implements, TrimAll(words[wi + 1:], ",")...)
      if classWordRange >= len(words) { // No extends already cut out of the string
        classWordRange = wi
      }
    }
  }

  words = words[:classWordRange]

  result.Name = words[len(words) - 1]
  result.Modifiers = words[:len(words) - 2]
  result.ClassVariables = []ParsedVariable{}
  result.NestedClasses = []ParsedClasses{}

  classBody := sourceString[bodyDivider + 1:IndexOfMatchingBrace(sourceString, bodyDivider)]

  // Start parsing the enum constants
  // Assumes that all the enum constants are at the top of the declaration
  // (No methods, etc...) before
  enumEnd := FindNextSemicolonIndex(classBody)

  enumBody := strings.Trim(classBody[:enumEnd], " \n")
  classBody = classBody[enumEnd + 1:]

  enumFields := TrimAll(strings.Split(enumBody, ","), " \n")

  for _, field := range enumFields {
    if fieldStart := strings.IndexRune(field, '('); fieldStart != -1 {
      result.EnumFields = append(result.EnumFields, EnumField{Name: field[:fieldStart], Parameters: ParseParameters(field[fieldStart + 1:len(field) - 1])})
    } else {
      result.EnumFields = append(result.EnumFields, EnumField{Name: field, Parameters: []ParsedVariable{}})
    }
  }

  var currentAnnotation string

  lastInterest := 0
  ci := 0
  for ; ci < len(classBody); ci++ {
    char := classBody[ci]
    if char == '@' { // Detected an annotation
      var newlineIndex, spaceIndex int // Also assumes that there are at least one of the characters in the file
      if strings.ContainsRune(classBody[ci:], '\n') {
        newlineIndex = strings.IndexRune(classBody[ci:], '\n') + ci
      }
      if strings.ContainsRune(classBody[ci:], ' ') {
        spaceIndex = strings.IndexRune(classBody[ci:], ' ') + ci
      }
      if spaceIndex == 0 || newlineIndex < spaceIndex {
        if currentAnnotation != "" { // Stacked annotaions
          currentAnnotation += "\n" + classBody[ci:newlineIndex]
        } else {
          currentAnnotation = classBody[ci:newlineIndex]
        }
        ci = newlineIndex
        lastInterest = ci
      } else {
        if currentAnnotation != "" {
          currentAnnotation += "\n" + classBody[ci:spaceIndex]
        } else {
          currentAnnotation = classBody[ci:spaceIndex]
        }
        ci = spaceIndex
        lastInterest = ci
      }
    } else if char == ';' || char == '=' { // Semicolon and equal detect class variables
      semicolonIndex := FindNextSemicolonIndex(classBody[ci:]) + ci
      result.ClassVariables = append(result.ClassVariables, ParseClassVariable(strings.Trim(classBody[lastInterest + 1:semicolonIndex], " \n"), currentAnnotation))
      ci = semicolonIndex
      currentAnnotation = ""
      lastInterest = ci
    } else if char == '{' {
      if strings.Trim(classBody[lastInterest:ci], " \n") == "static" { // Handle static block
        result.StaticBlocks = append(result.StaticBlocks, strings.Trim(classBody[strings.IndexRune(classBody[lastInterest:], '{') + lastInterest + 1:IndexOfMatchingBrace(classBody, ci)], " \n"))
        ci = IndexOfMatchingBrace(classBody, ci) + 1// Cut out the remaining brace
        lastInterest = ci
      } else if strings.Contains(classBody[lastInterest:ci], "class") { // Nested class
        result.NestedClasses = append(result.NestedClasses, ParseClass(strings.Trim(classBody[lastInterest + 1:IndexOfMatchingBrace(classBody, ci) + 1], " \n")))
        ci = IndexOfMatchingBrace(classBody, ci)
        lastInterest = ci
      } else if strings.Contains(classBody[lastInterest:ci], "interface") { // Nested interface
        result.NestedClasses = append(result.NestedClasses, ParseInterface(strings.Trim(classBody[lastInterest + 1:IndexOfMatchingBrace(classBody, ci) + 1], " \n")))
        ci = IndexOfMatchingBrace(classBody, ci)
        lastInterest = ci
      } else if strings.Contains(classBody[lastInterest:ci], "enum") { // Nested enum
        result.NestedClasses = append(result.NestedClasses, ParseEnum(strings.Trim(classBody[lastInterest + 1:IndexOfMatchingBrace(classBody, ci) + 1], " \n")))
        ci = IndexOfMatchingBrace(classBody, ci)
        lastInterest = ci
      }
    } else if char == '(' {
      startingBraceIndex := strings.IndexRune(classBody[ci:], '{') + ci
      result.Methods = append(result.Methods, ParseMethod(strings.Trim(classBody[lastInterest + 1:IndexOfMatchingBrace(classBody, startingBraceIndex)], " \n"), currentAnnotation))
      currentAnnotation = ""
      ci = IndexOfMatchingBrace(classBody, startingBraceIndex) + 1
      lastInterest = ci
    }
  }

  return result
}

package parsing

import (
  "strings"
  // "fmt"
)

func ParseClass(sourceString string) ParsedClass {
  sourceString = RemoveComments(sourceString)
  sourceString = RemoveImports(sourceString)
  sourceString = RemovePackage(sourceString)

  sourceString = strings.ReplaceAll(sourceString, "\t", "")
  sourceString = strings.ReplaceAll(sourceString, "\r", "")

  sourceString = strings.Trim(sourceString, "\n")

  var result ParsedClass

  bodyDivider := strings.IndexRune(sourceString, '{')

  words := discardBlankStrings(strings.Split(sourceString[:bodyDivider], " "))

  result.Implements = []string{}

  classWordRange := len(words)
  for wi, testWord := range words {
    if testWord == "extends" { // Extends comes before implements
      result.Extends = strings.Trim(words[wi + 1], ",")
      classWordRange = wi
    } else if testWord == "implements" {
      result.Implements = append(result.Implements, TrimAll(words[wi + 1:], ",")...)
      if classWordRange != len(words) && wi < classWordRange {
        classWordRange = wi
      }
    }
  }

  words = words[:classWordRange]

  result.Name = words[len(words) - 1]
  result.Modifiers = words[:len(words) - 2]
  result.ClassVariables = []ParsedVariable{}
  result.NestedClasses = []ParsedClasses{}
  result.StaticBlocks = []string{}

  classBody := sourceString[bodyDivider + 1:IndexOfMatchingBrace(sourceString, bodyDivider)]

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
      } else {
        if currentAnnotation != "" {
          currentAnnotation += "\n" + classBody[ci:spaceIndex]
        } else {
          currentAnnotation = classBody[ci:spaceIndex]
        }
        ci = spaceIndex
      }
      lastInterest = ci
    } else if char == ';' || char == '=' { // Semicolon and equal detect class variables
      semicolonIndex := FindNextSemicolonIndex(classBody[ci:]) + ci
      result.ClassVariables = append(result.ClassVariables, ParseClassVariable(strings.Trim(classBody[lastInterest + 1:semicolonIndex], " \n"), currentAnnotation))
      ci = semicolonIndex
      currentAnnotation = ""
      lastInterest = ci + 1
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
      closingParenthsIndex := strings.IndexRune(classBody[ci:], ')') + ci
      startingBraceIndex := strings.IndexRune(classBody[closingParenthsIndex:], '{') + closingParenthsIndex
      if rune(classBody[startingBraceIndex]) == '{' {
        result.Methods = append(result.Methods, ParseMethod(strings.Trim(classBody[lastInterest + 1:IndexOfMatchingBrace(classBody, startingBraceIndex)], " \n"), currentAnnotation))
        ci = IndexOfMatchingBrace(classBody, startingBraceIndex) + 1
      } else {
        result.Methods = append(result.Methods, ParseMethod(strings.Trim(classBody[lastInterest + 1:closingParenthsIndex + 1], " \n"), currentAnnotation))
        ci = closingParenthsIndex + 1 // Removes the semicolon
      }
      currentAnnotation = ""
      lastInterest = ci
    }
  }

  return result
}

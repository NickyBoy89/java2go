package parsing

import (
  "strings"
  "fmt"
)

func ParseInterface(sourceString string) ParsedInterface {
  sourceString = RemoveComments(sourceString)
  sourceString = RemoveImports(sourceString)
  sourceString = RemovePackage(sourceString)

  sourceString = strings.ReplaceAll(sourceString, "\t", "")
  sourceString = strings.ReplaceAll(sourceString, "\r", "")

  sourceString = strings.Trim(sourceString, "\n")

  var result ParsedInterface

  bodyDivider := strings.IndexRune(sourceString, '{')

  words := discardBlankStrings(strings.Split(sourceString[:bodyDivider], " "))

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
      fmt.Println("Annotation")
      var newlineIndex, spaceIndex int
      if strings.ContainsRune(classBody[ci:], '\n') {
        newlineIndex = strings.IndexRune(classBody[ci:], '\n') + ci
      }
      if strings.ContainsRune(classBody[ci:], ' ') {
        spaceIndex = strings.IndexRune(classBody[ci:], ' ') + ci
      }
      if newlineIndex < spaceIndex && newlineIndex != 0 {
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
      fmt.Println("Found semicolon or equals")
      semicolonIndex := FindNextSemicolonIndex(classBody[ci:]) + ci
      result.StaticFields = append(result.StaticFields, ParseClassVariable(strings.Trim(classBody[lastInterest + 1:semicolonIndex], " \n"), currentAnnotation))
      ci = semicolonIndex
      currentAnnotation = ""
      lastInterest = ci
    } else if char == '{' {
      fmt.Println("Found brace")
      if strings.Contains(classBody[lastInterest:ci], "class") { // Nested class
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
      fmt.Println("Found parenths")
      firstWordIndex := IndexOfNextNonBlankChar(classBody[lastInterest:]) + lastInterest
      if classBody[firstWordIndex:strings.IndexRune(classBody[firstWordIndex:], ' ') + firstWordIndex] == "default" {
        fmt.Println("Found block-level default method in interface")
        openingBracket := strings.IndexRune(classBody[lastInterest:], '{') + lastInterest
        closingBracket := IndexOfMatchingBrace(classBody, openingBracket)
        result.Methods = append(result.Methods, ParseMethod(strings.Trim(classBody[lastInterest + 1:closingBracket + 1], " \n"), currentAnnotation))
        ci = closingBracket + 1
      } else if classBody[firstWordIndex:strings.IndexRune(classBody[firstWordIndex:], ' ') + firstWordIndex] == "static" {
        openingBracket := strings.IndexRune(classBody[lastInterest:], '{') + lastInterest
        closingBracket := IndexOfMatchingBrace(classBody, openingBracket)
        result.Methods = append(result.Methods, ParseMethod(strings.Trim(classBody[lastInterest + 1:closingBracket + 1], " \n"), currentAnnotation))
        ci = closingBracket + 1
      } else {
        fmt.Println("Found inline block-level default method in interface")
        closingSemicolon := FindNextSemicolonIndex(classBody[ci:]) + ci
        result.Methods = append(result.Methods, ParseMethod(strings.Trim(classBody[lastInterest + 1:closingSemicolon], " \n"), currentAnnotation))
        ci = closingSemicolon + 1
      }
      fmt.Printf("Found default method in interface: %v\n", classBody[lastInterest:ci + 1])
      lastInterest = ci
      currentAnnotation = ""
    }
  }

  return result
}

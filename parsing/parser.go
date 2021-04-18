package parsing

import (
  "strings"
  // "fmt"

  "gitlab.nicholasnovak.io/snapdragon/java2go/keywords"
)

func ParseFile(sourceString string) ParsedClasses {
  modifierWords := strings.Split(sourceString[:strings.IndexRune(sourceString, '{')], " ")
  if Contains("class", modifierWords) {
    return ParseClass(sourceString)
  } else if Contains("interface", modifierWords) || Contains("@interface", modifierWords) {
    return ParseInterface(sourceString)
  } else if Contains("enum", modifierWords) {
    return ParseEnum(sourceString)
  } else {
    panic("No valid class in specified file")
  }
}

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

  result.Name = words[len(words) - 1]
  result.Modifiers = words[:len(words) - 2]
  result.ClassVariables = []ParsedVariable{}
  result.NestedClasses = []ParsedClasses{}

  classBody := sourceString[bodyDivider + 1:IndexOfMatchingBrace(sourceString, bodyDivider)]

  // Start parsing the enum constants
  enumEnd := FindNextSemicolonIndex(classBody)
  enumStart := strings.LastIndex(classBody[:enumEnd], "}")

  enumBody := strings.Trim(classBody[enumStart + 1:enumEnd], " \n")
  classBody = classBody[:enumStart + 1] + classBody[enumEnd + 1:]

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
      startingBraceIndex := strings.IndexRune(classBody[ci:], '{') + ci
      result.Methods = append(result.Methods, ParseMethod(strings.Trim(classBody[lastInterest + 1:IndexOfMatchingBrace(classBody, startingBraceIndex)], " \n"), currentAnnotation))
      currentAnnotation = ""
      ci = IndexOfMatchingBrace(classBody, startingBraceIndex) + 1
      lastInterest = ci
    }
  }

  return result
}

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
      if !strings.ContainsRune(classBody[ci:strings.IndexRune(classBody[ci:], ';') + ci], '{') {
        closingParenthsIndex := strings.IndexRune(classBody[ci:], ')') + ci
        result.Methods = append(result.Methods, ParseMethod(strings.Trim(classBody[lastInterest + 1:closingParenthsIndex + 1], " \n"), currentAnnotation))
        ci = closingParenthsIndex + 1 // Removes the semicolon
      } else {
        startingBraceIndex := strings.IndexRune(classBody[ci:], '{') + ci
        result.Methods = append(result.Methods, ParseMethod(strings.Trim(classBody[lastInterest + 1:IndexOfMatchingBrace(classBody, startingBraceIndex)], " \n"), currentAnnotation))
        ci = IndexOfMatchingBrace(classBody, startingBraceIndex) + 1
      }
      currentAnnotation = ""
      lastInterest = ci
    }
  }

  return result
}

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

  classBody := sourceString[bodyDivider + 1:IndexOfMatchingBrace(sourceString, bodyDivider)]

  var currentAnnotation string

  lastInterest := 0
  ci := 0
  for ; ci < len(classBody); ci++ {
    char := classBody[ci]
    if char == '@' { // Detected an annotation
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
    } else if char == '{' {
      if strings.Contains(classBody[lastInterest:ci], "class") {
        result.NestedClasses = append(result.NestedClasses, ParseClass(strings.Trim(classBody[lastInterest + 1:IndexOfMatchingBrace(classBody, ci) + 1], " \n")))
        ci = IndexOfMatchingBrace(classBody, ci)
        lastInterest = ci
      }
    } else if char == '(' {
      closingParenthsIndex := strings.IndexRune(classBody[ci:], ')') + ci
      if FindNextNonBlankChar(classBody[closingParenthsIndex + 1:]) == ';' {
        result.Methods = append(result.Methods, ParseMethod(strings.Trim(classBody[lastInterest + 1:closingParenthsIndex + 1], " \n"), currentAnnotation))
        currentAnnotation = ""
        ci = closingParenthsIndex + 1 // Removes the semicolon
      } else {
        startingBraceIndex := strings.IndexRune(classBody[ci:], '{') + ci
        if startingBraceIndex - ci == -1 {
          result.Methods = append(result.Methods, ParseMethod(strings.Trim(classBody[lastInterest + 1:strings.IndexRune(classBody[lastInterest + 1:], ';') + lastInterest + 1], " \n"), currentAnnotation))
          ci = strings.IndexRune(classBody[lastInterest + 1:], ';') + lastInterest + 1
        } else {
          result.Methods = append(result.Methods, ParseMethod(strings.Trim(classBody[lastInterest + 1:IndexOfMatchingBrace(classBody, startingBraceIndex)], " \n"), currentAnnotation))
          ci = IndexOfMatchingBrace(classBody, startingBraceIndex)
        }
        currentAnnotation = ""
      }
      lastInterest = ci
    }
  }

  return result
}

func ParseMethod(source, annotation string) ParsedMethod {
  indexOfDelimiter := strings.IndexRune(source, '(')

  words := discardBlankStrings(strings.Split(source[:indexOfDelimiter], " "))

  // Tests for a constructor by the presence of a modifier word where the
  // return type should be, or just the absence of a return type
  // and any other access modifiers
  if len(words) < 2 || Contains(words[len(words) - 2], append(keywords.AccessModifiers, keywords.NonAccessModifiers...)) {
    return ParsedMethod{
      Name: words[len(words) - 1],
      Modifiers: words[:len(words) - 1],
      Annotation: annotation,
      Parameters: ParseParameters(source[indexOfDelimiter + 1:strings.IndexRune(source, ')')]),
      ReturnType: "constructor",
      Body: RemoveIndentation(source[strings.IndexRune(source, '{') + 1:]),
    }
  }

  if !strings.ContainsRune(source, '{') {
    return ParsedMethod{
      Name: words[len(words) - 1],
      Modifiers: words[:len(words) - 2],
      Annotation: annotation,
      Parameters: ParseParameters(source[indexOfDelimiter + 1:strings.IndexRune(source, ')')]),
      ReturnType: words[len(words) - 2],
      Body: "",
    }
  }

  return ParsedMethod{
    Name: words[len(words) - 1],
    Modifiers: words[:len(words) - 2],
    Annotation: annotation,
    Parameters: ParseParameters(source[indexOfDelimiter + 1:strings.IndexRune(source, ')')]),
    ReturnType: words[len(words) - 2],
    Body: RemoveIndentation(source[strings.IndexRune(source, '{') + 1:]),
  }
}

func IndexOfMatchingBrace(searchString string, openingBraceIndex int) int {
  bracketBalance := -1
  if searchString[openingBraceIndex] != '{' {
    panic("Invalid starting brace")
  }
  for ci, char := range searchString[openingBraceIndex + 1:] {
    switch char {
    case '{':
      bracketBalance -= 1
    case '}':
      bracketBalance += 1
    }
    if bracketBalance == 0 {
      return openingBraceIndex + ci + 1 // Account for skipping the first character in the loop
    }
  }
  panic("No matching bracket found, the target code probably has unbalanced brackets")
}

func IndexOfMatchingParenths(searchString string, openingBraceIndex int) int {
  bracketBalance := -1
  if searchString[openingBraceIndex] != '(' {
    panic("Invalid starting parenths")
  }
  for ci, char := range searchString[openingBraceIndex + 1:] {
    switch char {
    case '(':
      bracketBalance -= 1
    case ')':
      bracketBalance += 1
    }
    if bracketBalance == 0 {
      return openingBraceIndex + ci + 1 // Account for skipping the first character in the loop
    }
  }
  panic("No matching parenths found, the target code probably has unbalanced brackets")
}

func ParseClassVariable(source, annotation string) ParsedVariable {
  if strings.ContainsRune(source, '=') {
    sides := discardBlankStrings(TrimAll(strings.Split(source, "="), " \n;"))
    words := discardBlankStrings(strings.Split(sides[0], " "))
    return ParsedVariable{
      Name: words[len(words) - 1],
      DataType: words[len(words) - 2],
      Annotation: annotation,
      Modifiers: words[:len(words) - 2],
      InitialValue: strings.Trim(sides[1], " \n"),
    }
  }


  words := discardBlankStrings(strings.Split(source, " "))
  return ParsedVariable{
    Name: words[len(words) - 1],
    DataType: words[len(words) - 2],
    Annotation: annotation,
    Modifiers: discardBlankStrings(TrimAll(words[:len(words) - 2], " \n;")),
    InitialValue: "",
  }
}

func ParseParameters(source string) []ParsedVariable {
  var parsedParameters []ParsedVariable

  if source == "" {
    return []ParsedVariable{}
  }

  params := strings.Split(source, ",")
  for _, param := range params {
    paramParts := discardBlankStrings(strings.Split(param, " "))
    if paramParts[0][0] == '@' { // First letter of first word is an annotation
      parsedParameters = append(parsedParameters,
        ParsedVariable{
          Name: strings.Trim(paramParts[len(paramParts) - 1], " ,"),
          Modifiers: []string{},
          Annotation: paramParts[0],
          DataType: paramParts[1],
          InitialValue: "",
        },
      )
    } else {
      parsedParameters = append(parsedParameters,
        ParsedVariable{
          Name: strings.Trim(paramParts[len(paramParts) - 1], " ,"),
          Modifiers: []string{},
          DataType: paramParts[0],
          InitialValue: "",
        },
      )
    }
  }

  return parsedParameters
}

func FindNextSemicolonIndex(source string) int {
  ci := 0
  for ; ci < len(source); ci++ {
    char := source[ci]
    if char == '{' {
      ci = IndexOfMatchingBrace(source, ci)
    } else if char == ';' {
      return ci
    }
  }
  panic("No semicolon found")
}

func RemoveIndentation(input string) string {
  var body string

  lines := strings.Split(input, "\n")
  for _, line := range discardBlankStrings(lines) {
    body += strings.Trim(line, " ")
  }

  return body
}

func TrimAll(raw []string, pattern string) []string {
  var trimmed []string

  for _, str := range raw {
    trimmed = append(trimmed, strings.Trim(str, pattern))
  }

  return trimmed
}

func RemoveComments(source string) string {
  modified := source

  for strings.Contains(modified, "/*") {
    openingIndex := strings.Index(modified, "/*")
    closingIndex := strings.Index(modified[openingIndex:], "*/") + openingIndex
    modified = modified[:openingIndex] + modified[closingIndex + 2:]
  }

  for strings.Contains(modified, "//") {
    openingIndex := strings.Index(modified, "//")
    closingIndex := strings.Index(modified[openingIndex:], "\n") + openingIndex
    modified = modified[:openingIndex] + modified[closingIndex + 1:]
  }

  return modified
}

func RemoveImports(source string) string {
  modified := source

  for strings.Contains(modified, "import") {
    openingIndex := strings.Index(modified, "import")
    closingIndex := strings.Index(modified[openingIndex:], "\n") + openingIndex
    modified = modified[:openingIndex] + modified[closingIndex + 1:]
  }

  return modified
}

func RemovePackage(source string) string {
  modified := source

  for strings.Contains(modified, "package") {
    openingIndex := strings.Index(modified, "package")
    closingIndex := strings.Index(modified[openingIndex:], "\n") + openingIndex
    modified = modified[:openingIndex] + modified[closingIndex + 1:]
  }

  return modified
}

func discardBlankStrings(arr []string) []string {
  result := []string{}

  for _, item := range arr {
    if item != "" {
      result = append(result, item)
    }
  }

  return result
}

func FindNextNonBlankChar(source string) rune {
  return rune(source[IndexOfNextNonBlankChar(source)])
}

func IndexOfNextNonBlankChar(source string) int {
  for ci, c := range source {
    if c != ' ' && c != '\n' {
      return ci
    }
  }
  panic("No next blank character found")
}

func Contains(str string, searchFields []string) bool {
  for _, field := range searchFields {
    if field == str {
      return true
    }
  }
  return false
}

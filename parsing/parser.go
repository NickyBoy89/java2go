package parsing

import (
  "strings"
  "fmt"

  "gitlab.nicholasnovak.io/snapdragon/java2go/keywords"
)

func ParseFile(sourceString string) ParsedClasses {
  sourceString = RemoveImports(sourceString)
  sourceString = RemovePackage(sourceString)

  // Splitwords contains all the annotations on the entire class,
  // and the rest of the modifier words as the last index
  // Currently, we are throwing all the annotations away
  splitWords := discardBlankStrings(TrimAll(strings.Split(sourceString[:strings.IndexRune(sourceString, '{')], "\n"), " \n"))

  fmt.Println(splitWords)

  modifierWords := TrimAll(strings.Split(splitWords[len(splitWords) - 1], " "), " \n")

  fmt.Println(modifierWords)

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

func ParseAnnotations(source string) []string {
  var annotations []string

  ci := 0
  for ci, c := range source {
    if c == '\n' {
      annotations = append(annotations, source)
    }
  }

  return annotations
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
    switch char {
    case '{':
      ci = IndexOfMatchingBrace(source, ci)
    case '"':
      ci = strings.IndexRune(source[ci + 1:], '"') + ci + 1
    case '\'':
      ci = strings.IndexRune(source[ci + 1:], '\'') + ci + 1
    case ';':
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

  for strings.Contains(modified, "import ") {
    openingIndex := strings.Index(modified, "import ")
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

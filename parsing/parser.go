package parsing

import (
  "strings"
  "fmt"

  "gitlab.nicholasnovak.io/snapdragon/java2go/keywords"
)

func ParseFile(sourceString string) ParsedClasses {
  sourceString = RemoveImports(sourceString)
  sourceString = RemovePackage(sourceString)
  sourceString = RemoveComments(sourceString)

  _, annotationEnd := ParseAnnotations(sourceString)

  sourceString = sourceString[annotationEnd:]

  modifierWords := TrimAll(strings.Split(sourceString[:strings.IndexRune(sourceString, '{')], " "), " \n")

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

func ParseAnnotations(source string) ([]string, int) {
  var annotations []string

  ci := 0
  for ; ci < len(source); ci++ {
    if source[ci] == '\n' {
      switch source[ci + 1] {
      case ' ', '\n':
        continue
      case '@': // Annotation detected
        newlineIndex, err := FindNextIndexOfChar(source[ci + 1:], '\n')
        if err != nil {
          panic(err)
        }
        annotations = append(annotations, source[ci + 1:newlineIndex + ci])
        ci += newlineIndex
      default:
        if len(annotations) == 0 {
          return []string{}, 0
        }
        return annotations, ci + 1
      }
    }
  }

  return annotations, ci
}

func IndexOfMatchingBrace(searchString string, openingBraceIndex int) int {
  index, err := IndexOfMatchingChar(searchString, openingBraceIndex, '{', '}')
  if err != nil {
    panic(err)
  }
  return index
}

func IndexOfMatchingParenths(searchString string, openingBraceIndex int) int {
  index, err := IndexOfMatchingChar(searchString, openingBraceIndex, '(', ')')
  if err != nil {
    panic(err)
  }
  return index
}

func IndexOfMatchingChar(searchString string, openingIndex int, openingChar, closingChar rune) (int, error) {
  if searchString[openingIndex] != byte(openingChar) {
    return 0, fmt.Errorf("Invalid starting character: %v", searchString[openingIndex])
  }

  bodyString := searchString[openingIndex + 1:]

  if strings.ContainsAny(string(openingChar), `\"'`) {
    panic("Tried to find the matching char of a skipped character")
  }

  balance := -1 // Start with the opening character

  ci := 0
  for ; ci < len(bodyString); ci++ { // Cut out the first character that has already been evaluated
    char := rune(bodyString[ci])
    switch char {
    case '\\':
      ci += 1
    case '"':
      ind, err := FindNextIndexOfCharWithSkip(bodyString[ci + 1:], '"', ``)
      if err != nil {
        panic(err)
      }
      ci += ind + 1
    case '\'':
      ind, err := FindNextIndexOfCharWithSkip(bodyString[ci + 1:], '\'', ``)
      if err != nil {
        panic(err)
      }
      ci += ind + 1
    case openingChar:
      balance -= 1
    case closingChar:
      balance += 1
    }
    if balance == 0 {
      return openingIndex + ci + 1, nil // Account for skipping the first character in the loop
    }
  }
  return 0, fmt.Errorf("No matching bracket found, the target code probably has unbalanced brackets")
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

// Finds the index of the next semicolon in the string, skipping over areas in brackets, or single + double quotes
func FindNextSemicolonIndex(source string) int {
  index, err := FindNextIndexOfChar(source, ';')
  if err != nil {
    panic(err)
  }
  return index
}

func FindNextIndexOfChar(source string, target rune) (int, error) {
  ind, err := FindNextIndexOfCharWithSkip(source, target, `"'({`)
  if err != nil {
    panic(err)
  }
  return ind, nil
}

func FindNextIndexOfCharWithSkip(source string, target rune, skiplist string) (int, error) {
  ci := 0
  for ; ci < len(source); ci++ {
    char := source[ci]
    if char == '\\' { // Skip escaped characters
      ci += 1
    } else if char == '{' && strings.ContainsRune(skiplist, '{') { // Enable skipping brackets only if enabled
      ci = IndexOfMatchingBrace(source, ci)
    } else if char == '(' && strings.ContainsRune(skiplist, '(') {
      ci = IndexOfMatchingParenths(source, ci)
    } else if strings.ContainsAny(string(char), skiplist) { // Just find the matching one for all of these
      ci = strings.IndexRune(source[ci + 1:], rune(char)) + ci + 1
    } else if char == byte(target) {
      return ci, nil
    }
  }
  return -1, fmt.Errorf("Could not find the character: %v", string(target))
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

  var visitedIndexes []int

  for {
    ind, _ := FindNextIndexOfCharWithSkip(modified, '/', `"'`)
    if ContainsInt(ind, visitedIndexes) { // First time it comes back around, it exits
      break
    }
    visitedIndexes = append(visitedIndexes, ind)

    switch modified[ind + 1] {
    case '/': // Inline comment
      visitedIndexes = append(visitedIndexes, ind + 1) // Because this character is a slash also
      closingIndex := strings.IndexRune(modified[ind:], '\n') + ind
      modified = modified[:ind] + modified[closingIndex + 1:]
    case '*': // Block-level comment
      closingIndex := strings.Index(modified[ind + 2:], "*/") + ind + 2
      modified = modified[:ind] + modified[closingIndex + 3:]
    }
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
    if strings.Trim(item, "\n") != "" {
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

func ContainsInt(str int, searchFields []int) bool {
  for _, field := range searchFields {
    if field == str {
      return true
    }
  }
  return false
}

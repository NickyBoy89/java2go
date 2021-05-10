package parsing

import (
  "strings"

  "gitlab.nicholasnovak.io/snapdragon/java2go/keywords"
  "gitlab.nicholasnovak.io/snapdragon/java2go/codeparser"
  "gitlab.nicholasnovak.io/snapdragon/java2go/parsetools"
)

func ParseFile(sourceString string) ParsedClasses {
  sourceString = RemoveImports(sourceString)
  sourceString = RemovePackage(sourceString)
  sourceString = RemoveComments(sourceString)

  _, annotationEnd := ParseAnnotations(sourceString)

  sourceString = sourceString[annotationEnd:]

  modifierWords := parsetools.TrimAll(strings.Split(sourceString[:strings.IndexRune(sourceString, '{')], " "), " \n")

  if parsetools.Contains("class", modifierWords) {
    return ParseClass(sourceString)
  } else if parsetools.Contains("interface", modifierWords) || parsetools.Contains("@interface", modifierWords) {
    return ParseInterface(sourceString)
  } else if parsetools.Contains("enum", modifierWords) {
    return ParseEnum(sourceString)
  } else {
    panic("No valid class in specified file")
  }
}

func ParseMethod(source, annotation string) ParsedMethod {
  indexOfDelimiter := strings.IndexRune(source, '(')

  words := parsetools.DiscardBlankStrings(strings.Split(source[:indexOfDelimiter], " "))

  // Tests for a constructor by the presence of a modifier word where the
  // return type should be, or just the absence of a return type
  // and any other access modifiers
  if len(words) < 2 || parsetools.Contains(words[len(words) - 2], append(keywords.AccessModifiers, keywords.NonAccessModifiers...)) {
    return ParsedMethod{
      Name: words[len(words) - 1],
      Modifiers: words[:len(words) - 1],
      Annotation: annotation,
      Parameters: ParseParameters(source[indexOfDelimiter + 1:strings.IndexRune(source, ')')]),
      ReturnType: "constructor",
      Body: codeparser.ParseContent(parsetools.RemoveIndentation(source[strings.IndexRune(source, '{') + 1:])),
    }
  }

  if !strings.ContainsRune(source, '{') {
    return ParsedMethod{
      Name: words[len(words) - 1],
      Modifiers: words[:len(words) - 2],
      Annotation: annotation,
      Parameters: ParseParameters(source[indexOfDelimiter + 1:strings.IndexRune(source, ')')]),
      ReturnType: words[len(words) - 2],
      Body: []codeparser.LineTyper{},
    }
  }

  return ParsedMethod{
    Name: words[len(words) - 1],
    Modifiers: words[:len(words) - 2],
    Annotation: annotation,
    Parameters: ParseParameters(source[indexOfDelimiter + 1:strings.IndexRune(source, ')')]),
    ReturnType: words[len(words) - 2],
    Body: codeparser.ParseContent(parsetools.RemoveIndentation(source[strings.IndexRune(source, '{') + 1:])),
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
        newlineIndex := parsetools.FindNextIndexOfChar(source[ci + 1:], '\n')

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

func ParseClassVariable(source, annotation string) ParsedVariable {
  if strings.ContainsRune(source, '=') {
    sides := parsetools.DiscardBlankStrings(parsetools.TrimAll(strings.Split(source, "="), " \n;"))
    words := parsetools.DiscardBlankStrings(strings.Split(sides[0], " "))
    return ParsedVariable{
      Name: words[len(words) - 1],
      DataType: words[len(words) - 2],
      Annotation: annotation,
      Modifiers: words[:len(words) - 2],
      InitialValue: strings.Trim(sides[1], " \n"),
    }
  }


  words := parsetools.DiscardBlankStrings(strings.Split(source, " "))
  return ParsedVariable{
    Name: words[len(words) - 1],
    DataType: words[len(words) - 2],
    Annotation: annotation,
    Modifiers: parsetools.DiscardBlankStrings(parsetools.TrimAll(words[:len(words) - 2], " \n;")),
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
    paramParts := parsetools.DiscardBlankStrings(strings.Split(param, " "))
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

func RemoveComments(source string) string {
  modified := source

  var visitedIndexes []int

  for {
    ind := parsetools.FindNextIndexOfCharWithSkip(modified, '/', `"'`)
    if parsetools.ContainsInt(ind, visitedIndexes) { // First time it comes back around, it exits
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

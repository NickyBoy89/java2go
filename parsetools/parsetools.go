package parsetools

import (
  "strings"
  "fmt"
  "unicode"
)

func IndexOfMatchingBrace(searchString string, openingBraceIndex int) int {
  return IndexOfMatchingChar(searchString, openingBraceIndex, '{', '}')
}

func IndexOfMatchingParenths(searchString string, openingBraceIndex int) int {
  return IndexOfMatchingChar(searchString, openingBraceIndex, '(', ')')
}

func IndexOfMatchingAngles(searchString string, openingBraceIndex int) int {
  return IndexOfMatchingChar(searchString, openingBraceIndex, '<', '>')
}

func IndexOfMatchingBrackets(searchString string, openingBraceIndex int) int {
  return IndexOfMatchingChar(searchString, openingBraceIndex, '[', ']')
}

func IndexOfMatchingChar(searchString string, openingIndex int, openingChar, closingChar rune) int {
  if searchString[openingIndex] != byte(openingChar) {
    panic(fmt.Errorf("Invalid starting character: [%s]", string(searchString[openingIndex])))
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
      ind := FindNextIndexOfCharWithSkip(bodyString[ci + 1:], '"', ``)
      ci += ind + 1
    case '\'':
      ind := FindNextIndexOfCharWithSkip(bodyString[ci + 1:], '\'', ``)
      ci += ind + 1
    case openingChar:
      balance -= 1
    case closingChar:
      balance += 1
    }
    if balance == 0 {
      return openingIndex + ci + 1 // Account for skipping the first character in the loop
    }
  }
  panic(fmt.Errorf("No matching [%s] found, the target code probably has unbalanced brackets", string(closingChar)))
}

// Finds the index of the next semicolon in the string, skipping over areas in brackets, or single + double quotes
func FindNextSemicolonIndex(source string) int {
  index := FindNextIndexOfChar(source, ';')
  return index
}

func FindNextIndexOfChar(source string, target rune) int {
  ind := FindNextIndexOfCharWithSkip(source, target, `"'({`)
  if ind == -1 {
    panic("No character [" + string(target) + "] found in string")
  }
  return ind
}

func FindAllIndexesOfChar(source string, target rune) []int {
  return FindAllIndexesOfCharWithSkip(source, target, `"'({`)
}

func FindAllIndexesOfCharWithSkip(source string, target rune, skiplist string) []int {
  indexes := []int{}
  var ind, cutout int
  for {
    ind = FindNextIndexOfCharWithSkip(source[cutout:], target, skiplist)
    if ind != -1 {
      indexes = append(indexes, ind + cutout)
      cutout += ind + 1
    } else {
      break
    }
  }
  return indexes
}

// Finds the index of the specified character, skipping the specified characters
// Returns -1 if no character was found
func FindNextIndexOfCharWithSkip(source string, target rune, skiplist string) int {
  ci := 0
  for ; ci < len(source); ci++ {
    char := source[ci]
    if char == '\\' { // Skip escaped characters
      ci += 1
    } else if char == '{' && strings.ContainsRune(skiplist, '{') { // Enable skipping brackets only if enabled
      ci = IndexOfMatchingBrace(source, ci)
    } else if char == '(' && strings.ContainsRune(skiplist, '(') {
      ci = IndexOfMatchingParenths(source, ci)
    } else if char == '<' && strings.ContainsRune(skiplist, '<') {
      ci = IndexOfMatchingAngles(source, ci)
    } else if char == '[' && strings.ContainsRune(skiplist, '[') {
      ci = IndexOfMatchingBrackets(source, ci)
    } else if strings.ContainsAny(string(char), skiplist) { // Just find the matching one for all of these
      ci = strings.IndexRune(source[ci + 1:], rune(char)) + ci + 1
    } else if char == byte(target) {
      return ci
    }
  }
  return -1
}

func IndexWithSkip(source, target, skiplist string) int {
  // Gets all the indexes of the first character of the index
  for _, charIndex := range FindAllIndexesOfCharWithSkip(source, rune(target[0]), skiplist) {
    // Out-of-bounds
    if charIndex + len(target) > len(source) {
      continue
    }

    for testInd := range target {
      if target[testInd] != source[charIndex + testInd] {
        break
      } else if testInd == len(target) - 1 { // Just tested the last value, and since the invalid check has failed
        return charIndex
      }
    }
  }

  return -1
}

func CountRuneWithSkip(source string, target rune, skiplist string) int {
  var total int
  modified := source

  nextInd := FindNextIndexOfCharWithSkip(modified, target, skiplist)
  if nextInd != -1 {
    modified = modified[nextInd + 1:]
  }
  for ; nextInd != -1; {
    nextInd = FindNextIndexOfCharWithSkip(modified, target, skiplist)
    total += 1
    modified = modified[nextInd + 1:]
  }

  return total
}

func RemoveIndentation(input string) string {
  var body string

  lines := strings.Split(input, "\n")
  for _, line := range DiscardBlankStrings(lines) {
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

func DiscardBlankStrings(arr []string) []string {
  result := []string{}

  for _, item := range arr {
    if strings.Trim(item, "\n") != "" {
      result = append(result, item)
    }
  }

  return result
}

func FindNextNonBlankChar(source string) (rune, int) {
  ind := IndexOfNextNonBlankChar(source)
  return rune(source[ind]), ind
}

func IndexOfNextNonBlankChar(source string) int {
  for ci, c := range source {
    if c != ' ' && c != '\n' {
      return ci
    }
  }
  panic("No next blank character found")
}

// "normal" is defined as not a digit or letter
func IndexOfNextNonNormal(source string) int {
  for ci, c := range source {
    if !unicode.IsLetter(c) || !unicode.IsDigit(c) {
      return ci
    }
  }
  panic("No normal character found")
}

func ContainsWithSkip(source, target, skiplist string) bool {
  if IndexWithSkip(source, target, skiplist) == -1 {
    return false
  }
  return true
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

package bytelevel

import (
  "strings"
  "fmt"
)

func WalkParse(targetString string) *ParsedFile {

  resultFile := new(ParsedFile)

  var words []string

  lastWordIndex := 0

  // Now we are going to go through the file
  for ci, char := range targetString {
    if char == ' ' {
      word := strings.Trim(targetString[lastWordIndex:ci], " \n")
      if word == "" { // Removes blank strings
        continue
      }
      words = append(words, word)
      lastWordIndex = ci
    }
  }

  bracketBalance := 0

  for _, word := range words {
    if Contains(word, AccessModifiers) || word == "static" {
      resultFile.Class.AccessModifiers = append(resultFile.Class.AccessModifiers, word)
    }

    fmt.Println(word)
    fmt.Println("---")
  }

  fmt.Println(resultFile)

  return resultFile
}

func IndexOfMatchingBrace(searchString string, openingBraceIndex int) int {
  bracketBalance := -1
  if searchString[openingBraceIndex] != '{' {
    panic("Invalid starting brace, no match detected")
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

func findNextChar(searchString string, target rune) int {
  result := 0
  for _, c := range searchString {
    if c == target {
      return result
    }
  }
  return -1
}

func Contains(str string, searchFields []string) bool {
  for _, field := range searchFields {
    if field == str {
      return true
    }
  }
  return false
}

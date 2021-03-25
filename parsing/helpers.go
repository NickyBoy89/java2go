package parsing

import (
  "encoding/json"
  "log"
  "strings"
)

func sliceContains(arr []string, target string) bool {
  for _, s := range arr {
    if s == target {
      return true
    }
  }
  return false
}

func prettyPrint(inp string) string {
  pretty, err := json.MarshalIndent(inp, "", "\t")
  if err != nil {
    log.Fatal(err)
  }
  return string(pretty)
}

func doBracketsMatchInString(text string) bool {

  bracketBalance := 0

  for _, char := range text {
    if char == '{' {
      bracketBalance--
    } else if char == '}' {
      bracketBalance++
    }
  }

  if bracketBalance != 0 {
    return false
  }
  return true
}

func findNextBracketIndex(text []string, startingIndex int) int {
  bracketBalance := -1

  index := startingIndex

  for _, line := range text[startingIndex:] {
    if bracketBalance == 0 {
      return index
    }
    if strings.ContainsRune(line, '{') && !strings.ContainsRune(line, '}') {
      bracketBalance--
    } else if !strings.ContainsRune(line, '{') && strings.ContainsRune(line, '}') {
      bracketBalance++
    }
    index++
  }
  log.Fatal("Could not find next bracket because brackets were unmatched")
  return -1
}

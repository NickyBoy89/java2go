package parsing

import (
  "log"
  "strings"
  // "encoding/json"
)

func Convert(fileData string) {

  // parsed := make(map[string]interface{})

  log.Println(strings.Split(fileData, "\n"))

  var parsed []string

  lastWordIndex := 0

  for charind, char := range fileData {
    if char == ' ' {
      // parsed[fileData[lastWordIndex:charind]] = 0
      parsed = append(parsed, strings.Trim(fileData[lastWordIndex:charind], " "))
      lastWordIndex = charind
    }
  }

  log.Println(parsed)
}

func ParseFileContents(fileContents string) string {
  if !doBracketsMatchInString(fileContents) {
    log.Fatal("Brackets do not match in Java input")
  }

  contentLines := strings.Split(fileContents, "\n")

  // First we need to parse the first line of the class file
  fileMainClass := ParseClassLine(contentLines[0])

  // Already parsed the header line, and las free bracket will be on last line
  ParseClass(fileMainClass, contentLines[1:len(contentLines) - 2])

  return ""
}

func ParseClass(classDeclaration map[string]interface{}, classBody []string) map[string]interface{} {

  var parsedMethods []map[string]interface{}
  var parsedMemberVariables []map[string]interface{}

  inMethod := false
  indexUntilOutOfMethod := 0

  for li, line := range classBody {
    if line != "" {
      // Opening parenths should get method
      // signatures and declarations
      if strings.ContainsRune(line, '(') { // Parenths detected
        if strings.ContainsRune(line, '{') { // Is a declaration
          if strings.Contains(line, "class") {
            parsedMethods = append(parsedMethods, ParseClass(ParseClassLine(line), classBody[li + 1:findNextBracketIndex(classBody, li + 1)]))
            log.Printf("Found class: %v", ParseClass(ParseClassLine(line), classBody[li + 1:findNextBracketIndex(classBody, li + 1)])["name"])
            indexUntilOutOfMethod = findNextBracketIndex(classBody, li + 1)
            inMethod = true
          } else {
            parsedMethods = append(parsedMethods, ParseMethod(ParseMethodLine(line), classBody[li + 1:findNextBracketIndex(classBody, li + 1)]))
            log.Printf("Found method: %v", ParseMethod(ParseMethodLine(line), classBody[li + 1:findNextBracketIndex(classBody, li + 1)])["name"])
            indexUntilOutOfMethod = findNextBracketIndex(classBody, li + 1)
            inMethod = true
          }
        } else { // Is a method signature
          ParseMethodSignatureLine(line)
          log.Printf("Found methodSignature: %v", ParseMethodSignatureLine(line)["name"])
        }
      } else if !inMethod { // No parenthesies or brackets detected, must be a member variable
        parsedMemberVariables = append(parsedMemberVariables, ParseMemberVariableLine(line))
        log.Printf("Found memberVariable: %v", ParseMemberVariableLine(line)["name"])
      }
    }
    log.Println(li, indexUntilOutOfMethod)
    if li > indexUntilOutOfMethod {
      inMethod = false
    }
  }

  log.Println(parsedMethods)
  log.Println(parsedMemberVariables)

  return nil

}

func ParseMethod(methodDeclaration map[string]interface{}, methodBody []string) map[string]interface{} {
  result := methodDeclaration
  result["contentLines"] = methodBody

  return result

}

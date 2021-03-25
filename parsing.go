package main

import (
  "log"
  "strings"
  "encoding/json"
)

type JavaClass struct {
  Name string `json:"name"`
  DeclarationType string `json:"declarationType"`
  AccessModifiers []string `json:"accessModifiers"`
  Methods []JavaMethodItem `json:"methods"`
  MethodVariables []JavaMethodItem `json:"methodVariables"`
}

type JavaMethodItem struct {
  Name string `json:"name"`
  DeclarationType string `json:"declarationType"`
  ReturnType string `json:"returnType"`
  AccessModifiers []string `json:"accessModifiers"`
}

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
  fileMainClass := ParseDeclaration(contentLines[0])

  inMethod = false

  for _, line := range contentLines[1:] { // Already parsed the header line
    if line != "" {
      // Opening parenths should get method
      // signatures and declarations
      if strings.Contains(line, "(") {
        if strings.Contains(line, "{") { // Is a declaration
          ParseDeclaration(line)
        } else { // Is a method signature
          ParseMethodSignature(line)
        }
      } else {
        ParseMemberVariable(line)
      }
    }
  }

  return ""
}

// Parses the declaration of a class or method in the Java code
// Ex: protected abstract class TestClass {
// or: public int GetValue {
// You can detect a declaration by the existence of a '{' character in the line,
// so methods in interfaces do not count
func ParseDeclaration(dec string) interface{} {
  declarationWords := strings.Split(dec, " ")

  // Class declarations have the keyword "class"
  if strings.Contains(declarationWords, "class") {
    return JavaClass{
      Name: declarationWords[len(declarationWords) - 2],
      DeclarationType: "class",
      AccessModifiers: declarationWords[:len(declarationWords) - 3],
      Methods: make([]JavaMethodItem),
      MethodVariables: make([]JavaMethodItem),
    }
  }
  return JavaMethodItem{
    Name: declarationWords[len(declarationWords) - 2],
    DeclarationType: "method",
    ReturnType: declarationWords[len(declarationWords) - 3],
    AccessModifiers: declarationWords[:len(declarationWords) - 3],
  }
}

// A method signature is a method that doesn't have a body section enclosed
// in brackets, but can be found if there are no brackets but there are
// parenthesies
// Ex: (In an interface) public int getX()
func ParseMethodSignature(line string) JavaMethodItem {
  words := strings.Split(line, " ")

  return JavaMethodItem{
    // Strip the last two chars off (the '();' at the end of the method)
    Name: words[len(words) - 1][:len(words[len(words) - 1]) - 3],
    DeclarationType: "methodSignature",
    ReturnType: words[len(words) - 2],
    AccessModifiers: words[:len(words) - 2],
  }
}

// This one is identified by not having any '(' or '{', and is essentially
// just the odd one out (not a declaration or method signature)
// Ex: public int value;
func ParseMemberVariable(line string) JavaMethodItem {
  words := strings.Split(line, " ")

  return JavaMethodItem{
    // Strip the last char off (the semicolon)
    Name: words[len(words) - 1][:len(words[len(words) - 1]) - 1],
    DeclarationType: "memberVariable",
    ReturnType: words[len(words) - 2],
    AccessModifiers: words[:len(words) - 2],
  }
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

package parsing

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
  fileMainClass := ParseDeclarationLine(contentLines[0])

  inMethod := false

  for _, line := range contentLines[1:] { // Already parsed the header line
    if line != "" {
      // Opening parenths should get method
      // signatures and declarations
      if strings.Contains(line, "(") {
        if strings.Contains(line, "{") { // Is a declaration
          ParseDeclarationLine(line)
        } else { // Is a method signature
          ParseMethodSignatureLine(line)
        }
      } else {
        ParseMemberVariableLine(line)
      }
    }
  }

  return ""
}

func ParseDeclaration()

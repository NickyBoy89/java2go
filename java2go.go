package main

import (
  "os"
  "log"
  "io/ioutil"
  "strings"
  "encoding/json"

  "gitlab.nicholasnovak.io/snapdragon/java2go/parsing"
)

func main() {
  if len(os.Args) < 2 {
    log.Fatal("No files specified to convert")
  }

  for _, filePath := range os.Args[1:] {
    if strings.ContainsRune(filePath, '.') {
      if filePath[strings.LastIndex(filePath, "."):] != ".java" {
        continue
      }
    } else {
      continue
    }
    log.Printf("Started parsing file %v", filePath)

    contents, err := ioutil.ReadFile(filePath)
    if err != nil {
      log.Printf("Failed to read some input files: %v", err)
      continue
    }

    formatted, err := json.MarshalIndent(parsing.ParseFile(string(contents)), "", "  ")
    if err != nil {
      log.Fatal(err)
    }

    err = ioutil.WriteFile(ChangeFileExtension(filePath, ".json"), formatted, 0775)
    if err != nil {
      log.Fatal(err)
    }
    log.Printf("Compiled %v", filePath)
  }
}

func ChangeFileExtension(filePath, to string) string {
  if strings.ContainsRune(filePath, '/') {
    if strings.ContainsRune(filePath[strings.LastIndex(filePath, "/"):], '.') {
      return filePath[:strings.LastIndex(filePath, ".")] + to
    }
    return filePath + to
  }
  if strings.ContainsRune(filePath[strings.LastIndex(filePath, "/"):], '.') {
    return filePath[:strings.LastIndex(filePath, ".")] + to
  }
  return filePath + to
}

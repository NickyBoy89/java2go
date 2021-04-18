package main

import (
  "log"
  "os"
  "io/ioutil"
  "strings"
  "encoding/json"
  "flag"

  "gitlab.nicholasnovak.io/snapdragon/java2go/parsing"
)

func main() {
  outputDir := flag.String("o", "", "Directory to put the parsed files into, defaults to the same directory that the files appear in")
  dryRun := flag.Bool("dry-run", false, "Don't create the parsed files (check if parsing succeeds)")
  verbose := flag.Bool("v", false, "Additional debug info")

  flag.Parse()

  if len(flag.Args()) == 0 {
    log.Fatal("No files specified to convert")
  }

  for _, filePath := range flag.Args() {
    if !strings.ContainsRune(filePath, '.') || filePath[strings.LastIndex(filePath, "."):] != ".java" {
      if *verbose {
        log.Printf("Skipping file %v", filePath)
      }
      continue // Skips all non-java files
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

    fileDirectory := filePath[:strings.LastIndex(filePath, "/")]

    if !*dryRun {
      if *outputDir == "" {
        outputFile, err := os.OpenFile(ChangeFileExtension(filePath, ".json"), os.O_CREATE|os.O_WRONLY, 0775)
        if err != nil {
          log.Fatalf("Failed to open output file: %v", err)
        }
        _, err = outputFile.Write([]byte(formatted))
        if err != nil {
          log.Fatalf("Failed to write output file: %v", err)
        }
      } else {
        if _, err := os.Stat(*outputDir + "/" + fileDirectory); os.IsNotExist(err) {
          os.MkdirAll(*outputDir + "/" + fileDirectory, 0775)
        }
        outputFile, err := os.OpenFile(*outputDir + "/" + ChangeFileExtension(filePath, ".json"), os.O_WRONLY|os.O_CREATE, 0775)
        if err != nil {
          log.Fatalf("Failed to open output file: %v", err)
        }
        _, err = outputFile.Write([]byte(formatted))
        if err != nil {
          log.Fatalf("Failed to write output file: %v", err)
        }
      }

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

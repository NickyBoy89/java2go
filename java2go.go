package main

import (
  "os"
  "path/filepath"
  "io/fs"
  "io/ioutil"
  "strings"
  "encoding/json"
  "flag"

  log "github.com/sirupsen/logrus"

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

  walkDirFunc := func(path string, d fs.DirEntry, err error) error {
    if !strings.ContainsRune(path, '.') || path[strings.LastIndex(path, "."):] != ".java" {
      if *verbose {
        log.Debugf("Skipping file %v", path)
      }
      return nil // Skips all non-java files
    }
    if *verbose {
      log.Printf("Started parsing file %v", path)
    }

    contents, err := ioutil.ReadFile(path)
    if err != nil {
      return err
    }

    formatted, err := json.MarshalIndent(parsing.ParseFile(string(contents)), "", "  ")
    if err != nil {
      log.Fatal(err)
    }

    fileDirectory := path[:strings.LastIndex(path, "/")]

    if !*dryRun {
      if *outputDir == "" {
        outputFile, err := os.OpenFile(ChangeFileExtension(path, ".json"), os.O_CREATE|os.O_WRONLY, 0775)
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
        outputFile, err := os.OpenFile(*outputDir + "/" + ChangeFileExtension(path, ".json"), os.O_WRONLY|os.O_CREATE, 0775)
        if err != nil {
          log.Fatalf("Failed to open output file: %v", err)
        }
        _, err = outputFile.Write([]byte(formatted))
        if err != nil {
          log.Fatalf("Failed to write output file: %v", err)
        }
      }

    }

    if *verbose {
      log.Printf("Compiled %v", path)
    }
    if err != nil {
      return err
    }
    return nil
  }

  for _, filePath := range flag.Args() {
    err := filepath.WalkDir(filePath, walkDirFunc)
    log.Fatalf("Unable to parse directory %v: %v", filePath, err)
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

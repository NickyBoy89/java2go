package main

import (
  "os"
  "path/filepath"
  "io/fs"
  "io/ioutil"
  "strings"
  "encoding/json"
  "flag"
  "fmt"
  "runtime/pprof"

  log "github.com/sirupsen/logrus"

  "gitlab.nicholasnovak.io/snapdragon/java2go/parsing"
  "gitlab.nicholasnovak.io/snapdragon/java2go/goparser"
)

func main() {
  outputDir := flag.String("o", "", "Directory to put the parsed files into, defaults to the same directory that the files appear in")
  writeFlag := flag.Bool("w", false, "Create files directly instead of just writing to stdout")
  verbose := flag.Bool("v", false, "Additional debug info")
  // Cpu profiling
  cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
  // Testing options
  parseJson := flag.String("json", "", "Parses the specified Java file directly into the intemediary JSON, instead of generated code")

  flag.Parse()

  if *cpuprofile != "" {
    f, err := os.Create(*cpuprofile)
    if err != nil {
      log.Fatal(err)
    }
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
  }

  if *parseJson != "" {
    jsonFile, err := ioutil.ReadFile(*parseJson)
    if err != nil {
      log.Fatal(err)
    }
    generated := goparser.ParseFile(parsing.ParseFile(string(jsonFile)), true)
    fmt.Println(generated)

    formatted, err := json.MarshalIndent(parsing.ParseFile(string(jsonFile)), "", "  ")
    if err != nil {
      log.Fatal(err)
    }

    ioutil.WriteFile(
      ChangeFileExtension(*parseJson, ".json"),
      []byte(formatted),
      0775,
    )
    os.Exit(0)
  }

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

    if *writeFlag { // If writing is enabled through the -w tag
      if *outputDir == "" { // If outputDir flag is empty (default), place files in the default location
        ioutil.WriteFile(
          ChangeFileExtension(path, ".go"), // Change the output file to a .json
          []byte(goparser.ParseFile(parsing.ParseFile(string(contents)), true)), // Pass the parsed json into the goparser
          0775,
        )
      } else {
        if _, err := os.Stat(*outputDir + "/" + fileDirectory); os.IsNotExist(err) {
          os.MkdirAll(*outputDir + "/" + fileDirectory, 0775)
        }
        outputFile, err := os.OpenFile(*outputDir + "/" + ChangeFileExtension(path, ".json"), os.O_WRONLY|os.O_CREATE, 0775)
        defer outputFile.Close()
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
    if err != nil {
      log.Fatalf("Unable to parse file or directory %v: %v", filePath, err)
    }
    log.Printf("Parsed file or directory %v", filePath)
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

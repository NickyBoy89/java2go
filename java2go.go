package main

import (
  "os"
  "os/exec"
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

var synchronizedFlag bool

func main() {
  outputDir := flag.String("o", "", "Directory to put the parsed files into, defaults to the same directory that the files appear in")
  writeFlag := flag.Bool("w", false, "Create files directly instead of just writing to stdout")
  verbose := flag.Bool("v", false, "Additional debug info")
  skipImports := flag.Bool("skip-imports", false, "Skip the process of automatically setting the imports of the generated files with goimports")
  // Cpu profiling
  cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
  // Testing options
  parseJson := flag.String("json", "", "Parses the specified Java file directly into the intemediary JSON, instead of generated code")
  sync := flag.Bool("sync", false, "Parses the files in-order with no multithreading")

  flag.Parse()

  // Synchronization
  synchronizedFlag = *sync

  // CPU profiling
  if *cpuprofile != "" {
    f, err := os.Create(*cpuprofile)
    if err != nil {
      log.Fatal(err)
    }
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
  }

  // Json flag
  if *parseJson != "" {
    jsonFile, err := ioutil.ReadFile(*parseJson)
    if err != nil {
      log.Fatal(err)
    }
    generated := goparser.ParseFile(parsing.ParseFile(string(jsonFile)), true, "main")
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
    // Synchronized global being true means that the files are parsed in-order
    if !synchronizedFlag {
      go ParseFile(path, verbose, writeFlag, skipImports, outputDir)
    } else {
      ParseFile(path, verbose, writeFlag, skipImports, outputDir)
    }
    return nil
  }

  // Parse each file on a new goroutine
  for _, filePath := range flag.Args() {
    err := filepath.WalkDir(filePath, walkDirFunc)
    if err != nil {
      log.Fatalf("Unable to parse file or directory %v: %v", filePath, err)
    }
    log.Printf("Parsed file or directory %v", filePath)
  }
}

func ParseFile(path string, verbose, writeFlag, skipImports *bool, outputDir *string) error {
  // Only parse .java files
  if !strings.ContainsRune(path, '.') || path[strings.LastIndex(path, "."):] != ".java" {
    if *verbose {
      log.Debugf("Skipping file %v", path)
    }
    return nil
  }

  // Verbose
  if *verbose {
    log.Printf("Started parsing file %v", path)
  }

  // Reads contents of java file into memory
  contents, err := ioutil.ReadFile(path)
  if err != nil {
    return err
  }

  // Gets the current path of directories to the file
  lastSlashInd := strings.LastIndex(path, "/")
  filePath := path[:lastSlashInd]

  // Gets the current relative directory that the file is in
  fileDirectory := path[strings.LastIndex(path[:lastSlashInd], "/") + 1:lastSlashInd]

  // If writing is enabled through the -w tag
  if *writeFlag {
    // If outputDir flag is not empty, create a folder (if needed) and alter the output directory
    if *outputDir != "" {
      // Add a slash to the output directory to make it a valid directory
      *outputDir = *outputDir + "/"
      // If a folder does not exist for the output files, then create one
      if _, err := os.Stat(*outputDir + filePath); os.IsNotExist(err) {
        os.MkdirAll(*outputDir + filePath, 0775)
      }
    }

    ioutil.WriteFile(
      *outputDir + ChangeFileExtension(path, ".go"), // Change the output file to .go
      []byte(goparser.ParseFile(parsing.ParseFile(string(contents)), true, fileDirectory)), // Parse the contents of the file
      0775,
    )

    // Auto-import dependencies
    if !*skipImports {
      // Run goimports to automatically generate imports
      goImport := exec.Command("goimports", "-w", ChangeFileExtension(path, ".go"))
      out, err := goImport.Output()
      if err != nil {
        if err.Error() == "exit status 2" { // Exit status 2 usually means that the generated file is not valid go code
          if *verbose {
            log.Warn("Error automatically generating imports, exit status 2")
            log.Warn("This can mean that goimports is malfunctioning, but usually just means that the generated code is not completely valid")
          } else {
            log.Warn("Invalid go code generated, exit status 2")
          }
        } else {
          log.Error("Automatically generating imports on generated files failed with the following error")
          log.Error("Please fix the error below to continue, or disable automatic import generation with the --skip-imports flag")
          log.Fatal(out, err)
        }
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

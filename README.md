# Java2go
## About

Java2go is a program that is intended to automatically convert Java files to Golang

It does this in three intermediary steps:

* Files are taken in through parsing in either a file or directory to the `java2go` binary, and all non-java files are ignored

* Files are parsed from java into JSON in the `parsing` package

* Files are parsed from json to Golang in the `golang` package and written to the output directory

## Usage

* Clone the repo

* `go build` to build the java2go binary

* `./java2go <files/directories>...` to parse files

## Options

### Command-line args:

* `-v` Verbose mode: every file being parsed is displayed

* `-w` Writes files directly, instead of checking if they can be parsed successfully

* `-o` Output dir: specify a custom directory that the files will be parsed into. By default, the files are put in the same folders as the inputs

* `--skip-imports` Skips the process of automatically adding imports for the generated files with goimports

#### Testing Args

* `--json` Parses the code and outputs the intemediary json format that the tool uses internally

* `--cpuprofile` The file path for a CPU profile to be written to during profiling

* `--sync` Disables multithreaded parsing of the files, and parses them sequentially

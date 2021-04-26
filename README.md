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

* `-v` verbose mode: every file being parsed is displayed

* `--dry-run` the files are parsed, but not written to disk

* `-o` output dir: specify a custom directory, by default, all the files will be placed alongside their original java files with the same directory structure

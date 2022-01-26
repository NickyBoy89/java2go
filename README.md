# Java2go
## About

Java2go is a program that is intended to automatically convert Java files to Golang

The process that it does this is througb several intemediary steps

1. Parse the java source code with the [golang bindings for the tree-sitter](git@github.com:smacker/go-tree-sitter.git) java parser into a `tree-sitter` AST

2. Parse the bindings into Golang's own [AST representation](https://pkg.go.dev/go/ast)

3. Use Golang's builtin [AST printer](https://pkg.go.dev/go/printer) to print out the generated code

## Usage

* Clone the repo

* `go build` to build the java2go binary

* `./java2go files...` to parse a list of files

## Options

* `-w` writes the files directly to their corresponding `.go` files, instead of `stdout`

* `-q` Prevents the outputs of the parsed files from appearing on stdout

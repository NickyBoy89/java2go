package tests

import (
	"bytes"
	"context"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/NickyBoy89/java2go/ng"
	"github.com/NickyBoy89/java2go/parsing"
	sitter "github.com/tree-sitter/go-tree-sitter"
	java "github.com/tree-sitter/tree-sitter-java/bindings/go"
)

func ParseFile(fileName string, t *testing.T) ast.Node {
	file, err := parsing.ReadSource(fileName)
	if err != nil {
		t.Fatal(err)
	}

	tree, err := ng.ParseProgram(file)
	if err != nil {
		t.Fatal(err)
	}

	return tree
}

func WriteProgramText(w io.Writer, tree ast.Node) error {
	return printer.Fprint(w, token.NewFileSet(), tree)
}

func ParseText(text string) *sitter.Tree {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(java.Language()))
	return parser.ParseCtx(context.Background(), []byte(text), nil)
}

func CompareParseFiles(inputFileName string, expectedFileName string, t *testing.T) (string, string) {
	expected, err := os.ReadFile(expectedFileName)
	if err != nil {
		t.Fatal(err)
	}

	var actual strings.Builder
	if err := WriteProgramText(&actual, ParseFile(inputFileName, t)); err != nil {
		t.Fatal(err)
	}

	return string(expected), actual.String()
}

func GetParsedProgram(javaText string, t *testing.T) ast.Node {
	source := ParseText(javaText)
	tree, err := ng.ParseProgram(parsing.SourceFile{
		Source: []byte(javaText),
		Ast:    source.RootNode(),
	})
	if err != nil {
		t.Fatal(err)
	}

	return tree
}

// `compareProgramOutputs` takes in Java source code, runs it, transpiles it,
// and tests if the output from the original code is the same as the transpiled
// output
func compareProgramOutputs(javaInput string, t *testing.T) {
	srcFile, err := os.CreateTemp("", "*.java")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(srcFile.Name())

	// Write the Java source file
	if _, err := srcFile.WriteString(javaInput); err != nil {
		t.Fatal(err)
	}

	transpiledSrc, err := os.CreateTemp("", "*.go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(transpiledSrc.Name())

	// Write the transpiled file
	WriteProgramText(transpiledSrc, GetParsedProgram(javaInput, t))

	programOutput, err := os.CreateTemp("", "java2go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(programOutput.Name())

	// Run the source code and capture the output
	javaCmd := exec.Command("java", srcFile.Name())
	javaCmd.Stdout = programOutput
	javaCmd.Stderr = os.Stderr

	if err := javaCmd.Run(); err != nil {
		t.Fatal(err)
	}

	transpiledOutput, err := os.CreateTemp("", "java2go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(transpiledOutput.Name())

	// Run the transpiled code and capture the output
	goCmd := exec.Command("go", "run", transpiledSrc.Name())
	goCmd.Stdout = transpiledOutput
	goCmd.Stderr = os.Stderr

	if err := goCmd.Run(); err != nil {
		t.Fatal(err)
	}

	diffFiles(transpiledOutput.Name(), programOutput.Name(), t)
}

func diffFiles(fstName, sndName string, t *testing.T) {
	diffCmd := exec.Command("git", "diff", "--color=always", "--no-index", fstName, sndName)

	var diffOutput bytes.Buffer
	diffCmd.Stdout = &diffOutput
	diffCmd.Stderr = os.Stderr

	err := diffCmd.Run()
	if err != nil {
		exitCode := err.(*exec.ExitError).ExitCode()

		// Exit code of 1 means that the files were not equal
		if exitCode == 1 {
			t.Log("error: generated code did not match")
			t.Error(diffOutput.String())
		} else if exitCode != 0 {
			t.Fatal(err)
		}
	}
}

func ComparePrograms(javaInput string, goOutput string, t *testing.T) {

	f1, err := os.CreateTemp("", "java2go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f1.Name())

	f2, err := os.CreateTemp("", "java2go")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(f2.Name())

	javaOutAst := GetParsedProgram(javaInput, t)
	WriteProgramText(f1, javaOutAst)

	if _, err := f2.WriteString(goOutput); err != nil {
		t.Fatal(err)
	}

	diffFiles(f1.Name(), f2.Name(), t)
}

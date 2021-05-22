package goparser

import (
	"testing"
	"encoding/json"

	"github.com/sergi/go-diff/diffmatchpatch"

	"gitlab.nicholasnovak.io/snapdragon/java2go/codeparser"
)

func TestParseVariable(t *testing.T) {
	testVar := "int value = 1;"
	testResult := "value := 1"
	testVar2 := "int value=1;" // Just without spacing

	DoubleTestTemplate(testVar, testVar2, testResult, t)
}

func TestParseVariableAlreadyCreated(t *testing.T) {
	testVar := "value = 1;"
	testResult := "value = 1"
	testVar2 := "value=1;" // Just without spacing

	DoubleTestTemplate(testVar, testVar2, testResult, t)
}

func TestParseFunctionFromDifferentPackage(t *testing.T) {
	testVar := "int pi = Math.GetPi();"
	testResult := "pi := Math.GetPi()"

	SingleTestTemplate(testVar, testResult, t)
}

func TestParseVarWithFunction(t *testing.T) {
	testVar := "Node curNode = GetNode();"
	testResult := "curNode := GetNode()"

	SingleTestTemplate(testVar, testResult, t)
}

func TestParseGenericType(t *testing.T) {
	testVar := "ArrayList<String> wordCounts = new ArrayList<String>();"
	testResult := "word := make([]string)"

	SingleTestTemplate(testVar, testResult, t)
}

func TestParseForLoop(t *testing.T) {
	testVar := `for (int i = 0; i < N; i++){
System.out.println(i);
}`
	// Same thing as above, but with no space in front of the start of the expression
	testVar2 := `for(int i = 0; i < N; i++){
System.out.println(i);
}`
	testResult := `for i := 0; i < N; i++ {
fmt.Println(i)
}`

	DoubleTestTemplate(testVar, testVar2, testResult, t)
}

func SingleTestTemplate(testVar, testResult string, ts *testing.T) {
	parsed, err := json.MarshalIndent(codeparser.ParseContent(testVar), "", "  ")
	if err != nil {
		ts.Fatal(err)
	}

	if string(parsed) != string(testResult) {
		diff := diffmatchpatch.New()
    ts.Log(diff.DiffPrettyText(diff.DiffMain(string(parsed), string(testResult), false)))
    ts.Error("Result and Original did not match")
	}
}

func DoubleTestTemplate(testVar, testVar2, testResult string, ts *testing.T) {
	parsed, err := json.MarshalIndent(codeparser.ParseContent(testVar), "", "  ")
	if err != nil {
		ts.Fatal(err)
	}
	parsed2, err := json.MarshalIndent(codeparser.ParseContent(testVar2), "", "  ")
	if err != nil {
		ts.Fatal(err)
	}

	if string(parsed) != string(testResult) {
		diff := diffmatchpatch.New()
    ts.Log(diff.DiffPrettyText(diff.DiffMain(string(parsed), string(testResult), false)))
    ts.Error("Result and Original did not match")
	}

	if string(parsed2) != string(testResult) {
		diff := diffmatchpatch.New()
    ts.Log(diff.DiffPrettyText(diff.DiffMain(string(parsed2), string(testResult), false)))
    ts.Error("Result and Original did not match")
	}
}

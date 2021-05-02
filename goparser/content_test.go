package goparser

import (
	"testing"
)

func TestParseVariable(t *testing.T) {
	testVar := "int value = 1;"
	testResult := "value := 1"
	testVar2 := "int value=1;" // Just without spacing

	if ParseContent(testVar) != testResult {
		t.Errorf("Original: [%s] and parsed: [%s] were not the same", testResult, ParseContent(testVar))
	}

	if ParseContent(testVar2) != testResult {
		t.Errorf("Original: [%s] and parsed: [%s] were not the same", testResult, ParseContent(testVar2))
	}
}

func TestParseVariableAlreadyCreated(t *testing.T) {
	testVar := "value = 1;"
	testResult := "value = 1"
	testVar2 := "value=1;" // Just without spacing

	if ParseContent(testVar) != testResult {
		t.Errorf("Original: [%s] and parsed: [%s] were not the same", testResult, ParseContent(testVar))
	}

	if ParseContent(testVar2) != testResult {
		t.Errorf("Original: [%s] and parsed: [%s] were not the same", testResult, ParseContent(testVar2))
	}
}

func TestParseVarWithFunction(t *testing.T) {
	testVar := "Node curNode = GetNode();"
	testResult := "curNode := GetNode()"

	if ParseContent(testVar) != testResult {
		t.Errorf("Original: [%s] and parsed: [%s] were not the same", testResult, ParseContent(testVar))
	}
}

func TestParseFunctionFromDifferentPackage(t *testing.T) {
	testVar := "int pi = Math.GetPi();"
	testResult := "pi := Math.GetPi()"

	if ParseContent(testVar) != testResult {
		t.Errorf("Original: [%s] and parsed: [%s] were not the same", testResult, ParseContent(testVar))
	}
}

func TestParseGenericType(t *testing.T) {
	testVar := "ArrayList<String> wordCounts = new ArrayList<String>();"
	testResult := "word := make([]string)"

	if ParseContent(testVar) != testResult {
		t.Errorf("Original: [%s] and parsed: [%s] were not the same", testResult, ParseContent(testVar))
	}
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

	if ParseContent(testVar) != testResult {
		t.Errorf("Original: [%s] and parsed: [%s] were not the same", testResult, ParseContent(testVar))
	}

	if ParseContent(testVar2) != testResult {
		t.Errorf("Original: [%s] and parsed: [%s] were not the same", testResult, ParseContent(testVar2))
	}
}

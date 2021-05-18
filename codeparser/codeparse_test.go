package codeparser

import (
	"encoding/json"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestLoopConditional(t *testing.T) {
	test := "i <= 10"
	result := []LineType{
		LineType{
			Name: "LocalVariableOrExpression",
			Words: map[string]interface{}{
				"Expression": "i",
			},
		},
		LineType{
			Name: "LocalVariableOrExpression",
			Words: map[string]interface{}{
				"Expression": "<=",
			},
		},
		LineType{
			Name: "LocalVariableOrExpression",
			Words: map[string]interface{}{
				"Expression": "10",
			},
		},
	}

	parsedTest, err := json.MarshalIndent(ParseExpression(test), "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	parsedResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if string(parsedTest) != string(parsedResult) {
		diff := diffmatchpatch.New()
    t.Log(diff.DiffPrettyText(diff.DiffMain(string(parsedTest), string(parsedResult), false)))
    t.Error("Result and Original did not match")
	}
}

func TestPlusEqualsModifier(t *testing.T) {
	test := "result += this.get(0)"
	result := LineType{
		Name: "CompoundAssignment",
		Words: map[string]interface{}{
			"Operator": "+",
			"VariableName": "result",
			"Expression": []LineType{
				LineType{
					Name: "RemoteVariableOrExpression",
					Words: map[string]interface{}{
						"Expression": "get(0)",
						"RemotePackage": "this",
					},
				},
			},
		},
	}

	parsedTest, err := json.MarshalIndent(ParseLine(test), "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	parsedResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if string(parsedTest) != string(parsedResult) {
		diff := diffmatchpatch.New()
		t.Log(diff.DiffPrettyText(diff.DiffMain(string(parsedTest), string(parsedResult), false)))
    t.Error("Result and Original did not match")
	}
}

func TestImplicitArrayCreation(t *testing.T) {
	test := "int[] values = {1, 2, 3, 4}"
	result := LineType{
		Name: "CreateAndAssignVariable",
		Words: map[string]interface{}{
			"VariableName": []LineType{
				LineType{
					Name: "LocalVariableOrExpression",
					Words: map[string]interface{}{
						"Expression": "values",
					},
				},
			},
			"VariableType": "int[]",
			"Expression": []LineType{
				LineType{
					Name: "ImplicitArrayAssignment",
					Words: map[string]interface{}{
						"ArrayType": "int[]",
						"Elements": []LineType{
							LineType{
								Name: "LocalVariableOrExpression",
								Words: map[string]interface{}{
									"Expression": "1",
								},
							},
							LineType{
								Name: "LocalVariableOrExpression",
								Words: map[string]interface{}{
									"Expression": "2",
								},
							},
							LineType{
								Name: "LocalVariableOrExpression",
								Words: map[string]interface{}{
									"Expression": "3",
								},
							},
							LineType{
								Name: "LocalVariableOrExpression",
								Words: map[string]interface{}{
									"Expression": "4",
								},
							},
						},
					},
				},
			},
		},
	}

	parsedTest, err := json.MarshalIndent(ParseLine(test), "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	parsedResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if string(parsedTest) != string(parsedResult) {
		diff := diffmatchpatch.New()
		t.Log(diff.DiffPrettyText(diff.DiffMain(string(parsedTest), string(parsedResult), false)))
    t.Error("Result and Original did not match")
	}
}

func TestNewConstructor(t *testing.T) {
	test := "throw new AssertionError(list.size())"
	result := LineType{
		Name: "ThrowException",
		Words: map[string]interface{}{
			"Expression": LineType{
				Name: "NewConstructor",
				Words: map[string]interface{}{
					"Expression": LineType{
						Name: "FunctionCall",
						Words: map[string]interface{}{
							"FunctionName": "AssertionError",
							"Parameters": []LineType{
								LineType{
									Name: "RemoteVariableOrExpression",
									Words: map[string]interface{}{
										"Expression": "size()",
										"RemotePackage": "list",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	parsedTest, err := json.MarshalIndent(ParseLine(test), "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	parsedResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if string(parsedTest) != string(parsedResult) {
		diff := diffmatchpatch.New()
		t.Log(diff.DiffPrettyText(diff.DiffMain(string(parsedTest), string(parsedResult), false)))
    t.Error("Result and Original did not match")
	}
}

func TestNewSimpleObject(t *testing.T) {
	test := "this.head = new Node(element)"
	result := LineType{
		Name: "AssignVariable",
		Words: map[string]interface{}{
			"VariableName": []LineType{
				LineType{
					Name: "RemoteVariableOrExpression",
					Words: map[string]interface{}{
						"Expression": "head",
						"RemotePackage": "this",
					},
				},
			},
			"Expression": []LineType{
				LineType{
					Name: "NewConstructor",
					Words: map[string]interface{}{
						"Expression": LineType{
							Name: "FunctionCall",
							Words: map[string]interface{}{
								"FunctionName": "Node",
								"Parameters": []LineType{
									LineType{
										Name: "LocalVariableOrExpression",
										Words: map[string]interface{}{
											"Expression": "element",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	parsedTest, err := json.MarshalIndent(ParseLine(test), "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	parsedResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if string(parsedTest) != string(parsedResult) {
		diff := diffmatchpatch.New()
		t.Log(diff.DiffPrettyText(diff.DiffMain(string(parsedTest), string(parsedResult), false)))
    t.Error("Result and Original did not match")
	}
}

func TestExpressionInFunctionCalls(t *testing.T) {
	test := `throw new AssertionError("Expected list of length " + answer.length + " but got " + list.size())`
	result := LineType{
		Name: "ThrowException",
		Words: map[string]interface{}{
			"Expression": LineType{
				Name: "NewConstructor",
				Words: map[string]interface{}{
					"Expression": LineType{
						Name: "FunctionCall",
						Words: map[string]interface{}{
							"FunctionName": "AssertionError",
							"Parameters": []LineType{
								LineType{
									Name: "StringLiteral",
									Words: map[string]interface{}{
										"String": "\"Expected list of length \"",
									},
								},
								LineType{
									Name: "LocalVariableOrExpression",
									Words: map[string]interface{}{
										"Expression": "+",
									},
								},
								LineType{
									Name: "RemoteVariableOrExpression",
									Words: map[string]interface{}{
										"Expression": "length",
										"RemotePackage": "answer",
									},
								},
								LineType{
									Name: "LocalVariableOrExpression",
									Words: map[string]interface{}{
										"Expression": "+",
									},
								},
								LineType{
									Name: "StringLiteral",
									Words: map[string]interface{}{
										"String": "\" but got \"",
									},
								},
								LineType{
									Name: "LocalVariableOrExpression",
									Words: map[string]interface{}{
										"Expression": "+",
									},
								},
								LineType{
									Name: "RemoteVariableOrExpression",
									Words: map[string]interface{}{
										"Expression": "size()",
										"RemotePackage": "list",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	parsedTest, err := json.MarshalIndent(ParseLine(test), "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	parsedResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if string(parsedTest) != string(parsedResult) {
		diff := diffmatchpatch.New()
		t.Log(diff.DiffPrettyText(diff.DiffMain(string(parsedTest), string(parsedResult), false)))
    t.Error("Result and Original did not match")
	}
}

func TestAppendExpression(t *testing.T) {
	test := `"Expected list of length " + answer.length + " but got " + list.size()`
	result := []LineType{
		LineType{
			Name: "StringLiteral",
			Words: map[string]interface{}{
				"String": "\"Expected list of length \"",
			},
		},
		LineType{
			Name: "LocalVariableOrExpression",
			Words: map[string]interface{}{
				"Expression": "+",
			},
		},
		LineType{
			Name: "RemoteVariableOrExpression",
			Words: map[string]interface{}{
				"Expression": "length",
				"RemotePackage": "answer",
			},
		},
		LineType{
			Name: "LocalVariableOrExpression",
			Words: map[string]interface{}{
				"Expression": "+",
			},
		},
		LineType{
			Name: "StringLiteral",
			Words: map[string]interface{}{
				"String": "\" but got \"",
			},
		},
		LineType{
			Name: "LocalVariableOrExpression",
			Words: map[string]interface{}{
				"Expression": "+",
			},
		},
		LineType{
			Name: "RemoteVariableOrExpression",
			Words: map[string]interface{}{
				"Expression": "size()",
				"RemotePackage": "list",
			},
		},
	}

	parsedTest, err := json.MarshalIndent(ParseExpression(test), "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	parsedResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if string(parsedTest) != string(parsedResult) {
		diff := diffmatchpatch.New()
		t.Log(diff.DiffPrettyText(diff.DiffMain(string(parsedTest), string(parsedResult), false)))
    t.Error("Result and Original did not match")
	}
}

func TestParseCharExpression(t *testing.T) {
	test := "char first = 'A'"
	result := []LineType{
		LineType{
			Name: "LocalVariableOrExpression",
			Words: map[string]interface{}{
				"Expression": "char",
			},
		},
		LineType{
			Name: "LocalVariableOrExpression",
			Words: map[string]interface{}{
				"Expression": "first",
			},
		},
		LineType{
			Name: "LocalVariableOrExpression",
			Words: map[string]interface{}{
				"Expression": "=",
			},
		},
		LineType{
			Name: "RuneLiteral",
			Words: map[string]interface{}{
				"Rune": "A",
			},
		},
	}

	parsedTest, err := json.MarshalIndent(ParseExpression(test), "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	parsedResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if string(parsedTest) != string(parsedResult) {
		diff := diffmatchpatch.New()
		t.Log(diff.DiffPrettyText(diff.DiffMain(string(parsedTest), string(parsedResult), false)))
    t.Error("Result and Original did not match")
	}
}

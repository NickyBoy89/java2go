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
						"Expression": []LineType{
							LineType{
								Name: "FunctionCall",
								Words: map[string]interface{}{
									"FunctionName": "get",
									"Parameters": [][]LineType{
										[]LineType{
											LineType{
												Name: "LocalVariableOrExpression",
												Words: map[string]interface{}{
													"Expression": "0",
												},
											},
										},
									},
								},
							},
						},
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
						"Elements": [][]LineType{
							[]LineType{
								LineType{
									Name: "LocalVariableOrExpression",
									Words: map[string]interface{}{
										"Expression": "1",
									},
								},
							},
							[]LineType{
								LineType{
									Name: "LocalVariableOrExpression",
									Words: map[string]interface{}{
										"Expression": "2",
									},
								},
							},
							[]LineType{
								LineType{
									Name: "LocalVariableOrExpression",
									Words: map[string]interface{}{
										"Expression": "3",
									},
								},
							},
							[]LineType{
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
							"Parameters": [][]LineType{
								[]LineType{
									LineType{
										Name: "RemoteVariableOrExpression",
										Words: map[string]interface{}{
											"Expression": []LineType{
												LineType{
													Name: "FunctionCall",
													Words: map[string]interface{}{
														"FunctionName": "size",
														"Parameters": []LineType{},
													},
												},
											},
											"RemotePackage": "list",
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

func TestNewSimpleObject(t *testing.T) {
	test := "this.head = new Node(element)"
	result := LineType{
		Name: "AssignVariable",
		Words: map[string]interface{}{
			"VariableName": []LineType{
				LineType{
					Name: "RemoteVariableOrExpression",
					Words: map[string]interface{}{
						"Expression": []LineType{
							LineType{
								Name: "LocalVariableOrExpression",
								Words: map[string]interface{}{
									"Expression": "head",
								},
							},
						},
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
								"Parameters": [][]LineType{
									[]LineType{
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
							"Parameters": [][]LineType{
								[]LineType{
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
											"Expression": []LineType{
												LineType{
													Name: "LocalVariableOrExpression",
													Words: map[string]interface{}{
														"Expression": "length",
													},
												},
											},
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
											"Expression": []LineType{
												LineType{
													Name: "FunctionCall",
													Words: map[string]interface{}{
														"FunctionName": "size",
														"Parameters": []LineType{},
													},
												},
											},
											"RemotePackage": "list",
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
				"Expression": []LineType{
					LineType{
						Name: "LocalVariableOrExpression",
						Words: map[string]interface{}{
							"Expression": "length",
						},
					},
				},
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
				"Expression": []LineType{
					LineType{
						Name: "FunctionCall",
						Words: map[string]interface{}{
							"FunctionName": "size",
							"Parameters": []LineType{},
						},
					},
				},
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

func TestParseNewArraylist(t *testing.T) {
	test := "IntLinkedList list = new IntLinkedList()"
	result := LineType{
		Name: "CreateAndAssignVariable",
		Words: map[string]interface{}{
			"VariableName": []LineType{
				LineType{
					Name: "LocalVariableOrExpression",
					Words: map[string]interface{}{
						"Expression": "list",
					},
				},
			},
			"VariableType": "IntLinkedList",
			"Expression": []LineType{
				LineType{
					Name: "NewConstructor",
					Words: map[string]interface{}{
						"Expression": LineType{
							Name: "FunctionCall",
							Words: map[string]interface{}{
								"FunctionName": "IntLinkedList",
								"Parameters": []LineType{},
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

func TestConstructArray(t *testing.T) {
	test := "new int[0]"
	result := []LineType{
		LineType{
			Name: "ConstructArray",
			Words: map[string]interface{}{
				"ArrayType": "int",
				"InitialSize": "0",
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

func TestParseEnhancedForLoop(t *testing.T) {
	test := `for (Compass c : Compass.values()) {
    System.out.println(c);
  }`
	result := []LineTyper{
		LineBlock{
			Name: "EnhancedForLoop",
			Words: map[string]interface{}{
				"DeclarationType": "Compass",
				"DeclarationName": "c",
				"Iterable": []LineType{
					LineType{
						Name: "RemoteVariableOrExpression",
						Words: map[string]interface{}{
							"Expression": []LineType{
								LineType{
									Name: "FunctionCall",
									Words: map[string]interface{}{
										"FunctionName": "values",
										"Parameters": []LineType{},
									},
								},
							},
							"RemotePackage": "Compass",
						},
					},
				},
			},
			Lines: []LineTyper{
				LineType{
					Name: "GenericLine",
					Words: map[string]interface{}{
						"Statement": []LineType{
							LineType{
								Name: "RemoteVariableOrExpression",
								Words: map[string]interface{}{
									"Expression": []LineType{
										LineType{
											Name: "RemoteVariableOrExpression",
											Words: map[string]interface{}{
												"Expression": []LineType{
													LineType{
														Name: "FunctionCall",
														Words: map[string]interface{}{
															"FunctionName": "println",
															"Parameters": [][]LineType{
																[]LineType{
																	LineType{
																		Name: "LocalVariableOrExpression",
																		Words: map[string]interface{}{
																			"Expression": "c",
																		},
																	},
																},
															},
														},
													},
												},
												"RemotePackage": "out",
											},
										},
									},
									"RemotePackage": "System",
								},
							},
						},
					},
				},
			},
		},
	}

	parsedTest, err := json.MarshalIndent(ParseContent(test), "", "  ")
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

func TestParseUnholyJava(t *testing.T) {
	test := "GLFWVidMode gLFWVidMode = l == 0L ? null : GLFW.glfwGetVideoMode(l)"
	result := LineType{
		Name: "CreateAndAssignVariable",
		Words: map[string]interface{}{
			"VariableName": []LineType{
				LineType{
					Name: "LocalVariableOrExpression",
					Words: map[string]interface{}{
						"Expression": "gLFWVidMode",
					},
				},
			},
			"VariableType": "GLFWVidMode",
			"Expression": []LineType{
				LineType{
					Name: "LocalVariableOrExpression",
					Words: map[string]interface{}{
						"Expression": "l",
					},
				},
				LineType{
					Name: "ComparisonOperator",
					Words: map[string]interface{}{
						"Operator": "==",
					},
				},
				LineType{
					Name: "LocalVariableOrExpression",
					Words: map[string]interface{}{
						"Expression": "0L",
					},
				},
				LineType{
					Name: "TernaryOperator",
					Words: map[string]interface{}{
						"TrueExpression": []LineType{
							LineType{
								Name: "LocalVariableOrExpression",
								Words: map[string]interface{}{
									"Expression": "null",
								},
							},
						},
						"FalseExpression": []LineType{
							LineType{
								Name: "RemoteVariableOrExpression",
								Words: map[string]interface{}{
									"RemotePackage": "GLFW",
									"Expression": []LineType{
										LineType{
											Name: "FunctionCall",
											Words: map[string]interface{}{
												"FunctionName": "glfwGetVideoMode",
												"Parameters": [][]LineType{
													[]LineType{
														LineType{
															Name: "LocalVariableOrExpression",
															Words: map[string]interface{}{
																"Expression": "l",
															},
														},
													},
												},
											},
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

func TestParseTypeAssertion(t *testing.T) {
	test := "int value = (int)12"
	result := LineType{
		Name: "CreateAndAssignVariable",
		Words: map[string]interface{}{
			"VariableName": []LineType{
				LineType{
					Name: "LocalVariableOrExpression",
					Words: map[string]interface{}{
						"Expression": "value",
					},
				},
			},
			"VariableType": "int",
			"Expression": []LineType{
				LineType{
					Name: "TypeAssertion",
					Words: map[string]interface{}{
						"AssertedType": "int",
					},
				},
				LineType{
					Name: "LocalVariableOrExpression",
					Words: map[string]interface{}{
						"Expression": "12",
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

func TestParseDoubleNestedArray(t *testing.T) {
	test := "return new int[][]{{1, 2, 3}}"
	result := LineType{
		Name: "ReturnStatement",
		Words: map[string]interface{}{
			"Expression": []LineType{
				LineType{
					Name: "ConstructArrayWithImplicit",
					Words: map[string]interface{}{
						"ArrayType": "int[]",
						"Elements": []LineType{
							LineType{
								Name: "ImplicitArrayAssignment",
								Words: map[string]interface{}{
									"ArrayType": "int[]",
									"Elements": [][]LineType{
										[]LineType{
											LineType{
												Name: "ImplicitArrayAssignment",
												Words: map[string]interface{}{
													"ArrayType": "int",
													"Elements": [][]LineType{
														[]LineType{
															LineType{
																Name: "LocalVariableOrExpression",
																Words: map[string]interface{}{
																	"Expression": "1",
																},
															},
														},
														[]LineType{
															LineType{
																Name: "LocalVariableOrExpression",
																Words: map[string]interface{}{
																	"Expression": "2",
																},
															},
														},
														[]LineType{
															LineType{
																Name: "LocalVariableOrExpression",
																Words: map[string]interface{}{
																	"Expression": "3",
																},
															},
														},
													},
												},
											},
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

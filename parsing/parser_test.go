package parsing

import (
  "io/ioutil"
  "testing"
  "reflect"
)

func TestParseDeclaration(t *testing.T) {
  testClassDec := "public protected static class TestClass {"
  parsedClassDec := map[string]interface{}{
    "name": "TestClass",
    "declarationType": "class",
    "accessModifiers": []string{"public", "protected", "static"},
    "methods": []map[string]interface{}{},
    "methodVariables": []map[string]interface{}{},
  }
  testMethodDec := "public int GetValue() {"
  parsedMethodDec := map[string]interface{}{
    "name": "GetValue",
    "declarationType": "method",
    "returnType": "int",
    "accessModifiers": []string{"public"},
    "contentLines": []string{},
  }

  if !reflect.DeepEqual(ParseClassLine(testClassDec), parsedClassDec) {
    t.Errorf("Parsing example class failed: %v did not match %v", ParseClassLine(testClassDec), parsedClassDec)
  }

  if !reflect.DeepEqual(ParseMethodLine(testMethodDec), parsedMethodDec) {
    t.Errorf("Parsing example method failed: %v did not match %v", ParseMethodLine(testMethodDec), parsedMethodDec)
  }
}

func TestParseMethodSignature(t *testing.T) {
  testMethodSig := "public int getX();"
  parsedMethodSig := map[string]interface{}{
    "name": "getX",
    "declarationType": "methodSignature",
    "returnType": "int",
    "accessModifiers": []string{"public"},
    "contentLines": []string{},
  }

  if !reflect.DeepEqual(ParseMethodSignatureLine(testMethodSig), parsedMethodSig) {
    t.Errorf("Parsing example method signature failed: result %v and example %v do not match", ParseMethodSignatureLine(testMethodSig), parsedMethodSig)
  }
}

func TestParseMethodVariable(t *testing.T) {
  testMemVar := "public int value;"
  parsedMemVar := map[string]interface{}{
    "name": "value",
    "declarationType": "memberVariable",
    "returnType": "int",
    "accessModifiers": []string{"public"},
  }

  if !reflect.DeepEqual(ParseMemberVariableLine(testMemVar), parsedMemVar) {
    t.Errorf("Parsing example member variable failed: result %v and example %v do not match", ParseMemberVariableLine(testMemVar), parsedMemVar)
  }
}

func TestSimpleClass(t *testing.T) {
  f, err := ioutil.ReadFile("../testsnippets/simple.java")
  if err != nil {
    t.Fatalf("Opening source file failed with err: %v", err)
  }

  o, err := ioutil.ReadFile("../testsnippets/simple.go")
  if err != nil {
    t.Fatalf("Opening example file failed with err: %v", err)
  }

  if ParseFileContents(string(f)) != string(o) {
    t.Errorf("Parsing simple class failed: parsed %v and example %v did not match", ParseFileContents(string(f)), string(o))
  }
}

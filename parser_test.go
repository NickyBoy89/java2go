package main

import (
  "io/ioutil"
  "testing"
  "reflect"
)

func TestParseDeclaration(t *testing.T) {
  testClassDec := "public protected static class TestClass {"
  parsedClassDec := JavaClass{
    Name: "TestClass",
    DeclarationType: "class",
    AccessModifiers: []string{"public", "protected", "static"},
  }
  testMethodDec := "public int GetValue {"
  parsedMethodDec := JavaMethodItem{
    Name: "GetValue",
    DeclarationType: "method",
    ReturnType: "int",
    AccessModifiers: []string{"public"},
  }

  if !reflect.DeepEqual(ParseDeclaration(testClassDec), parsedClassDec) {
    t.Errorf("Parsing example class failed: %v did not match %v", ParseDeclaration(testClassDec), parsedClassDec)
  }

  if !reflect.DeepEqual(ParseDeclaration(testMethodDec), parsedMethodDec) {
    t.Errorf("Parsing example method failed: %v did not match %v", ParseDeclaration(testMethodDec), parsedMethodDec)
  }
}

func TestParseMethodSignature(t *testing.T) {
  testMethodSig := "public int getX();"
  parsedMethodSig := JavaMethodItem{
    Name: "getX",
    DeclarationType: "methodSignature",
    ReturnType: "int",
    AccessModifiers: []string{"public"},
  }

  if !reflect.DeepEqual(ParseMethodSignature(testMethodSig), parsedMethodSig) {
    t.Errorf("Parsing example method signature failed: result %v and example %v do not match", ParseMethodSignature(testMethodSig), parsedMethodSig)
  }
}

func TestParseMethodVariable(t *testing.T) {
  testMemVar := "public int value;"
  parsedMemVar := JavaMethodItem{
    Name: "value",
    DeclarationType: "memberVariable",
    ReturnType: "int",
    AccessModifiers: []string{"public"},
  }

  if !reflect.DeepEqual(ParseMemberVariable(testMemVar), parsedMemVar) {
    t.Errorf("Parsing example member variable failed: result %v and example %v do not match", ParseMemberVariable(testMemVar), parsedMemVar)
  }
}

func TestSimpleClass(t *testing.T) {
  f, err := ioutil.ReadFile("testsnippets/simple.java")
  if err != nil {
    t.Fatalf("Opening source file failed with err: %v", err)
  }

  o, err := ioutil.ReadFile("testsnippets/simple.go")
  if err != nil {
    t.Fatalf("Opening example file failed with err: %v", err)
  }

  if ParseFileContents(string(f)) != string(o) {
    t.Errorf("Parsing simple class failed: parsed %v and example %v did not match", ParseFileContents(string(f)), string(o))
  }
}

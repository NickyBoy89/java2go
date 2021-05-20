package parsing

import (
  "fmt"

  "gitlab.nicholasnovak.io/snapdragon/java2go/codeparser"
)

type ParsedClasses interface {
  // Returns the type of the parsed class structure
  // Ex: "class" for a class, "interface" for an interface
  GetType() string
  MethodNames() []string
}

// Represents a Java class
type ParsedClass struct {
  Name string
  Modifiers []string
  Implements []string
  Extends string // A class can only extend one class
  ClassVariables []ParsedVariable
  Methods []ParsedMethod
  NestedClasses []ParsedClasses
  StaticBlocks []string
}

func (c ParsedClass) GetType() string {
  return "class"
}

func (c ParsedClass) MethodNames() []string {
  names := []string{}
  // Get the names of all the methods in the class
  for _, method := range c.Methods {
    names = append(names, method.Name)
  }
  // Get the names of all the methods in the nested classes
  for _, nested := range c.NestedClasses {
    names = append(names, nested.MethodNames()...)
  }
  return names
}

// Represents a Java interface
type ParsedInterface struct {
  Name string
  Modifiers []string
  Methods []ParsedMethod
  StaticFields []ParsedVariable
  DefaultMethods []ParsedMethod
  NestedClasses []ParsedClasses
}

func (i ParsedInterface) GetType() string {
  return "interface"
}

func (i ParsedInterface) MethodNames() []string {
  names := []string{}
  // Get the names of all the methods in the class
  for _, method := range i.Methods {
    names = append(names, method.Name)
  }
  // Get the names of all the methods in the nested classes
  for _, nested := range i.NestedClasses {
    names = append(names, nested.MethodNames()...)
  }
  return names
}

// Represents a Java Enum
type ParsedEnum struct {
  Name string
  Modifiers []string
  Implements []string
  ClassVariables []ParsedVariable
  Methods []ParsedMethod
  EnumFields []EnumField
  NestedClasses []ParsedClasses
  StaticBlocks []string
}

func (e ParsedEnum) GetType() string {
  return "enum"
}

func (e ParsedEnum) MethodNames() []string {
  names := []string{}
  // Get the names of all the methods in the class
  for _, method := range e.Methods {
    names = append(names, method.Name)
  }
  // Get the names of all the methods in the nested classes
  for _, nested := range e.NestedClasses {
    names = append(names, nested.MethodNames()...)
  }
  return names
}

type EnumField struct {
  Name string
  Parameters []ParsedVariable
}

type ParsedVariable struct {
  Name string
  Modifiers []string
  Annotation string
  DataType string
  InitialValue string
}

func (v ParsedVariable) String() string {
  return fmt.Sprintf("Name: %v DataType: %v Modifiers: %v InitialValue: %v", v.Name, v.DataType, v.Modifiers, v.InitialValue)
}

func (c ParsedClass) String() string {
  return fmt.Sprintf("Name: %v Modifiers: %v Variables: %v Methods: %v", c.Name, c.Modifiers, c.ClassVariables, c.Methods)
}

type ParsedMethod struct {
  Name string
  Modifiers []string
  Annotation string
  Parameters []ParsedVariable
  ReturnType string
  Body []codeparser.LineTyper
}

package parsing

import (
  "fmt"
)

type ParsedClasses interface {
  // Returns the type of the parsed class structure
  // Ex: "class" for a class, "interface" for an interface
  GetType() string
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

// Represents a Java interface
type ParsedInterface struct {
  Name string
  Modifiers []string
  Methods []ParsedMethod
  DefaultMethods []ParsedMethod
  NestedClasses []ParsedClasses
}

func (i ParsedInterface) GetType() string {
  return "interface"
}

// Represents a Java Enum
type ParsedEnum struct {
  Name string
  Modifiers []string
  ClassVariables []ParsedVariable
  Methods []ParsedMethod
  EnumFields []EnumField
  NestedClasses []ParsedClasses
}

func (e ParsedEnum) GetType() string {
  return "enum"
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
  Body string
}

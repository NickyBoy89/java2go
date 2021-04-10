package parsing

import (
  "fmt"
)

// Represents a Java class
type ParsedClass struct {
  Name string
  Modifiers []string
  ClassVariables []ParsedVariable
  Methods []ParsedMethod
  Classes []ParsedClass
}

// Represents a Java interface
type ParsedInterface struct {
  Name string
  Modifiers []string
  Methods []ParsedMethod
  DefaultMethods []ParsedMethod
}

// Represents a Java Enum
type ParsedEnum struct {
  Name string
  Modifiers []string
  ClassVariables []ParsedVariable
  Methods []ParsedMethod
  EnumFields []EnumField
}

type EnumField struct {
  Name string
  Parameters []ParsedVariable
}

type ParsedVariable struct {
  Name string
  Modifiers []string
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
  Parameters []ParsedVariable
  ReturnType string
  Body string
}

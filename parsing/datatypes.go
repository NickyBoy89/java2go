package parsing

import (
  "fmt"
  "unicode"

  "gitlab.nicholasnovak.io/snapdragon/java2go/codeparser"
  "gitlab.nicholasnovak.io/snapdragon/java2go/parsetools"
)

type ParsedClasses interface {
  // Returns the type of the parsed class structure
  // Ex: "class" for a class, "interface" for an interface
  GetType() string
  MethodContext() map[string][]string
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
  StaticBlocks [][]codeparser.LineTyper
}

func (c ParsedClass) GetType() string {
  return "class"
}

func (c ParsedClass) MethodContext() map[string][]string {
  methods := make(map[string][]string)
  // Get the names of all the methods in the class
  for _, method := range c.Methods {
    if IsPublic(method.Modifiers) {
      methods[ToPublic(method.Name)] = method.ParameterTypes()
    } else {
      methods[ToPrivate(method.Name)] = method.ParameterTypes()
    }
  }
  // Get the names of all the methods in the nested classes
  for _, nested := range c.NestedClasses {
    for nestedMethodName, nestedMethodParams := range nested.MethodContext() {
      methods[nestedMethodName] = nestedMethodParams
    }
  }
  return methods
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

func (i ParsedInterface) MethodContext() map[string][]string {
  methods := make(map[string][]string)
  // Get the names of all the methods in the class
  for _, method := range i.Methods {
    methods[method.Name] = method.ParameterTypes()
  }
  // Get the names of all the methods in the nested classes
  for _, nested := range i.NestedClasses {
    for nestedMethodName, nestedMethodParams := range nested.MethodContext() {
      methods[nestedMethodName] = nestedMethodParams
    }
  }
  return methods
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
  StaticBlocks [][]codeparser.LineTyper
}

func (e ParsedEnum) GetType() string {
  return "enum"
}

func (e ParsedEnum) MethodContext() map[string][]string {
  methods := make(map[string][]string)
  // Get the names of all the methods in the class
  for _, method := range e.Methods {
    methods[method.Name] = method.ParameterTypes()
  }

  // Get the names of all the methods in the nested classes
  for _, nested := range e.NestedClasses {
    for nestedMethodName, nestedMethodParams := range nested.MethodContext() {
      methods[nestedMethodName] = nestedMethodParams
    }
  }
  return methods
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

func (pm ParsedMethod) ParameterTypes() []string {
  names := []string{}
  for _, param := range pm.Parameters {
    names = append(names, param.DataType)
  }
  return names
}

func ToPublic(name string) string {
	return string(unicode.ToUpper(rune(name[0]))) + name[1:]
}

func ToPrivate(name string) string {
	return string(unicode.ToLower(rune(name[0]))) + name[1:]
}

// Tests if an object is public, given its modifiers
func IsPublic(modifiers []string) bool {
	if parsetools.Contains("public", modifiers) || parsetools.Contains("protected", modifiers) {
		return true
	}
	return false
}

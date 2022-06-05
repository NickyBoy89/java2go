package symbol

import "fmt"

// Definition represents the name and type of a single symbol
type Definition struct {
	// The original Java name
	originalName string
	// The display name of the definition, may be different from the original name
	name string
	// Original Java type of the object
	originalType string
	// Display type of the object
	typ string

	// If the definition is a constructor
	// This is used so that the definition handles its special naming and
	// type rules correctly
	constructor bool
	// If the object is a function, it has parameters
	parameters []*Definition
	// Children of the declaration, if the declaration is a scope
	children []*Definition
}

// The display name of the definition
func (d Definition) Name() string {
	return d.name
}

// The display type of the definition
func (d Definition) Type() string {
	return d.typ
}

// The original Java type of the definition
func (d Definition) OriginalType() string {
	return d.originalType
}

func (d Definition) String() string {
	if d.originalName != d.name {
		return fmt.Sprintf("Name: %s (Was %s) Type: %s", d.name, d.originalName, d.typ)

	}
	return fmt.Sprintf("Name: %s Type: %s", d.name, d.typ)
}

// Rename changes the display name of a definition
func (d *Definition) Rename(name string) {
	d.name = name
}

// ParameterByName returns a parameter's definition, given its original name
func (d *Definition) ParameterByName(name string) *Definition {
	for _, param := range d.parameters {
		if param.originalName == name {
			return param
		}
	}
	return nil
}

// OriginalParameterTypes returns a list of the original types for all the parameters
func (d *Definition) OriginalParameterTypes() []string {
	names := make([]string, len(d.parameters))
	for ind, param := range d.parameters {
		names[ind] = param.originalType
	}
	return names
}

// FindVariable searches a definition's immediate children and parameters
// to try and find a given variable by its original name
func (d *Definition) FindVariable(name string) *Definition {
	for _, param := range d.parameters {
		if param.originalName == name {
			return param
		}
	}
	for _, child := range d.children {
		if child.originalName == name {
			return child
		}
	}
	return nil
}

// ExistsIn reports whether this definition conflicts with an already existing
// definition in the given scope
func (d *Definition) MethodExistsIn(scope Scope) bool {
	parameterTypes := []string{}
	for _, param := range d.parameters {
		parameterTypes = append(parameterTypes, param.originalType)
	}
	return scope.FindMethodByName(d.Name(), parameterTypes) != nil
}

// FieldExistsInPackage searches for a given field in all the classes in a package
// This is useful for finding duplicate global variables, as an optional class
// name can be provided to skip over, meaning that it will not find any duplicates in the same class
func (d *Definition) FieldExistsInPackage(packageScope *PackageScope, skippedClassName string) bool {
	for _, classFile := range packageScope.files {
		if classFile.BaseClass.Class.name == skippedClassName {
			continue
		}
		if classFile.BaseClass.FindFieldByDisplayName(d.Name()) != nil {
			return true
		}
	}
	return false
}

func (d Definition) isEmpty() bool {
	return d.originalName == "" && len(d.children) == 0
}

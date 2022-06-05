package symbol

// Definition represents the name and type of a single symbol
type Definition struct {
	// The original Java name
	OriginalName string
	// The display name of the definition, may be different from the original name
	Name string
	// Original Java type of the object
	OriginalType string
	// Display type of the object
	Type string

	// If the definition is a constructor
	// This is used so that the definition handles its special naming and
	// type rules correctly
	Constructor bool
	// If the object is a function, it has parameters
	Parameters []*Definition
	// Children of the declaration, if the declaration is a scope
	Children []*Definition
}

// Rename changes the display name of a definition
func (d *Definition) Rename(name string) {
	d.Name = name
}

// ParameterByName returns a parameter's definition, given its original name
func (d *Definition) ParameterByName(name string) *Definition {
	for _, param := range d.Parameters {
		if param.OriginalName == name {
			return param
		}
	}
	return nil
}

// OriginalParameterTypes returns a list of the original types for all the parameters
func (d *Definition) OriginalParameterTypes() []string {
	names := make([]string, len(d.Parameters))
	for ind, param := range d.Parameters {
		names[ind] = param.OriginalType
	}
	return names
}

// FindVariable searches a definition's immediate children and parameters
// to try and find a given variable by its original name
func (d *Definition) FindVariable(name string) *Definition {
	for _, param := range d.Parameters {
		if param.OriginalName == name {
			return param
		}
	}
	for _, child := range d.Children {
		if child.OriginalName == name {
			return child
		}
	}
	return nil
}

// ExistsIn reports whether this definition conflicts with an already existing
// definition in the given scope
func (d *Definition) MethodExistsIn(scope Scope) bool {
	parameterTypes := []string{}
	for _, param := range d.Parameters {
		parameterTypes = append(parameterTypes, param.OriginalType)
	}
	return scope.FindMethodByName(d.Name, parameterTypes) != nil
}

// FieldExistsInPackage searches for a given field in all the classes in a package
// This is useful for finding duplicate global variables, as an optional class
// name can be provided to skip over, meaning that it will not find any duplicates in the same class
func (d *Definition) FieldExistsInPackage(packageScope *PackageScope, skippedClassName string) bool {
	for _, classFile := range packageScope.files {
		if classFile.BaseClass.Class.Name == skippedClassName {
			continue
		}
		if classFile.BaseClass.FindFieldByDisplayName(d.Name) != nil {
			return true
		}
	}
	return false
}

func (d Definition) IsEmpty() bool {
	return d.OriginalName == "" && len(d.Children) == 0
}

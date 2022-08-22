package symbol

// ClassScope represents a single defined class, and the declarations in it
type ClassScope struct {
	// The definition for the class defined within the class
	Class *Definition
	// Every class that is nested within the base class
	Subclasses []*ClassScope
	// Any normal and static fields associated with the class
	Fields []*Definition
	// Methods and constructors
	Methods []*Definition
}

type ClassMethodFinder ClassScope

func (cs *ClassScope) FindMethod() Finder {
	cm := ClassMethodFinder(*cs)
	return &cm
}

func (cm *ClassMethodFinder) By(criteria func(d *Definition) bool) []*Definition {
	results := []*Definition{}
	for _, method := range cm.Methods {
		if criteria(method) {
			results = append(results, method)
		}
	}
	return results
}

func (cm *ClassMethodFinder) ByName(name string) []*Definition {
	return cm.By(func(d *Definition) bool {
		return d.Name == name
	})
}

func (cm *ClassMethodFinder) ByOriginalName(originalName string) []*Definition {
	return cm.By(func(d *Definition) bool {
		return d.Name == originalName
	})
}

// FindMethodByDisplayName searches for a given method by its display name
// If some ignored parameter types are specified as non-nil, it will skip over
// any function that matches these ignored parameter types exactly
func (cs *ClassScope) FindMethodByName(name string, ignoredParameterTypes []string) *Definition {
	return cs.findMethodWithComparison(func(method *Definition) bool { return method.OriginalName == name }, ignoredParameterTypes)
}

// FindMethodByDisplayName searches for a given method by its display name
// If some ignored parameter types are specified as non-nil, it will skip over
// any function that matches these ignored parameter types exactly
func (cs *ClassScope) FindMethodByDisplayName(name string, ignoredParameterTypes []string) *Definition {
	return cs.findMethodWithComparison(func(method *Definition) bool { return method.Name == name }, ignoredParameterTypes)
}

func (cs *ClassScope) findMethodWithComparison(comparison func(method *Definition) bool, ignoredParameterTypes []string) *Definition {
	for _, method := range cs.Methods {
		if comparison(method) {
			// If no parameters were specified to ignore, then return the first match
			if ignoredParameterTypes == nil {
				return method
			} else if len(method.Parameters) != len(ignoredParameterTypes) { // Size of parameters were not equal, instantly not equal
				return method
			}

			// Check the remaining paramters one-by-one
			for index, parameter := range method.Parameters {
				if parameter.OriginalType != ignoredParameterTypes[index] {
					return method
				}
			}
		}
	}

	// Not found
	return nil
}

// FindClass searches through a class file and returns the definition for the
// found class, or nil if none was found
func (cs *ClassScope) FindClass(name string) *Definition {
	if cs.Class.OriginalName == name {
		return cs.Class
	}
	for _, subclass := range cs.Subclasses {
		class := subclass.FindClass(name)
		if class != nil {
			return class
		}
	}
	return nil
}

// FindFieldByName searches for a field by its original name, and returns its definition
// or nil if none was found
func (cs *ClassScope) FindFieldByName(name string) *Definition {
	for _, field := range cs.Fields {
		if field.OriginalName == name {
			return field
		}
	}
	return nil
}

func (cs *ClassScope) FindFieldByDisplayName(name string) *Definition {
	for _, field := range cs.Fields {
		if field.Name == name {
			return field
		}
	}
	return nil
}

package goparser

type ClassContext struct {
	Name string
	Package string
	Methods map[string][]string
	ClassVariables []string
}

func (c ClassContext) ContainsMethod(target string) bool {
	// Look through all of the methods in the class's context
	for name := range c.Methods {
		if ToPublic(name) == target {
			return true
		}
	}
	// Also look through a list of other generic functions
	for _, lookup := range functionTable {
		if lookup == target {
			return true
		}
	}
	return false
}

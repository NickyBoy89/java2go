package goparser

type ClassContext struct {
	Name string
	Methods []string
	ClassVariables []string
}

func (c ClassContext) ContainsMethod(target string) bool {
	for _, method := range c.Methods {
		if ToPublic(method) == target {
			return true
		}
	}
	return false
}

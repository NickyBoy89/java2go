package codeparser

type LineTyper interface {
	GetName() string
}

type LineType struct {
	Name  string
	Words map[string]interface{}
}

func (t LineType) GetName() string {
	return t.Name
}

type LineBlock struct {
	Name  string
	Words map[string]interface{}
	Lines []LineTyper
}

func (b LineBlock) GetName() string {
	return b.Name
}

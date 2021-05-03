package codeparser

type LineTyper interface {
	Name() string
}

type LineType struct {
	name  string
	Words map[string]interface{}
}

func (t LineType) Name() string {
	return t.name
}

type LineBlock struct {
	name  string
	Words []string
	Lines []LineTyper
}

func (b LineBlock) Name() string {
	return b.name
}

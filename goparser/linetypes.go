package goparser

type LineTyper interface {
  GetName() string
}

type LineType struct {
  Name string
  Words []string
}

func (t LineType) GetName() string {
  return t.Name
}

type LineBlock struct {
  Name string
  Words []string
  Lines []LineTyper
}

func (b LineBlock) GetName() string {
  return b.Name
}

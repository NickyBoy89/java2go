package main

type Test struct {
	value int
}

func NewTest(val int) *Test {
	ts := new(Test)
	ts.value = val
	return ts
}

func (ts *Test) GetValue() int {
	return rs.value
}

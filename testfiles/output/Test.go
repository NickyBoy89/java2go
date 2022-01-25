package main

type Test struct {
	value int
}

func NewTest(val int) *Test {
	tt := new(Test)
	tt.value = val
	return tt
}

func (tt *Test) GetValue() int {
	return tt.value
}

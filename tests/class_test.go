package tests

import "testing"

func TestCallOwnMethod(t *testing.T) {
	javaInput := `class Foo {
void hello() {
	this.bar();
}

void bar() {}
}`
	goOutput := `package main

type Foo struct {
}

func (this *Foo) Hello() {
	this.bar()
}
func (this *Foo) Bar() {
}
`

	ComparePrograms(javaInput, goOutput, t)
}

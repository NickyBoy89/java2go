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

type foo struct {
}

func (this *foo) hello() {
	this.bar()
}
func (this *foo) bar() {
}
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestInstanceVariables(t *testing.T) {
	javaInput := `class Foo {
	int x;
}`
	goOutput := `package main

type foo struct {
	x int32
}
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestFieldAccess(t *testing.T) {
	javaInput := `class Foo {
	int x;

	void bar() {
		this.x;
	}
}`
	goOutput := `package main

type foo struct {
	x int32
}

func (this *foo) bar() {
	this.x;
}
`

	ComparePrograms(javaInput, goOutput, t)
}

package tests

import "testing"

/*
func Test(t *testing.T) {
	javaInput := ``
	goOutput := ``

	ComparePrograms(javaInput, goOutput, t)
}
*/

func TestVariableDeclaration(t *testing.T) {
	javaInput := `class Foo {
	void bar() {
		int x = 1;
	}
}`
	goOutput := `package main

type Foo struct {
}

func (this *Foo) Bar() {
	var x int32 = 1
}
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestPostIncStmt(t *testing.T) {
	javaInput := `class Foo {
	void bar() {
		int x = 1;
		x++;
	}
}`
	goOutput := `package main

type Foo struct {
}

func (this *Foo) Bar() {
	var x int32 = 1
	PostUpdate(x)
}
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestMathAssign(t *testing.T) {
	javaInput := `class Foo {
	void bar() {
		int x = 1 + 2;
	}
	}`
	goOutput := `package main

type Foo struct {
}

func (this *Foo) Bar() {
	var x int32 = 1 + 2
}
`

	ComparePrograms(javaInput, goOutput, t)
}

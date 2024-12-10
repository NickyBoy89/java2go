package tests

import (
	"testing"
)

func TestParseClass(t *testing.T) {
	javaInput := "class Foo {}"
	goOutput := `package main

type Foo struct {
}
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestMethod(t *testing.T) {
	javaInput := `class Foo {

void Hello() {}

}`
	goOutput := `package main

type Foo struct {
}

func (this *Foo) Hello() {
}
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestStaticMethod(t *testing.T) {
	javaInput := `class Foo {

static void Hello() {}

}
`
	goOutput := `package main

type Foo struct {
}

func Hello() {}
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestClassFields(t *testing.T) {
	javaInput := `class Foo {
	int x;
}
`
	goOutput := `package main

type Foo struct {
	x int32
}
`

	ComparePrograms(javaInput, goOutput, t)
}

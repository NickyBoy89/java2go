package tests

import "testing"

func TestSimpleGenericClass(t *testing.T) {
	javaInput := `class Foo<T> {}`
	goOutput := `package main

type foo[T any] struct {
}
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestSimpleGenericMethod(t *testing.T) {
	javaInput := `class Foo {
	<T> void bar() {}
}`
	goOutput := `package main

type foo struct {
}

func (this *foo) bar[T any]() {
}
`

	ComparePrograms(javaInput, goOutput, t)
}

package tests

import "testing"

func TestSimpleGenericClass(t *testing.T) {
	javaInput := `class Foo<T> {}`
	goOutput := `package main

type Foo[T any] struct {
}
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestSimpleGenericMethod(t *testing.T) {
	javaInput := `class Foo {
	<T> void bar() {}
}`
	goOutput := `package main

type Foo struct {
}

func (this *Foo) Bar[T any]() {
}
`

	ComparePrograms(javaInput, goOutput, t)
}

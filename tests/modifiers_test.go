package tests

import "testing"

func TestDefaultModifier(t *testing.T) {
	javaInput := `class Foo {}`
	goOutput := `package main

type Foo struct {
}
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestPublicModifier(t *testing.T) {
	javaInput := `public class Foo {}`
	goOutput := `package main

type Foo struct {
}
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestPrivateModifier(t *testing.T) {
	javaInput := `private class Foo {}`
	goOutput := `package main

type foo struct {
}
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestProtectedModifier(t *testing.T) {
	javaInput := `public class Foo {}`
	goOutput := `package main

type Foo struct {
}
`

	ComparePrograms(javaInput, goOutput, t)
}

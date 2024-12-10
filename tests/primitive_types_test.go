package tests

import "testing"

func TestPrimitiveIntegralTypes(t *testing.T) {
	javaInput := `class Foo {
	byte a;
	short b;
	int c;
	long d;
	char e;
}`

	goOutput := `package main

type Foo struct {
	a	byte
	b	int16
	c	int32
	d	int64
	e	rune
}
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestFloatingPointTypes(t *testing.T) {
	javaInput := `class Foo {
	float a;
	double b;
}`

	goOutput := `package main

type Foo struct {
	a	float32
	b	float64
}
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestBooleanType(t *testing.T) {
	javaInput := `class Foo {
	bool a;
}`

	goOutput := `package main

type Foo struct {
	a bool
}
`

	ComparePrograms(javaInput, goOutput, t)
}

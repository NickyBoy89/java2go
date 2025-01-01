package tests

import "testing"

// $.decimal_integer_literal,
// $.hex_integer_literal,
// $.octal_integer_literal,
// $.binary_integer_literal,
// $.decimal_floating_point_literal,
// $.hex_floating_point_literal,
// $.true,
// $.false,
// $.character_literal,
// $.string_literal,
// $.null_literal,

func TestDecimalLiteral(t *testing.T) {
	javaInput := `int i = 1;
long i = 1l;
long i = 1L;`
	goOutput := `package main

var i int32 = 1
var i int64 = int64(1)
var i int64 = int64(1)
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestHexLiteral(t *testing.T) {
	javaInput := `int i = 0x01;
int i = 0X01;
long i = 0x01l;
long i = 0X01L;`
	goOutput := `package main

var i int32 = 0x01
var i int32 = 0X01
var i int64 = int64(0x01)
var i int64 = int64(0X01)
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestOctalLiteral(t *testing.T) {
	javaInput := `int i = 01;
int i = 0o1;
int i = 0O1;
long i = 01l;`
	goOutput := `package main

var i int32 = 01
var i int32 = 0o1
var i int32 = 0O1
var i int64 = int64(01)
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestBinaryLiteral(t *testing.T) {
	javaInput := `int i = 0b01;
int i = 0B01;
long i = 0b01l;`
	goOutput := `package main

var i int32 = 0b01
var i int32 = 0B01
var i int64 = int64(0b01)
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestFloatingPointLiteral(t *testing.T) {
	javaInput := `float i = 1.0;
float i = 1.0f;
float i = 1.0F;
double i = 1.0d;
double i = 1.0D;`
	goOutput := `package main

var i float32 = 1.0
var i float32 = float32(1.0)
var i float32 = float32(1.0)
var i float64 = float64(1.0)
var i float64 = float64(1.0)
`

	ComparePrograms(javaInput, goOutput, t)
}

// TODO: Needs a lot more testing
func TestHexFloatingPointLiteral(t *testing.T) {
	javaInput := "float i = 0x01f;"
	goOutput := `package main

var i float32 = float32(0x01)
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestTrue(t *testing.T) {
	javaInput := "boolean i = true;"
	goOutput := `package main

var i bool = true
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestFalse(t *testing.T) {
	javaInput := "boolean i = false;"
	goOutput := `package main

var i bool = false
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestCharacterLiteral(t *testing.T) {
	javaInput := "char i = 'a';"
	goOutput := `package main

var i rune = 'a'
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestStringLiteral(t *testing.T) {
	javaInput := "String i = \"hello\";"
	goOutput := `package main

var i string = "hello"
`

	ComparePrograms(javaInput, goOutput, t)
}

func TestNullLiteral(t *testing.T) {
	javaInput := "Object i = null;"
	goOutput := `package main

var i *Object = nil`

	ComparePrograms(javaInput, goOutput, t)
}

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
int i = 1l;
int i = 1L;`
	goOutput := ""

	ComparePrograms(javaInput, goOutput, t)
}

func TestHexLiteral(t *testing.T) {
	javaInput := `int i = 0x01;
int i = 0X01;
int i = 0x01l;
int i = 0X01L;`
	goOutput := ""

	ComparePrograms(javaInput, goOutput, t)
}

func TestOctalLiteral(t *testing.T) {
	javaInput := `int i = 01;
int i = 0o1;
int i = 0O1;
int i = 01l;`
	goOutput := ""

	ComparePrograms(javaInput, goOutput, t)
}

func TestBinaryLiteral(t *testing.T) {
	javaInput := `int i = 0b01;
int i = 0B01;
int i = 0b01l;`
	goOutput := ""

	ComparePrograms(javaInput, goOutput, t)
}

func TestFloatingPointLiteral(t *testing.T) {
	javaInput := `float i = 1.0f;
float i = 1.0F;
double i = 1.0d;
double i = 1.0D;`
	goOutput := ""

	ComparePrograms(javaInput, goOutput, t)
}

// TODO: Needs a lot more testing
func TestHexFloatingPointLiteral(t *testing.T) {
	javaInput := "float i = 0x01f;"
	goOutput := ""

	ComparePrograms(javaInput, goOutput, t)
}

func TestTrue(t *testing.T) {
	javaInput := "boolean i = true;"
	goOutput := ""

	ComparePrograms(javaInput, goOutput, t)
}

func TestFalse(t *testing.T) {
	javaInput := "boolean i = false;"
	goOutput := ""

	ComparePrograms(javaInput, goOutput, t)
}

func TestCharacterLiteral(t *testing.T) {
	javaInput := "char i = 'a';"
	goOutput := ""

	ComparePrograms(javaInput, goOutput, t)
}

func TestStringLiteral(t *testing.T) {
	javaInput := "String i = \"hello\";"
	goOutput := ""

	ComparePrograms(javaInput, goOutput, t)
}

func TestNullLiteral(t *testing.T) {
	javaInput := "Object i = null;"
	goOutput := ""

	ComparePrograms(javaInput, goOutput, t)
}

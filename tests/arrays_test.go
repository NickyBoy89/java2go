package tests

import "testing"

func TestArrayCreation(t *testing.T) {
	javaInput := `class Foo {
	void bar() {
		int[] temp = new int[0];
	}
}`
	goOutput := ``

	ComparePrograms(javaInput, goOutput, t)
}

func TestMultiDimArray(t *testing.T) {
	javaInput := `class Foo {
	void bar() {
		int[][] temp = new int[0][0];
	}
}`
	goOutput := ``

	ComparePrograms(javaInput, goOutput, t)
}

func TestArrayInitializer(t *testing.T) {
	javaInput := `class Foo {
	void bar() {
		int[] temp = {0, 1, 2, 3};
	}
}`
	goOutput := ``

	ComparePrograms(javaInput, goOutput, t)
}

func TestMultiDimArrayInit(t *testing.T) {
	javaInput := `class Foo {
	void bar() {
		int ia[][] = { {1, 2}, null };
	}
}`
	goOutput := ``

	ComparePrograms(javaInput, goOutput, t)
}

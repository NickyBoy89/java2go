package tests

import "testing"

func TestPrintHello(t *testing.T) {
	javaSource := `class Foo {
	public static void main(String[] args) {
		System.out.println("Hello World!");
	}
}`
	compareProgramOutputs(javaSource, t)
}

type Test struct {
  Value int
  Value2 int
  value3 int
}

func NewTest(value int) *Test {
  test := new(Test)

  test.Value = value
  test.Value2 = value + 1
  test.value3 = value + 2

  return test
}

func (t Test) GetValue(specified int) int {
  if (specified == 1) {
    return this.value
  } else if (specified == 2) {
    return this.value2
  } else {
    return this.value3
  }
}

func Hello() string {
  return "Hello World!"
}

package testsnippets

type TestClass struct {
  value int
  testClassType string
}

func (tc TestClass) GetValue() int {
  return tc.value
}

func (tc TestClass) GetType() string {
  return tc.testClassType
}

package bytelevel

import (
  "testing"
)

func TestExtractSimpleMethod(t *testing.T) {
  methodString := `public static int divideByTwo(int num1, boolean fast) {
      if (fast) {
        return num1 << 1;
      } else {
        return num1 / 2;
      }
    }`

  WalkParse(methodString)
}

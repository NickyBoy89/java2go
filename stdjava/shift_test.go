package stdjava

import "testing"

// Examples from the JavaScript reference at https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Unsigned_right_shift

func TestRightShiftPositive(t *testing.T) {
	if UnsignedRightShift(9, 2) != 2 {
		t.Errorf("Shifted 9 >>> 2. Expected 2 but got %d", UnsignedRightShift(9, 2))
	}
}

func TestRightShiftNegative(t *testing.T) {
	if UnsignedRightShift(-9, 2) != 1073741821 {
		t.Errorf("Shifted -9 >>> 2. Expected 1073741821 but got %d", UnsignedRightShift(-9, 2))
	}
}

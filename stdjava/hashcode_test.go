package stdjava

import "testing"

func TestBasicStrings(t *testing.T) {
	in := "Hello"
	if HashCode(in) != 69609650 {
		t.Errorf("Expected the hash to be 69609650. Got %d", HashCode(in))
	}
}

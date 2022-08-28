package stdjava

import "testing"

func TestSimpleGrid(t *testing.T) {
	arr := MultiDimensionArray([]string{}, 3, 5)
	if len(arr) != 3 {
		t.Errorf("Got %d rows, expected %d", len(arr), 3)
	}
	if len(arr[0]) != 5 {
		t.Errorf("Got %d cols, expected %d", len(arr[0]), 5)
	}
}

func TestAdvancedGrid(t *testing.T) {
	arr := MultiDimensionArray3([][]int{}, 1, 2, 3)
	if len(arr) != 1 {
		t.Errorf("Got %d rows, expected %d", len(arr), 1)
	}
	if len(arr[0]) != 2 {
		t.Errorf("Got %d cols, expected %d", len(arr[0]), 2)
	}
	if len(arr[0][0]) != 3 {
		t.Errorf("Got %d third, expected %d", len(arr), 3)
	}
}

package stdjava

import (
	"math"

	"golang.org/x/exp/constraints"
)

// Ternary represents Java's ternary operator (condition ? result1 : result2)
func Ternary[T any](condition bool, result1, result2 T) T {
	if condition {
		return result1
	}
	return result2
}

// UnsignedRightShift is an implementation of Java's unsigned right shift
// operation where a number is shifted over the number of times specified, but
// the topmost bits are always filled in with zeroes
func UnsignedRightShift[V, A constraints.Integer](value V, amount A) V {
	return V(uint32(value) >> amount)
}

// UnsignedRightShiftAssignment represents a right-shift assignment (`>>>=`)
// where a value is assigned the result of an unsigned right shift
func UnsignedRightShiftAssignment[A any, V constraints.Integer](assignTo *A, value V) {
	// TODO: Fix this conversion hack, and change the function to take proper values
	*assignTo = interface{}(UnsignedRightShift(value, 2)).(A)
}

// HashCode is an implementation of Java's String `hashCode` method
func HashCode(s string) int {
	var total int
	n := len(s)
	for ind, char := range s {
		total += int(char) * int(math.Pow(float64(31), float64(n - (ind+1))))
	}
	return total
}

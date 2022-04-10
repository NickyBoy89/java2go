package stdjava

// Option formally represents a value that can be nil
type Optional[T any] struct {
	value *T
}

// Some returns true if a value is present
func (o Optional) Some() bool {
	return o.value != nil
}

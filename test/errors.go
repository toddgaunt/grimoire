package test

// ErrorTextEqual compares two error values and returns true if both are nil or their error messages are equal.
func ErrorTextEqual(a, b error) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil && b != nil {
		return false
	}
	if a != nil && b == nil {
		return false
	}
	return a.Error() == b.Error()
}

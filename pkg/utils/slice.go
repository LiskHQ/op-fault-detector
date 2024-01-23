// Package utils implements various utilities that assist in the application development.
package utils

// Contains checks if the specified element exists in the given slice. The supplied element must be a comparable type.
func Contains[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

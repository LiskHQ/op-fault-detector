package utils

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestContains_String(t *testing.T) {
	testCases := []struct {
		name  string
		slice []string
		elem  string
		want  bool
	}{
		{
			name:  "should return true when the specified string contained in the provided slice",
			slice: []string{"a", "b", "c"},
			elem:  "a",
			want:  true,
		},
		{
			name:  "should return false when the specified string not contained in the provided slice",
			slice: []string{"a", "b", "c"},
			elem:  "z",
			want:  false,
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Contains(tc.slice, tc.elem)
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestContains_Int(t *testing.T) {
	testCases := []struct {
		name  string
		slice []int
		elem  int
		want  bool
	}{
		{
			name:  "should return true when the specified int contained in the provided slice",
			slice: []int{1, 2, 3},
			elem:  3,
			want:  true,
		},
		{
			name:  "should return false when the specified int not contained in the provided slice",
			slice: []int{1, 2, 3},
			elem:  99,
			want:  false,
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Contains(tc.slice, tc.elem)
			assert.Equal(t, got, tc.want)
		})
	}
}

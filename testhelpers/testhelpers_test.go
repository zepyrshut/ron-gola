package testhelpers

import (
	"testing"
)

func Test_CheckSlicesEquality(t *testing.T) {
	tests := []struct {
		name   string
		ok     bool
		sliceA []any
		sliceB []any
	}{
		{
			name:   "Integers",
			ok:     true,
			sliceA: []any{2, 3, 1},
			sliceB: []any{1, 2, 3},
		},
		{
			name:   "Strings",
			ok:     true,
			sliceA: []any{"x", "y", "z"},
			sliceB: []any{"z", "y", "x"},
		},
		{
			name:   "Integers 2",
			ok:     true,
			sliceA: []any{1, 2, 3},
			sliceB: []any{1, 2, 3},
		},
		{
			name:   "Different lengths",
			ok:     false,
			sliceA: []any{1, 2, 3},
			sliceB: []any{1, 2, 3, 4},
		},
		{
			name:   "Different lengths 2",
			ok:     false,
			sliceA: []any{1, 2, 3, 4},
			sliceB: []any{1, 2, 3},
		},
		{
			name:   "Different types",
			ok:     false,
			sliceA: []any{1, 2, 3},
			sliceB: []any{"1", "2", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := CheckSlicesEquality(tt.sliceA, tt.sliceB); result != tt.ok {
				t.Errorf("CheckSlicesEquality() = %v, want %v", result, tt.ok)
			}
		})
	}

}

func Test_StringSliceToAnySlice(t *testing.T) {
	expected := []any{"a", "b", "c"}
	actual := StringSliceToAnySlice([]string{"a", "b", "c"})

	if !CheckSlicesEquality(expected, actual) {
		t.Errorf("StringSliceToAnySlice() = %v, want %v", actual, expected)
	}

}

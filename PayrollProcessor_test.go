package main

import (
	"testing"
	// "TaxBracket"
)

// tests for round() method
func TestRounding(t *testing.T) {
	var tests = []struct {
		input float64
		want float64
	} {
		{0.0, 0.0},
		{0.3, 0.0},
		{0.5, 1.0},
		{0.6, 1.0},
		{10.3, 10.0},
		{10.5, 11.0},
		{10.8, 11.0},
	}

	for _, test := range tests {
		if got := round(test.input); got != test.want {
			t.Errorf("FAILED: round(%f) = %v", test, got)
		}
	}
}

package controllers_test

import (
	"testing"

	"github.com/khrees/veilo/controllers"
)

// TestClampInt tests the clampInt helper used for safe query-parameter parsing.
func TestClampInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		min      int
		max      int
		fallback int
		expected int
	}{
		{"valid in range", "50", 1, 100, 10, 50},
		{"below min", "0", 1, 100, 10, 1},
		{"above max", "200", 1, 100, 10, 100},
		{"at min", "1", 1, 100, 10, 1},
		{"at max", "100", 1, 100, 10, 100},
		{"empty string", "", 1, 100, 10, 10},
		{"non-numeric", "abc", 1, 100, 10, 10},
		{"mixed string", "12abc", 1, 100, 10, 10},
		{"negative", "-5", 0, 100, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := controllers.ClampInt(tt.input, tt.min, tt.max, tt.fallback)
			if result != tt.expected {
				t.Errorf("clampInt(%q, %d, %d, %d) = %d; want %d",
					tt.input, tt.min, tt.max, tt.fallback, result, tt.expected)
			}
		})
	}
}


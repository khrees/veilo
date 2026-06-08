package controllers_test

import (
	"testing"

	"github.com/khrees/veilo/controllers"
)

// Test the parseInt helper function (it's not exported, so we test it via the routes package)
func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"0", 0},
		{"123", 123},
		{"999", 999},
		{"abc", 0},
		{"123abc", 123},
		{"abc123", 0},
		{"100200", 100200},
		{"-123", 0}, // negative sign stops parsing
		{"123-456", 123},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := controllers.ParseInt(tt.input)
			if result != tt.expected {
				t.Errorf("parseInt(%q) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}

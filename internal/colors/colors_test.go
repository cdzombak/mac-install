package colors

import (
	"os"
	"testing"
)

func TestColorize(t *testing.T) {
	originalTerm := os.Getenv("TERM")
	originalNoColor := os.Getenv("NO_COLOR")
	
	defer func() {
		_ = os.Setenv("TERM", originalTerm)
		_ = os.Setenv("NO_COLOR", originalNoColor)
	}()

	_ = os.Setenv("TERM", "xterm-256color")
	_ = os.Setenv("NO_COLOR", "")

	result := colorize(Red, "test")
	expected := Red + "test" + Reset
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestColorizeNoColor(t *testing.T) {
	originalTerm := os.Getenv("TERM")
	originalNoColor := os.Getenv("NO_COLOR")
	
	defer func() {
		_ = os.Setenv("TERM", originalTerm)
		_ = os.Setenv("NO_COLOR", originalNoColor)
	}()

	_ = os.Setenv("NO_COLOR", "1")

	result := colorize(Red, "test")
	expected := "test"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestIsColorSupported(t *testing.T) {
	originalTerm := os.Getenv("TERM")
	originalNoColor := os.Getenv("NO_COLOR")
	
	defer func() {
		_ = os.Setenv("TERM", originalTerm)
		_ = os.Setenv("NO_COLOR", originalNoColor)
	}()

	tests := []struct {
		name     string
		term     string
		noColor  string
		expected bool
	}{
		{"color supported", "xterm-256color", "", true},
		{"no color env", "xterm-256color", "1", false},
		{"dumb terminal", "dumb", "", false},
		{"no term", "", "", false},
	}

	for _, test := range tests {
		_ = os.Setenv("TERM", test.term)
		_ = os.Setenv("NO_COLOR", test.noColor)
		
		result := isColorSupported()
		if result != test.expected {
			t.Errorf("Test '%s': expected %v, got %v", test.name, test.expected, result)
		}
	}
}

func TestColorFunctions(t *testing.T) {
	originalTerm := os.Getenv("TERM")
	originalNoColor := os.Getenv("NO_COLOR")
	
	defer func() {
		_ = os.Setenv("TERM", originalTerm)
		_ = os.Setenv("NO_COLOR", originalNoColor)
	}()

	_ = os.Setenv("TERM", "xterm-256color")
	_ = os.Setenv("NO_COLOR", "")

	tests := []struct {
		function func(string) string
		color    string
	}{
		{Success, Green},
		{Warning, Yellow},
		{Error, Red},
		{Info, Blue},
		{Prompt, Cyan + Bold},
		{Group, Magenta + Bold},
		{Software, Bold},
		{Dim, DimColor},
	}

	for _, test := range tests {
		result := test.function("test")
		expected := test.color + "test" + Reset
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	}
}

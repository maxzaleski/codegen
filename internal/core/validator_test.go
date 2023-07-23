package core

import (
	"testing"
)

func TestDirLikeValidation(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"dir/like/path", true},
		{"Dir/Like/Path", true},
		{"dir123", true},
		{"123dir", true},
		{"dir_like123", true},
		{"dir like", false},
		{"dir\\not\\like", false},
	}

	val := newValidator()
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if valid := val.Var(test.input, "dirlike"); valid == nil != test.expected {
				t.Errorf("Expected validation result %v for input '%s', but got %v", test.expected, test.input, valid)
			}
		})
	}
}

func TestPropTypeValidation(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"prop_type", true},
		{"prop-type", true},
		{"prop123", true},
		{"123prop", true},
		{"prop_type123", true},
		{"prop type", false},
		{"prop\\not\\type", false},
	}

	val := newValidator()
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if valid := val.Var(test.input, "proptype"); valid == nil != test.expected {
				t.Errorf("Expected validation result %v for input '%s', but got %v", test.expected, test.input, valid)
			}
		})
	}
}

func TestFileNameValidation(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"file.java", true},
		{"File.Java", true},
		{"foo-bar.java", true},
		{"\\{pkg.asTitle.asUpperCase\\}.java", true},
		{"\\{pkg.asTitle.asUpperCase\\}Service.java", true},
		{"invalid/file.java", false},
		{"file.java.invalid", false},
		{"\\{token1.token2.token3\\}", false},
		{"\\{token1.token2.token3.java", false},
		{"file.java.java", false},
	}

	val := newValidator()
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if valid := val.Var(test.input, "filename"); valid == nil != test.expected {
				t.Errorf("Expected validation result %v for input '%s', but got %v", test.expected, test.input, valid)
			}
		})
	}
}

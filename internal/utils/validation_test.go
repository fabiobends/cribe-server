package utils

import (
	"testing"
)

func TestValidation(t *testing.T) {
	// Test email validation
	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"invalid-email", false},
		{"", false},
		{"test@", false},
		{"@example.com", false},
		{"user@domain.co.uk", true},
	}

	for _, test := range tests {
		result := IsValidEmail(test.email)
		if result != test.expected {
			t.Errorf("IsValidEmail(%s) = %v; want %v", test.email, result, test.expected)
		}
	}

	// Test password validation
	passwordTests := []struct {
		password string
		expected bool
	}{
		{"12345678", true},
		{"short", false},
		{"", false},
		{"verylongpasswordthatmeetsrequirements", true},
	}

	for _, test := range passwordTests {
		result := IsValidPassword(test.password)
		if result != test.expected {
			t.Errorf("IsValidPassword(%s) = %v; want %v", test.password, result, test.expected)
		}
	}

	// Test name validation
	nameTests := []struct {
		name     string
		expected bool
	}{
		{"John", true},
		{"", false},
		{"  ", false},
		{"Mary Jane", true},
		{"José", true},
	}

	for _, test := range nameTests {
		result := IsValidName(test.name)
		if result != test.expected {
			t.Errorf("IsValidName(%s) = %v; want %v", test.name, result, test.expected)
		}
	}
}

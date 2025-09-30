package utils

import (
	"testing"
)

type TestStruct struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
	Name     string `validate:"required,alpha"`
}

func TestValidateStruct(t *testing.T) {
	// --- Valid struct ---
	valid := TestStruct{
		Email:    "User@example.com",
		Password: "password123",
		Name:     "Alice",
	}
	errResp := ValidateStruct(valid)
	if errResp != nil {
		t.Errorf("Expected nil error, got: %+v", errResp)
	}

	// --- Invalid struct ---
	invalid := TestStruct{
		Email:    "bad-email",
		Password: "short",
		Name:     "Alice123",
	}
	errResp = ValidateStruct(invalid)
	if errResp == nil {
		t.Errorf("Expected error response for invalid struct")
	}
}

func TestIsValidEmail(t *testing.T) {
	if !IsValidEmail("good@example.com") {
		t.Errorf("Expected valid email")
	}
	if IsValidEmail("bad-email") {
		t.Errorf("Expected invalid email")
	}
	if IsValidEmail("") {
		t.Errorf("Expected empty email to be invalid")
	}
}

func TestIsValidPassword(t *testing.T) {
	if !IsValidPassword("longenough") {
		t.Errorf("Expected valid password")
	}
	if IsValidPassword("short") {
		t.Errorf("Expected short password to be invalid")
	}
	if IsValidPassword("") {
		t.Errorf("Expected empty password to be invalid")
	}
}

func TestIsValidName(t *testing.T) {
	if !IsValidName("Alice") {
		t.Errorf("Expected valid name")
	}
	if IsValidName("Alice123") {
		t.Errorf("Expected invalid name with digits")
	}
	if IsValidName("") {
		t.Errorf("Expected empty name to be invalid")
	}
	if IsValidName("   ") {
		t.Errorf("Expected whitespace name to be invalid")
	}
}

func TestValidateRequiredFields(t *testing.T) {
	fields := []RequiredField{
		{Name: "Email", Value: "user@example.com"},
		{Name: "Password", Value: ""},
		{Name: "Name", Value: ""},
	}
	missing := ValidateRequiredFields(fields...)
	if len(missing) != 2 {
		t.Errorf("Expected 2 missing fields, got %d", len(missing))
	}
	if missing[0] != "Password" || missing[1] != "Name" {
		t.Errorf("Unexpected missing fields: %+v", missing)
	}
}

func TestValidateRequiredField(t *testing.T) {
	field := ValidateRequiredField("Username", "john")
	if field.Name != "Username" || field.Value != "john" {
		t.Errorf("Unexpected RequiredField: %+v", field)
	}
}

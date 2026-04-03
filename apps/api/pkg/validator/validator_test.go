package validator_test

import (
	"testing"

	"github.com/financeos/api/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"valid simple", "user@example.com", true},
		{"valid with subdomain", "user@mail.example.com", true},
		{"valid with plus", "user+tag@example.com", true},
		{"empty", "", false},
		{"no at sign", "userexample.com", false},
		{"no domain", "user@", false},
		{"no tld", "user@example", false},
		{"spaces", "user @example.com", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, validator.IsValidEmail(tc.email))
		})
	}
}

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"valid 8 chars", "password", true},
		{"valid long", "supersecurepassword123!", true},
		{"too short 7", "passwor", false},
		{"empty", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, validator.IsValidPassword(tc.password))
		})
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name string
		uuid string
		want bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", true},
		{"empty", "", false},
		{"too short", "550e8400-e29b", false},
		{"no hyphens", "550e8400e29b41d4a716446655440000", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, validator.IsValidUUID(tc.uuid))
		})
	}
}

func TestValidationErrors(t *testing.T) {
	var errs validator.ValidationErrors

	assert.False(t, errs.HasErrors())

	errs.Add("email", "is required")
	errs.Add("password", "must be at least 8 characters")

	assert.True(t, errs.HasErrors())
	assert.Len(t, errs, 2)
	assert.Contains(t, errs.Error(), "email")
	assert.Contains(t, errs.Error(), "password")
}

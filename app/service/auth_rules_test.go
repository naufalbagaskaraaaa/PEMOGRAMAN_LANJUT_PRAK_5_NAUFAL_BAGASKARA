package service

import "testing"

func TestCheckPasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{name: "terlalu pendek", password: "Ab1", valid: false},
		{name: "tanpa angka", password: "abcdefgh", valid: false},
		{name: "tanpa huruf", password: "12345678", valid: false},
		{name: "password umum", password: "password123", valid: false},
		{name: "valid", password: "Rahasia123", valid: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkPasswordStrength(tt.password) == ""
			if got != tt.valid {
				t.Fatalf("validitas password = %v, want %v", got, tt.valid)
			}
		})
	}
}

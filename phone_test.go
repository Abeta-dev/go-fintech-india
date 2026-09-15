package fintechin

import (
	"testing"
)

func TestMobileValidation(t *testing.T) {
	validNumbers := []string{
		"9876543210",
		"+919876543210",
		"+91 98765 43210",
		"+91-98765-43210",
		"919876543210",
		"09876543210",
		"8765432109",
		"7654321098",
		"6543210987",
	}

	for _, num := range validNumbers {
		if err := ValidateMobile(num); err != nil {
			t.Errorf("expected valid mobile for %s, got error: %v", num, err)
		}
		if !IsValidMobile(num) {
			t.Errorf("expected IsValidMobile to be true for %s", num)
		}
	}

	invalidCases := []struct {
		input       string
		expectedErr error
	}{
		{"", ErrInvalidMobileLength},
		{"987654321", ErrInvalidMobileLength},   // 9 digits
		{"98765432100", ErrInvalidMobileLength}, // 11 digits without 0
		{"5876543210", ErrInvalidMobileStartDigit},
		{"1876543210", ErrInvalidMobileStartDigit},
		{"987654321A", ErrInvalidMobileFormat},
		{"+19876543210", ErrInvalidMobileLength}, // Not +91
	}

	for _, tc := range invalidCases {
		err := ValidateMobile(tc.input)
		if err != tc.expectedErr {
			t.Errorf("ValidateMobile(%q) expected error %v, got %v", tc.input, tc.expectedErr, err)
		}
		if IsValidMobile(tc.input) {
			t.Errorf("IsValidMobile(%q) expected false, got true", tc.input)
		}
	}
}

func TestFormatE164(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"9876543210", "+919876543210"},
		{"+91 98765 43210", "+919876543210"},
		{"09876543210", "+919876543210"},
		{"919876543210", "+919876543210"},
		{"invalid", ""},
	}

	for _, tc := range tests {
		got := FormatE164(tc.input)
		if got != tc.expected {
			t.Errorf("FormatE164(%q) = %q, expected %q", tc.input, got, tc.expected)
		}
	}
}

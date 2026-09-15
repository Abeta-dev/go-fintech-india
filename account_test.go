package fintechin

import (
	"testing"
)

func TestBankAccountValidation(t *testing.T) {
	validAccounts := []string{
		"123456789",          // 9 digits (e.g. standard SBI/savings minimum)
		"123456789012",       // 12 digits
		"12345678901234",     // 14 digits
		"1234567890123456",   // 16 digits
		"123456789012345678", // 18 digits (maximum in Indian core banking)
	}

	for _, acc := range validAccounts {
		if err := ValidateBankAccount(acc); err != nil {
			t.Errorf("expected valid bank account for %s, got error: %v", acc, err)
		}
		if !IsValidBankAccount(acc) {
			t.Errorf("expected IsValidBankAccount to be true for %s", acc)
		}
	}

	invalidCases := []struct {
		input       string
		expectedErr error
	}{
		{"", ErrInvalidBankAccountLength},
		{"12345678", ErrInvalidBankAccountLength},            // 8 digits (too short)
		{"1234567890123456789", ErrInvalidBankAccountLength}, // 19 digits (too long)
		{"12345678A012", ErrInvalidBankAccountFormat},
		{"0000000000", ErrInvalidBankAccountAllZeros},
		{"000000000000000000", ErrInvalidBankAccountAllZeros},
	}

	for _, tc := range invalidCases {
		err := ValidateBankAccount(tc.input)
		if err != tc.expectedErr {
			t.Errorf("ValidateBankAccount(%q) expected error %v, got %v", tc.input, tc.expectedErr, err)
		}
		if IsValidBankAccount(tc.input) {
			t.Errorf("IsValidBankAccount(%q) expected false, got true", tc.input)
		}
	}
}

func TestMaskBankAccount(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123456789012", "XXXXXXXX9012"},
		{"123456789", "XXXXX6789"},
		{"1234", "XXXX"},
		{"12", "XX"},
	}

	for _, tc := range tests {
		got := MaskBankAccount(tc.input)
		if got != tc.expected {
			t.Errorf("MaskBankAccount(%q) = %q, expected %q", tc.input, got, tc.expected)
		}
	}
}

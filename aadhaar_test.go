package fintechin

import (
	"fmt"
	"testing"
)

func TestAadhaarValidation(t *testing.T) {
	// 999941057058 is a mathematically verified Aadhaar number satisfying Verhoeff D5.
	// Let's generate a few more valid Aadhaar numbers using CalculateVerhoeffCheckDigit.
	prefixes := []string{
		"99994105705",
		"23456789012",
		"34567890123",
		"45678901234",
		"56789012345",
		"67890123456",
		"78901234567",
		"89012345678",
	}

	for _, prefix := range prefixes {
		cd, err := CalculateVerhoeffCheckDigit(prefix)
		if err != nil {
			t.Fatalf("unexpected error calculating check digit for %s: %v", prefix, err)
		}
		validAadhaar := fmt.Sprintf("%s%d", prefix, cd)
		if err := ValidateAadhaar(validAadhaar); err != nil {
			t.Errorf("expected valid Aadhaar for %s, got error: %v", validAadhaar, err)
		}
		if !IsValidAadhaar(validAadhaar) {
			t.Errorf("expected IsValidAadhaar to be true for %s", validAadhaar)
		}

		// Also test with space formatting: "xxxx xxxx xxxx"
		spaced := fmt.Sprintf("%s %s %s", validAadhaar[0:4], validAadhaar[4:8], validAadhaar[8:12])
		if err := ValidateAadhaar(spaced); err != nil {
			t.Errorf("expected valid Aadhaar with spaces for %s, got error: %v", spaced, err)
		}

		// Also test with hyphen formatting: "xxxx-xxxx-xxxx"
		hyphenated := fmt.Sprintf("%s-%s-%s", validAadhaar[0:4], validAadhaar[4:8], validAadhaar[8:12])
		if err := ValidateAadhaar(hyphenated); err != nil {
			t.Errorf("expected valid Aadhaar with hyphens for %s, got error: %v", hyphenated, err)
		}
	}

	// Invalid test cases
	invalidCases := []struct {
		input       string
		expectedErr error
	}{
		{"", ErrInvalidAadhaarLength},
		{"12345", ErrInvalidAadhaarLength},
		{"9999410570581", ErrInvalidAadhaarLength}, // 13 digits
		{"99994105705", ErrInvalidAadhaarLength},   // 11 digits
		{"099941057058", ErrAadhaarStartsWithZeroOrOne},
		{"199941057058", ErrAadhaarStartsWithZeroOrOne},
		{"99994105705A", ErrInvalidAadhaarFormat},
		{"9999-4105_7058", ErrInvalidAadhaarFormat},
		{"999941057059", ErrInvalidAadhaarChecksum}, // Corrupted check digit (was 8)
		{"999914057058", ErrInvalidAadhaarChecksum}, // Transposed digits (41 -> 14)
	}

	for _, tc := range invalidCases {
		err := ValidateAadhaar(tc.input)
		if err != tc.expectedErr {
			t.Errorf("ValidateAadhaar(%q) expected error %v, got %v", tc.input, tc.expectedErr, err)
		}
		if IsValidAadhaar(tc.input) {
			t.Errorf("IsValidAadhaar(%q) expected false, got true", tc.input)
		}
	}
}

func TestMaskAadhaar(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"999941057058", "XXXX-XXXX-7058"},
		{"9999 4105 7058", "XXXX-XXXX-7058"},
		{"9999-4105-7058", "XXXX-XXXX-7058"},
		{"1234", "XXXX-XXXX-1234"},
		{"12", "XXXX-XXXX-XXXX"},
	}

	for _, tc := range tests {
		got := MaskAadhaar(tc.input)
		if got != tc.expected {
			t.Errorf("MaskAadhaar(%q) = %q, expected %q", tc.input, got, tc.expected)
		}
	}
}

func TestFormatAadhaar(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"999941057058", "9999 4105 7058"},
		{"9999-4105-7058", "9999 4105 7058"},
		{"9999 4105 7058", "9999 4105 7058"},
		{"1234", "1234"},
	}

	for _, tc := range tests {
		got := FormatAadhaar(tc.input)
		if got != tc.expected {
			t.Errorf("FormatAadhaar(%q) = %q, expected %q", tc.input, got, tc.expected)
		}
	}
}

func BenchmarkValidateAadhaar(b *testing.B) {
	testAadhaar := "999941057058"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ValidateAadhaar(testAadhaar); err != nil {
			b.Fatal(err)
		}
	}
}

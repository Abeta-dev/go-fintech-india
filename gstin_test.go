package fintechin

import (
	"testing"
)

func TestGSTINValidation(t *testing.T) {
	validGSTINs := []string{
		"29AACCG0527D1Z0", // Google India - Karnataka
		"27AACCG0527D1Z4", // Google India - Maharashtra
		"33AACCG0527D1ZB", // Google India - Tamil Nadu
		"06AACCG0527D1Z8", // Google India - Haryana
	}

	for _, g := range validGSTINs {
		if err := ValidateGSTIN(g); err != nil {
			t.Errorf("expected valid GSTIN for %s, got error: %v", g, err)
		}
		if !IsValidGSTIN(g) {
			t.Errorf("expected IsValidGSTIN to be true for %s", g)
		}
	}

	invalidCases := []struct {
		input       string
		expectedErr error
	}{
		{"", ErrInvalidGSTINLength},
		{"29AACCG0527D1Z", ErrInvalidGSTINLength},
		{"29AACCG0527D1Z00", ErrInvalidGSTINLength},
		{"00AACCG0527D1Z0", ErrInvalidStateCode},     // 00 is not a valid state code
		{"45AACCG0527D1Z0", ErrInvalidStateCode},     // 45 is not a valid state code
		{"29123CG0527D1Z0", ErrInvalidPANFormat},     // Invalid PAN
		{"29AACCG0527D-Z0", ErrInvalidGSTINFormat},   // Invalid 13th char
		{"29AACCG0527D1A0", ErrInvalidGSTINFormat},   // 14th char not 'Z'
		{"29AACCG0527D1Z1", ErrInvalidGSTINChecksum}, // Corrupted checksum digit
		{"27AACCG0527D1Z0", ErrInvalidGSTINChecksum}, // Wrong check digit for state 27
	}

	for _, tc := range invalidCases {
		err := ValidateGSTIN(tc.input)
		if err != tc.expectedErr {
			t.Errorf("ValidateGSTIN(%q) expected error %v, got %v", tc.input, tc.expectedErr, err)
		}
		if IsValidGSTIN(tc.input) {
			t.Errorf("IsValidGSTIN(%q) expected false, got true", tc.input)
		}
	}
}

func TestStateNameAndCode(t *testing.T) {
	// Test StateName
	tests := []struct {
		code         string
		expectedName string
	}{
		{"01", "Jammu and Kashmir"},
		{"07", "Delhi"},
		{"27", "Maharashtra"},
		{"29", "Karnataka"},
		{"33", "Tamil Nadu"},
		{"38", "Ladakh"},
		{"97", "Other Territory"},
		{"99", "Centre Jurisdiction"},
	}

	for _, tc := range tests {
		name, err := StateName(tc.code)
		if err != nil {
			t.Errorf("StateName(%q) unexpected error: %v", tc.code, err)
		}
		if name != tc.expectedName {
			t.Errorf("StateName(%q) = %q, expected %q", tc.code, name, tc.expectedName)
		}
	}

	_, err := StateName("98")
	if err != ErrInvalidStateCode {
		t.Errorf("expected ErrInvalidStateCode for code 98, got %v", err)
	}

	// Test StateCode
	if code := StateCode("29AACCG0527D1Z0"); code != "29" {
		t.Errorf("StateCode() = %q, expected 29", code)
	}
	if code := StateCode("2"); code != "" {
		t.Errorf("StateCode() expected empty string for short input, got %q", code)
	}
}

func TestExtractPAN(t *testing.T) {
	gstin := "29AACCG0527D1Z0"
	pan := ExtractPAN(gstin)
	if pan != "AACCG0527D" {
		t.Errorf("ExtractPAN(%q) = %q, expected AACCG0527D", gstin, pan)
	}

	// Invalid GSTIN should return empty string
	if ExtractPAN("INVALID") != "" {
		t.Errorf("expected empty string for invalid GSTIN")
	}
	if ExtractPAN("2912345678901Z0") != "" {
		t.Errorf("expected empty string for invalid embedded PAN")
	}
}

func BenchmarkValidateGSTIN(b *testing.B) {
	testGSTIN := "29AACCG0527D1Z0"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ValidateGSTIN(testGSTIN); err != nil {
			b.Fatal(err)
		}
	}
}

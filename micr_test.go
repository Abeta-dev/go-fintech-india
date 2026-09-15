package fintechin

import (
	"testing"
)

func TestMICRValidation(t *testing.T) {
	validMICRs := []string{
		"110002001", // Delhi SBI
		"400240002", // Mumbai HDFC
		"560002015", // Bangalore Canara
	}

	for _, m := range validMICRs {
		if err := ValidateMICR(m); err != nil {
			t.Errorf("expected valid MICR for %s, got error: %v", m, err)
		}
		if !IsValidMICR(m) {
			t.Errorf("expected IsValidMICR to be true for %s", m)
		}
	}

	invalidCases := []struct {
		input       string
		expectedErr error
	}{
		{"", ErrInvalidMICRLength},
		{"11000200", ErrInvalidMICRLength},   // 8 digits
		{"1100020011", ErrInvalidMICRLength}, // 10 digits
		{"11000200A", ErrInvalidMICRFormat},
		{"000000000", ErrInvalidMICRAllZeros},
	}

	for _, tc := range invalidCases {
		err := ValidateMICR(tc.input)
		if err != tc.expectedErr {
			t.Errorf("ValidateMICR(%q) expected error %v, got %v", tc.input, tc.expectedErr, err)
		}
		if IsValidMICR(tc.input) {
			t.Errorf("IsValidMICR(%q) expected false, got true", tc.input)
		}
	}
}

func TestMICRComponents(t *testing.T) {
	micr := "110002001"
	if city := MICRCityCode(micr); city != "110" {
		t.Errorf("MICRCityCode(%q) = %q, expected 110", micr, city)
	}
	if bank := MICRBankCode(micr); bank != "002" {
		t.Errorf("MICRBankCode(%q) = %q, expected 002", micr, bank)
	}
	if branch := MICRBranchCode(micr); branch != "001" {
		t.Errorf("MICRBranchCode(%q) = %q, expected 001", micr, branch)
	}

	if city := MICRCityCode("11"); city != "" {
		t.Errorf("expected empty string for short MICR in MICRCityCode")
	}
	if bank := MICRBankCode("1100"); bank != "" {
		t.Errorf("expected empty string for short MICR in MICRBankCode")
	}
	if branch := MICRBranchCode("110002"); branch != "" {
		t.Errorf("expected empty string for short MICR in MICRBranchCode")
	}
}

package fintechin

import (
	"testing"
)

func TestIFSCValidation(t *testing.T) {
	validIFSCs := []string{
		"SBIN0000001",
		"HDFC0000123",
		"ICIC0001234",
		"UTIB0000456",
		"KKBK0000789",
		"PUNB0123456",
		"BARB0VJALWA",
	}

	for _, code := range validIFSCs {
		if err := ValidateIFSC(code); err != nil {
			t.Errorf("expected valid IFSC for %s, got error: %v", code, err)
		}
		if !IsValidIFSC(code) {
			t.Errorf("expected IsValidIFSC to be true for %s", code)
		}
	}

	invalidCases := []struct {
		input       string
		expectedErr error
	}{
		{"", ErrInvalidIFSCLength},
		{"SBIN000001", ErrInvalidIFSCLength},   // 10 chars
		{"SBIN00000001", ErrInvalidIFSCLength}, // 12 chars
		{"1BIN0000001", ErrInvalidIFSCBankCode},
		{"sbiN0000001", ErrInvalidIFSCBankCode},
		{"SBIN1000001", ErrInvalidIFSCFifthChar}, // 5th char is '1', must be '0'
		{"SBIN000000$", ErrInvalidIFSCBranchCode},
	}

	for _, tc := range invalidCases {
		err := ValidateIFSC(tc.input)
		if err != tc.expectedErr {
			t.Errorf("ValidateIFSC(%q) expected error %v, got %v", tc.input, tc.expectedErr, err)
		}
		if IsValidIFSC(tc.input) {
			t.Errorf("IsValidIFSC(%q) expected false, got true", tc.input)
		}
	}
}

func TestBankAndBranchExtraction(t *testing.T) {
	ifsc := "SBIN0000001"
	if b := BankCode(ifsc); b != "SBIN" {
		t.Errorf("BankCode(%q) = %q, expected SBIN", ifsc, b)
	}
	if br := BranchCode(ifsc); br != "000001" {
		t.Errorf("BranchCode(%q) = %q, expected 000001", ifsc, br)
	}

	if b := BankCode("SB"); b != "" {
		t.Errorf("expected empty string for short IFSC in BankCode")
	}
	if br := BranchCode("SHORT"); br != "" {
		t.Errorf("expected empty string for short IFSC in BranchCode")
	}
}

func TestBankNameFromIFSC(t *testing.T) {
	tests := []struct {
		ifsc         string
		expectedName string
	}{
		{"SBIN0000001", "State Bank of India"},
		{"HDFC0001234", "HDFC Bank"},
		{"ICIC0001234", "ICICI Bank"},
		{"UTIB0000456", "Axis Bank"},
		{"KKBK0000789", "Kotak Mahindra Bank"},
		{"UNKNOWN0001", ""},
		{"", ""},
	}

	for _, tc := range tests {
		name := BankNameFromIFSC(tc.ifsc)
		if name != tc.expectedName {
			t.Errorf("BankNameFromIFSC(%q) = %q, expected %q", tc.ifsc, name, tc.expectedName)
		}
	}
}

func BenchmarkValidateIFSC(b *testing.B) {
	testIFSC := "SBIN0000001"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ValidateIFSC(testIFSC); err != nil {
			b.Fatal(err)
		}
	}
}

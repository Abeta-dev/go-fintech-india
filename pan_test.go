package fintechin

import (
	"testing"
)

func TestPANValidation(t *testing.T) {
	validPANs := []string{
		"ABCDE1234F",
		"AAAPL1234C",
		"BNZPP9876K",
		"ZZZZZ9999Z",
	}

	for _, p := range validPANs {
		if err := ValidatePAN(p); err != nil {
			t.Errorf("expected valid PAN for %s, got error: %v", p, err)
		}
		if !IsValidPAN(p) {
			t.Errorf("expected IsValidPAN to be true for %s", p)
		}
	}

	invalidPANs := []struct {
		input       string
		expectedErr error
	}{
		{"", ErrInvalidPANLength},
		{"ABCDE1234", ErrInvalidPANLength},
		{"ABCDE1234FA", ErrInvalidPANLength},
		{"1BCDE1234F", ErrInvalidPANFormat}, // digit at pos 0
		{"ABCD11234F", ErrInvalidPANFormat}, // digit at pos 4
		{"ABCDE123AF", ErrInvalidPANFormat}, // letter at numeric pos
		{"ABCDE12341", ErrInvalidPANFormat}, // digit at last pos
		{"abcde1234f", ErrInvalidPANFormat}, // lowercase letters
	}

	for _, tc := range invalidPANs {
		err := ValidatePAN(tc.input)
		if err != tc.expectedErr {
			t.Errorf("ValidatePAN(%q) expected error %v, got %v", tc.input, tc.expectedErr, err)
		}
		if IsValidPAN(tc.input) {
			t.Errorf("IsValidPAN(%q) expected false, got true", tc.input)
		}
	}
}

func TestPANEntityType(t *testing.T) {
	cases := []struct {
		pan          string
		expectedType PANEntityType
	}{
		{"ABCPA1234F", EntityTypeIndividual},
		{"ABCCA1234F", EntityTypeCompany},
		{"ABCHA1234F", EntityTypeHUF},
		{"ABCFA1234F", EntityTypeFirm},
		{"ABCAA1234F", EntityTypeAOP},
		{"ABCTA1234F", EntityTypeTrust},
		{"ABCBA1234F", EntityTypeBOI},
		{"ABCLA1234F", EntityTypeLocalAuthority},
		{"ABCJA1234F", EntityTypeArtificialJuridicalPerson},
		{"ABCGA1234F", EntityTypeGovernmentAgency},
	}

	for _, tc := range cases {
		et, err := EntityType(tc.pan)
		if err != nil {
			t.Errorf("EntityType(%q) unexpected error: %v", tc.pan, err)
		}
		if et != tc.expectedType {
			t.Errorf("EntityType(%q) = %q, expected %q", tc.pan, et, tc.expectedType)
		}
	}

	// Invalid entity code
	_, err := EntityType("ABCZA1234F")
	if err != ErrInvalidPANEntityType {
		t.Errorf("expected ErrInvalidPANEntityType, got %v", err)
	}

	// Malformed PAN
	_, err = EntityType("INVALID")
	if err == nil {
		t.Errorf("expected error for malformed PAN, got nil")
	}
}

func TestMatchesSurname(t *testing.T) {
	// 5th character of "ABCDS1234F" is 'S'
	pan := "ABCDS1234F"
	if !MatchesSurname(pan, "Sharma") {
		t.Errorf("expected true for Sharma, got false")
	}
	if !MatchesSurname(pan, "sharma") {
		t.Errorf("expected true for case-insensitive sharma, got false")
	}
	if !MatchesSurname(pan, "  Singh  ") {
		t.Errorf("expected true for leading space Singh, got false")
	}
	if MatchesSurname(pan, "Verma") {
		t.Errorf("expected false for Verma, got true")
	}
	if MatchesSurname("INVALID", "Sharma") {
		t.Errorf("expected false for invalid PAN, got true")
	}
	if MatchesSurname(pan, "") {
		t.Errorf("expected false for empty surname, got true")
	}
}

func TestMaskPAN(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ABCDE1234F", "XXXXX1234X"},
		{"BNZPP9876K", "XXXXX9876X"},
		{"SHORT", "XXXXX"},
		{"12", "XX"},
	}

	for _, tc := range tests {
		got := MaskPAN(tc.input)
		if got != tc.expected {
			t.Errorf("MaskPAN(%q) = %q, expected %q", tc.input, got, tc.expected)
		}
	}
}

func BenchmarkValidatePAN(b *testing.B) {
	testPAN := "ABCDE1234F"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ValidatePAN(testPAN); err != nil {
			b.Fatal(err)
		}
	}
}

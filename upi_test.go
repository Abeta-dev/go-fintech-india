package fintechin

import (
	"strings"
	"testing"
)

func TestUPIValidation(t *testing.T) {
	validVPAs := []string{
		"merchant@okhdfcbank",
		"user.name@okaxis",
		"9876543210@paytm",
		"user-sub_acc@ybl",
		"test1234@upi",
		"john.doe.123@icici",
	}

	for _, vpa := range validVPAs {
		if err := ValidateUPI(vpa); err != nil {
			t.Errorf("expected valid UPI for %s, got error: %v", vpa, err)
		}
		if !IsValidUPI(vpa) {
			t.Errorf("expected IsValidUPI to be true for %s", vpa)
		}
	}

	invalidCases := []struct {
		input       string
		expectedErr error
	}{
		{"", ErrInvalidUPIFormat},
		{"noatsign", ErrInvalidUPIFormat},
		{"multiple@@signs", ErrInvalidUPIFormat},
		{"@nohandle", ErrInvalidUPIFormat},
		{"nouser@", ErrInvalidUPIFormat},
		{".startwithdot@upi", ErrInvalidUPIUsername},
		{"endwithdot.@upi", ErrInvalidUPIUsername},
		{"consecutive..dots@upi", ErrInvalidUPIUsername},
		{"invalid char@upi", ErrInvalidUPIUsername},
		{"user@a", ErrInvalidUPIHandle}, // handle < 2 chars
		{"user@bad#handle", ErrInvalidUPIHandle},
	}

	for _, tc := range invalidCases {
		err := ValidateUPI(tc.input)
		if err != tc.expectedErr {
			t.Errorf("ValidateUPI(%q) expected error %v, got %v", tc.input, tc.expectedErr, err)
		}
		if IsValidUPI(tc.input) {
			t.Errorf("IsValidUPI(%q) expected false, got true", tc.input)
		}
	}
}

func TestKnownPSPHandle(t *testing.T) {
	if !IsKnownPSPHandle("okaxis") {
		t.Errorf("expected okaxis to be recognized as known PSP")
	}
	if !IsKnownPSPHandle("OKHDFCBANK") {
		t.Errorf("expected case-insensitive match for OKHDFCBANK")
	}
	if !IsKnownPSPHandle("paytm") {
		t.Errorf("expected paytm to be recognized as known PSP")
	}
	if IsKnownPSPHandle("unknownpsp123") {
		t.Errorf("expected unknownpsp123 to not be recognized as known PSP")
	}
}

func TestGenerateAndParseUPIURI(t *testing.T) {
	params := UPIParams{
		PayeeAddress: "store@okhdfcbank",
		PayeeName:    "Super Mart",
		Amount:       "500.50",
		Currency:     "INR",
		Note:         "Groceries & Essentials",
		RefID:        "TXN123456789",
		MerchantCode: "5411",
		URL:          "https://example.com/order/123",
	}

	uri, err := GenerateUPIURI(params)
	if err != nil {
		t.Fatalf("GenerateUPIURI failed: %v", err)
	}

	if !strings.HasPrefix(uri, "upi://pay?") {
		t.Fatalf("expected URI to start with 'upi://pay?', got %q", uri)
	}

	parsed, err := ParseUPIURI(uri)
	if err != nil {
		t.Fatalf("ParseUPIURI failed: %v", err)
	}

	if parsed.PayeeAddress != params.PayeeAddress {
		t.Errorf("PayeeAddress = %q, expected %q", parsed.PayeeAddress, params.PayeeAddress)
	}
	if parsed.PayeeName != params.PayeeName {
		t.Errorf("PayeeName = %q, expected %q", parsed.PayeeName, params.PayeeName)
	}
	if parsed.Amount != params.Amount {
		t.Errorf("Amount = %q, expected %q", parsed.Amount, params.Amount)
	}
	if parsed.Currency != params.Currency {
		t.Errorf("Currency = %q, expected %q", parsed.Currency, params.Currency)
	}
	if parsed.Note != params.Note {
		t.Errorf("Note = %q, expected %q", parsed.Note, params.Note)
	}
	if parsed.RefID != params.RefID {
		t.Errorf("RefID = %q, expected %q", parsed.RefID, params.RefID)
	}
	if parsed.MerchantCode != params.MerchantCode {
		t.Errorf("MerchantCode = %q, expected %q", parsed.MerchantCode, params.MerchantCode)
	}
	if parsed.URL != params.URL {
		t.Errorf("URL = %q, expected %q", parsed.URL, params.URL)
	}
}

func TestGenerateUPIURICRLFInjection(t *testing.T) {
	malicious := UPIParams{
		PayeeAddress: "hacker@upi\r\nInjected-Header: evil",
		PayeeName:    "Hacker",
	}

	_, err := GenerateUPIURI(malicious)
	if err != ErrInvalidUPICRLF {
		t.Errorf("expected ErrInvalidUPICRLF, got %v", err)
	}
}

func TestGenerateUPIURIInvalidAmount(t *testing.T) {
	invalidAmounts := []string{"-10.00", "0", "abc", "10.999"}
	for _, amt := range invalidAmounts {
		p := UPIParams{
			PayeeAddress: "merchant@upi",
			Amount:       amt,
		}
		_, err := GenerateUPIURI(p)
		if err != ErrInvalidUPIAmount {
			t.Errorf("GenerateUPIURI(amount=%q) expected ErrInvalidUPIAmount, got %v", amt, err)
		}
	}
}

func BenchmarkValidateUPI(b *testing.B) {
	vpa := "merchant.store-123@okhdfcbank"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ValidateUPI(vpa); err != nil {
			b.Fatal(err)
		}
	}
}

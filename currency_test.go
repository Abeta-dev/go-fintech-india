package fintechin

import (
	"testing"
)

func TestFormatINR(t *testing.T) {
	tests := []struct {
		paise int64
		want  string
	}{
		{0, "0.00"},
		{1, "0.01"},
		{50, "0.50"},
		{100, "1.00"},
		{100000, "1,000.00"},
		{10000000, "1,00,000.00"},      // 1 Lakh Rupees
		{1000000000, "1,00,00,000.00"}, // 1 Crore Rupees
		{123456789, "12,34,567.89"},
		{-5000000, "-50,000.00"},
		{-1, "-0.01"},
		{-100, "-1.00"},
	}

	for _, tt := range tests {
		got := FormatINR(tt.paise)
		if got != tt.want {
			t.Errorf("FormatINR(%d) = %q, want %q", tt.paise, got, tt.want)
		}
	}
}

func TestFormatINRSymbol(t *testing.T) {
	tests := []struct {
		paise int64
		want  string
	}{
		{0, "₹0.00"},
		{123456789, "₹12,34,567.89"},
		{-5000000, "-₹50,000.00"},
		{100, "₹1.00"},
		{-100, "-₹1.00"},
	}

	for _, tt := range tests {
		got := FormatINRSymbol(tt.paise)
		if got != tt.want {
			t.Errorf("FormatINRSymbol(%d) = %q, want %q", tt.paise, got, tt.want)
		}
	}
}

func TestParseINR(t *testing.T) {
	valid := []struct {
		input string
		want  int64
	}{
		{"₹12,34,567.89", 123456789},
		{"12,34,567.89", 123456789},
		{"1234567.89", 123456789},
		{"-₹50,000.00", -5000000},
		{"₹ -50,000.00", -5000000},
		{"-50,000.00", -5000000},
		{"Rs. 1,000.50", 100050},
		{"Rs 1,000.50", 100050},
		{"INR 500", 50000},
		{"(500.00)", -50000},
		{"0.5", 50},
		{".5", 50},
		{".05", 5},
		{"0.05", 5},
		{"0", 0},
		{"100", 10000},
		{"-100", -10000},
		{"-.5", -50},
	}

	for _, tt := range valid {
		m, err := ParseINR(tt.input)
		if err != nil {
			t.Errorf("ParseINR(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if m.Paise() != tt.want {
			t.Errorf("ParseINR(%q) = %d paise, want %d paise", tt.input, m.Paise(), tt.want)
		}
	}

	invalid := []string{
		"",
		"   ",
		"₹",
		"abc",
		"1.2.3",
		"12a34",
		"₹ --50",
	}

	for _, s := range invalid {
		if _, err := ParseINR(s); err == nil {
			t.Errorf("ParseINR(%q) expected error, got nil", s)
		}
	}
}

func BenchmarkFormatINR(b *testing.B) {
	paise := int64(123456789)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatINR(paise)
	}
}

func BenchmarkParseINR(b *testing.B) {
	s := "₹12,34,567.89"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseINR(s)
	}
}

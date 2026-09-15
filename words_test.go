package fintechin

import (
	"testing"
)

func TestNumberToIndianWords(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "Zero"},
		{1, "One"},
		{10, "Ten"},
		{14, "Fourteen"},
		{20, "Twenty"},
		{23, "Twenty-Three"},
		{99, "Ninety-Nine"},
		{100, "One Hundred"},
		{105, "One Hundred Five"},
		{456, "Four Hundred Fifty-Six"},
		{1000, "One Thousand"},
		{23000, "Twenty-Three Thousand"},
		{100000, "One Lakh"},
		{123456, "One Lakh Twenty-Three Thousand Four Hundred Fifty-Six"},
		{10000000, "One Crore"},
		{100000000, "Ten Crore"},
		{1000000000, "One Hundred Crore"},
		{1200000000, "One Hundred Twenty Crore"},
		{-50, "Minus Fifty"},
	}

	for _, tt := range tests {
		got := NumberToIndianWords(tt.n)
		if got != tt.want {
			t.Errorf("NumberToIndianWords(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestInWords(t *testing.T) {
	tests := []struct {
		m    Money
		want string
	}{
		// Prompt examples
		{
			NewMoney(12345678),
			"Rupees One Lakh Twenty-Three Thousand Four Hundred Fifty-Six and Seventy-Eight Paise Only",
		},
		{
			NewMoney(0),
			"Rupees Zero Only",
		},
		{
			NewMoney(-12345678),
			"Minus Rupees One Lakh Twenty-Three Thousand Four Hundred Fifty-Six and Seventy-Eight Paise Only",
		},
		{
			NewMoney(78),
			"Seventy-Eight Paise Only",
		},
		{
			NewMoney(-78),
			"Minus Seventy-Eight Paise Only",
		},
		// Additional edge cases
		{
			NewMoney(100),
			"Rupees One Only",
		},
		{
			NewMoney(1),
			"One Paisa Only",
		},
		{
			NewMoney(101),
			"Rupees One and One Paisa Only",
		},
		{
			NewMoney(10000), // Rs 100
			"Rupees One Hundred Only",
		},
		{
			NewMoney(1000000000), // 1 Crore Rs
			"Rupees One Crore Only",
		},
		{
			NewMoney(-5000000),
			"Minus Rupees Fifty Thousand Only",
		},
	}

	for _, tt := range tests {
		got := InWords(tt.m)
		if got != tt.want {
			t.Errorf("InWords(%d) = %q, want %q", tt.m.Paise(), got, tt.want)
		}
	}
}

func BenchmarkInWords(b *testing.B) {
	m := NewMoney(12345678)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InWords(m)
	}
}

func BenchmarkNumberToIndianWords(b *testing.B) {
	n := int64(123456)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NumberToIndianWords(n)
	}
}

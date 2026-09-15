package fintechin

import (
	"fmt"
	"math"
	"strings"
)

var ones = []string{
	"", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine",
	"Ten", "Eleven", "Twelve", "Thirteen", "Fourteen", "Fifteen", "Sixteen",
	"Seventeen", "Eighteen", "Nineteen",
}

var tens = []string{
	"", "", "Twenty", "Thirty", "Forty", "Fifty", "Sixty", "Seventy", "Eighty", "Ninety",
}

// convertTwoDigits converts numbers from 1 to 99 into words (e.g. 23 -> "Twenty-Three").
func convertTwoDigits(n int64) string {
	if n < 20 {
		return ones[n]
	}
	t := tens[n/10]
	u := ones[n%10]
	if u != "" {
		return t + "-" + u
	}
	return t
}

// NumberToIndianWords converts an integer into words according to Indian numbering convention
// (Crores, Lakhs, Thousands, Hundreds, Units).
func NumberToIndianWords(n int64) string {
	if n == 0 {
		return "Zero"
	}
	if n < 0 {
		if n == math.MinInt64 {
			// Avoid 64-bit overflow on negation
			return "Minus Nine Quintillion Two Hundred Twenty-Three Quadrillion Three Hundred Seventy-Two Trillion Thirty-Six Billion Eight Hundred Fifty-Four Million Seven Hundred Seventy-Five Thousand Eight Hundred Eight"
		}
		return "Minus " + NumberToIndianWords(-n)
	}

	var parts []string

	// Crores (10,000,000 = 10^7)
	if n >= 10000000 {
		crores := n / 10000000
		n %= 10000000
		parts = append(parts, NumberToIndianWords(crores)+" Crore")
	}

	// Lakhs (100,000 = 10^5)
	if n >= 100000 {
		lakhs := n / 100000
		n %= 100000
		parts = append(parts, convertTwoDigits(lakhs)+" Lakh")
	}

	// Thousands (1,000 = 10^3)
	if n >= 1000 {
		thousands := n / 1000
		n %= 1000
		parts = append(parts, convertTwoDigits(thousands)+" Thousand")
	}

	// Hundreds (100 = 10^2)
	if n >= 100 {
		hundreds := n / 100
		n %= 100
		parts = append(parts, ones[hundreds]+" Hundred")
	}

	// Tens and Ones (1-99)
	if n > 0 {
		parts = append(parts, convertTwoDigits(n))
	}

	return strings.Join(parts, " ")
}

// InWords converts Money into words according to Indian banking convention.
// Example: 12345678 paise (Rs 1,23,456.78) ->
// "Rupees One Lakh Twenty-Three Thousand Four Hundred Fifty-Six and Seventy-Eight Paise Only"
func InWords(m Money) string {
	if m.paise == 0 {
		return "Rupees Zero Only"
	}

	isNegative := m.paise < 0
	absPaise := m.paise
	if isNegative {
		absPaise = -absPaise
	}

	rupees := absPaise / 100
	paise := absPaise % 100

	paiseUnit := "Paise"
	if paise == 1 {
		paiseUnit = "Paisa"
	}

	var result string
	switch {
	case rupees > 0 && paise > 0:
		result = fmt.Sprintf("Rupees %s and %s %s Only", NumberToIndianWords(rupees), convertTwoDigits(paise), paiseUnit)
	case rupees > 0 && paise == 0:
		result = fmt.Sprintf("Rupees %s Only", NumberToIndianWords(rupees))
	case rupees == 0 && paise > 0:
		result = fmt.Sprintf("%s %s Only", convertTwoDigits(paise), paiseUnit)
	}

	if isNegative {
		result = "Minus " + result
	}

	return result
}

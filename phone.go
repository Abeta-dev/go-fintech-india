package fintechin

import (
	"errors"
	"strings"
)

var (
	// ErrInvalidMobileLength indicates the mobile number does not resolve to 10 digits.
	ErrInvalidMobileLength = errors.New("mobile: number must resolve to exactly 10 digits")
	// ErrInvalidMobileFormat indicates the mobile number contains invalid characters.
	ErrInvalidMobileFormat = errors.New("mobile: number must contain only numeric digits")
	// ErrInvalidMobileStartDigit indicates the mobile number does not begin with DoT allocated blocks (6, 7, 8, or 9).
	ErrInvalidMobileStartDigit = errors.New("mobile: number must start with 6, 7, 8, or 9")
)

// NormalizeMobile cleans an Indian mobile number by removing formatting, "+91", "91" prefix, or leading "0".
// Returns the clean 10-digit number or an error if invalid.
func NormalizeMobile(s string) (string, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return "", ErrInvalidMobileLength
	}

	// Filter out common formatting characters: space, dash, parentheses, plus
	var digits [16]byte
	dLen := 0
	hasPlus := false

	for i := 0; i < len(trimmed); i++ {
		c := trimmed[i]
		if c == '+' && i == 0 {
			hasPlus = true
			continue
		}
		if c == ' ' || c == '-' || c == '(' || c == ')' {
			continue
		}
		if c < '0' || c > '9' {
			return "", ErrInvalidMobileFormat
		}
		if dLen < len(digits) {
			digits[dLen] = c
			dLen++
		} else {
			return "", ErrInvalidMobileLength
		}
	}

	raw := string(digits[:dLen])

	// Handle country code prefixes
	switch {
	case hasPlus:
		// Expect +91 followed by 10 digits (12 total numeric digits)
		if len(raw) == 12 && raw[0] == '9' && raw[1] == '1' {
			raw = raw[2:]
		} else {
			return "", ErrInvalidMobileLength
		}
	case len(raw) == 12 && raw[0] == '9' && raw[1] == '1':
		// Without +, but starts with 91 and total 12 digits
		raw = raw[2:]
	case len(raw) == 11 && raw[0] == '0':
		// Leading trunk prefix '0'
		raw = raw[1:]
	}

	if len(raw) != 10 {
		return "", ErrInvalidMobileLength
	}

	// Department of Telecommunications (DoT) rule: Indian mobile numbers start with 6, 7, 8, or 9
	start := raw[0]
	if start != '6' && start != '7' && start != '8' && start != '9' {
		return "", ErrInvalidMobileStartDigit
	}

	return raw, nil
}

// ValidateMobile validates an Indian 10-digit mobile number according to DoT allocation rules.
func ValidateMobile(s string) error {
	_, err := NormalizeMobile(s)
	return err
}

// IsValidMobile reports whether the given string is a valid Indian mobile number.
func IsValidMobile(s string) bool {
	return ValidateMobile(s) == nil
}

// FormatE164 formats a valid Indian mobile number into international E.164 format: "+919876543210".
// Returns an empty string if the input mobile number is invalid.
func FormatE164(s string) string {
	clean, err := NormalizeMobile(s)
	if err != nil {
		return ""
	}
	return "+91" + clean
}

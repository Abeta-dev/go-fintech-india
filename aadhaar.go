package fintechin

import (
	"errors"
	"strings"
)

var (
	// ErrInvalidAadhaarLength indicates the Aadhaar number is not 12 digits.
	ErrInvalidAadhaarLength = errors.New("aadhaar: number must be exactly 12 digits")
	// ErrInvalidAadhaarFormat indicates the Aadhaar number contains invalid characters.
	ErrInvalidAadhaarFormat = errors.New("aadhaar: number must contain only numeric digits")
	// ErrAadhaarStartsWithZeroOrOne indicates the Aadhaar number starts with 0 or 1, which UIDAI disallows.
	ErrAadhaarStartsWithZeroOrOne = errors.New("aadhaar: number cannot start with 0 or 1")
	// ErrInvalidAadhaarChecksum indicates the Aadhaar number fails the Verhoeff checksum check.
	ErrInvalidAadhaarChecksum = errors.New("aadhaar: invalid verhoeff checksum")
)

// Official UIDAI Verhoeff Dihedral D5 algorithm tables.
// verhoeffD is the multiplication table for the dihedral group D5.
var verhoeffD = [10][10]int{
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	{1, 2, 3, 4, 0, 6, 7, 8, 9, 5},
	{2, 3, 4, 0, 1, 7, 8, 9, 5, 6},
	{3, 4, 0, 1, 2, 8, 9, 5, 6, 7},
	{4, 0, 1, 2, 3, 9, 5, 6, 7, 8},
	{5, 9, 8, 7, 6, 0, 4, 3, 2, 1},
	{6, 5, 9, 8, 7, 1, 0, 4, 3, 2},
	{7, 6, 5, 9, 8, 2, 1, 0, 4, 3},
	{8, 7, 6, 5, 9, 3, 2, 1, 0, 4},
	{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
}

// verhoeffP is the permutation table based on the permutation (1 5 8 9 4 2 7 0)(3 6).
var verhoeffP = [8][10]int{
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	{1, 5, 7, 6, 2, 8, 3, 0, 9, 4},
	{5, 8, 0, 3, 7, 9, 6, 1, 4, 2},
	{8, 9, 1, 6, 0, 4, 3, 5, 2, 7},
	{9, 4, 5, 3, 1, 2, 6, 8, 7, 0},
	{4, 2, 8, 6, 5, 7, 3, 9, 0, 1},
	{2, 7, 9, 3, 8, 0, 6, 4, 1, 5},
	{7, 0, 4, 6, 9, 1, 3, 2, 5, 8},
}

// verhoeffInv is the multiplicative inverse table in D5.
var verhoeffInv = [10]int{0, 4, 3, 2, 1, 5, 6, 7, 8, 9}

// CalculateVerhoeffCheckDigit computes the Verhoeff check digit for any numeric string without allocating.
func CalculateVerhoeffCheckDigit(num string) (int, error) {
	if num == "" {
		return -1, ErrInvalidAadhaarLength
	}
	c := 0
	l := len(num)
	for i := 0; i < l; i++ {
		ch := num[l-1-i]
		if ch < '0' || ch > '9' {
			return -1, ErrInvalidAadhaarFormat
		}
		digit := int(ch - '0')
		c = verhoeffD[c][verhoeffP[(i+1)%8][digit]]
	}
	return verhoeffInv[c], nil
}

// ValidateAadhaar validates an Indian Aadhaar number according to UIDAI rules:
// - Must be exactly 12 numeric digits (accepts raw 12 digits or formatted 14 chars with spaces/hyphens)
// - Cannot start with 0 or 1
// - Must pass the official Verhoeff Dihedral D5 checksum algorithm
// Zero heap allocations when validating standard 12-digit strings.
func ValidateAadhaar(s string) error {
	var digits [12]byte
	switch len(s) {
	case 12:
		for i := 0; i < 12; i++ {
			c := s[i]
			if c < '0' || c > '9' {
				return ErrInvalidAadhaarFormat
			}
			digits[i] = c
		}
	case 14:
		// Accept standard formatted: "1234 5678 9012" or "1234-5678-9012"
		sep1, sep2 := s[4], s[9]
		if (sep1 != ' ' || sep2 != ' ') && (sep1 != '-' || sep2 != '-') {
			return ErrInvalidAadhaarFormat
		}
		j := 0
		for i := 0; i < 14; i++ {
			if i == 4 || i == 9 {
				continue
			}
			c := s[i]
			if c < '0' || c > '9' {
				return ErrInvalidAadhaarFormat
			}
			digits[j] = c
			j++
		}
	default:
		return ErrInvalidAadhaarLength
	}

	// UIDAI rule: Aadhaar number cannot start with 0 or 1
	if digits[0] == '0' || digits[0] == '1' {
		return ErrAadhaarStartsWithZeroOrOne
	}

	// Verhoeff checksum validation
	c := 0
	for i := 0; i < 12; i++ {
		digit := int(digits[11-i] - '0')
		c = verhoeffD[c][verhoeffP[i%8][digit]]
	}

	if c != 0 {
		return ErrInvalidAadhaarChecksum
	}

	return nil
}

// IsValidAadhaar reports whether the given string is a valid Aadhaar number.
func IsValidAadhaar(s string) bool {
	return ValidateAadhaar(s) == nil
}

// MaskAadhaar returns the masked Aadhaar representation: "XXXX-XXXX-1234".
// Accepts raw 12 digits or formatted Aadhaar strings.
func MaskAadhaar(s string) string {
	// Extract digits into fixed buffer
	var digits [12]byte
	dLen := 0
	for i := 0; i < len(s) && dLen < 12; i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			digits[dLen] = c
			dLen++
		}
	}
	if dLen < 4 {
		return "XXXX-XXXX-XXXX"
	}
	last4 := string(digits[dLen-4 : dLen])
	return "XXXX-XXXX-" + last4
}

// FormatAadhaar formats an Aadhaar string into the standard UIDAI visual pattern "1234 5678 9012".
func FormatAadhaar(s string) string {
	var digits [12]byte
	dLen := 0
	for i := 0; i < len(s) && dLen < 12; i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			digits[dLen] = c
			dLen++
		}
	}
	if dLen != 12 {
		return strings.TrimSpace(s)
	}
	var sb strings.Builder
	sb.Grow(14)
	sb.Write(digits[0:4])
	sb.WriteByte(' ')
	sb.Write(digits[4:8])
	sb.WriteByte(' ')
	sb.Write(digits[8:12])
	return sb.String()
}

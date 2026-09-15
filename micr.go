package fintechin

import (
	"errors"
)

var (
	// ErrInvalidMICRLength indicates the MICR code does not have exactly 9 digits.
	ErrInvalidMICRLength = errors.New("micr: code must be exactly 9 digits")
	// ErrInvalidMICRFormat indicates the MICR code contains non-numeric characters.
	ErrInvalidMICRFormat = errors.New("micr: code must contain only numeric digits")
	// ErrInvalidMICRAllZeros indicates the MICR code is all zeros, which is invalid.
	ErrInvalidMICRAllZeros = errors.New("micr: code cannot be all zeros")
)

// ValidateMICR validates a 9-digit Magnetic Ink Character Recognition (MICR) code used for cheque clearing:
// - Format: exactly 9 numeric digits
// - Digits 1-3: City Code (aligned with first 3 digits of PIN code)
// - Digits 4-6: Bank Code
// - Digits 7-9: Branch Code
// - Cannot be all zeros
// Zero heap allocations.
func ValidateMICR(s string) error {
	if len(s) != 9 {
		return ErrInvalidMICRLength
	}

	allZero := true
	for i := 0; i < 9; i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return ErrInvalidMICRFormat
		}
		if c != '0' {
			allZero = false
		}
	}

	if allZero {
		return ErrInvalidMICRAllZeros
	}

	return nil
}

// IsValidMICR reports whether the given string is a valid MICR code.
func IsValidMICR(s string) bool {
	return ValidateMICR(s) == nil
}

// MICRCityCode extracts the 3-digit city code prefix from a MICR code.
func MICRCityCode(micr string) string {
	if len(micr) >= 3 {
		return micr[:3]
	}
	return ""
}

// MICRBankCode extracts the 3-digit bank code from a MICR code.
func MICRBankCode(micr string) string {
	if len(micr) >= 6 {
		return micr[3:6]
	}
	return ""
}

// MICRBranchCode extracts the 3-digit branch code suffix from a MICR code.
func MICRBranchCode(micr string) string {
	if len(micr) == 9 {
		return micr[6:9]
	}
	return ""
}

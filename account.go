package fintechin

import (
	"errors"
	"strings"
)

var (
	// ErrInvalidBankAccountLength indicates the bank account number length is not between 9 and 18 digits.
	ErrInvalidBankAccountLength = errors.New("account: bank account number must be between 9 and 18 digits")
	// ErrInvalidBankAccountFormat indicates the account number contains non-digit characters.
	ErrInvalidBankAccountFormat = errors.New("account: bank account number must contain only numeric digits")
	// ErrInvalidBankAccountAllZeros indicates the account number consists entirely of zeros.
	ErrInvalidBankAccountAllZeros = errors.New("account: bank account number cannot be all zeros")
)

// ValidateBankAccount validates an Indian bank account number according to RBI/banking conventions:
// - Length must be between 9 and 18 numeric digits
// - Must contain only numeric digits
// - Cannot be all zeros
// Zero heap allocations.
func ValidateBankAccount(accountNumber string) error {
	trimmed := strings.TrimSpace(accountNumber)
	l := len(trimmed)
	if l < 9 || l > 18 {
		return ErrInvalidBankAccountLength
	}

	allZero := true
	for i := 0; i < l; i++ {
		c := trimmed[i]
		if c < '0' || c > '9' {
			return ErrInvalidBankAccountFormat
		}
		if c != '0' {
			allZero = false
		}
	}

	if allZero {
		return ErrInvalidBankAccountAllZeros
	}

	return nil
}

// IsValidBankAccount reports whether the given string is a valid bank account number.
func IsValidBankAccount(accountNumber string) bool {
	return ValidateBankAccount(accountNumber) == nil
}

// MaskBankAccount masks an Indian bank account number, hiding all leading digits and revealing only the last 4:
// e.g. "123456789012" -> "XXXXXXXX9012".
func MaskBankAccount(acc string) string {
	trimmed := strings.TrimSpace(acc)
	l := len(trimmed)
	if l <= 4 {
		return strings.Repeat("X", l)
	}
	return strings.Repeat("X", l-4) + trimmed[l-4:]
}

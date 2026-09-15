package fintechin

import (
	"errors"
)

var (
	// ErrInvalidIFSCLength indicates the IFSC string does not have length 11.
	ErrInvalidIFSCLength = errors.New("ifsc: code must be exactly 11 characters")
	// ErrInvalidIFSCBankCode indicates the first 4 characters are not alphabetic.
	ErrInvalidIFSCBankCode = errors.New("ifsc: first 4 characters must be uppercase alphabetic bank code")
	// ErrInvalidIFSCFifthChar indicates the 5th character is not '0'.
	ErrInvalidIFSCFifthChar = errors.New("ifsc: 5th character must be '0'")
	// ErrInvalidIFSCBranchCode indicates the last 6 characters are not alphanumeric.
	ErrInvalidIFSCBranchCode = errors.New("ifsc: last 6 characters must be alphanumeric branch code")
)

// knownBanks maps 4-letter IFSC bank prefixes to their official bank names.
var knownBanks = map[string]string{
	"SBIN": "State Bank of India",
	"HDFC": "HDFC Bank",
	"ICIC": "ICICI Bank",
	"UTIB": "Axis Bank",
	"KKBK": "Kotak Mahindra Bank",
	"PUNB": "Punjab National Bank",
	"BARB": "Bank of Baroda",
	"CNRB": "Canara Bank",
	"UBIN": "Union Bank of India",
	"INDB": "IndusInd Bank",
	"IDFB": "IDFC FIRST Bank",
	"YESB": "Yes Bank",
	"IOBA": "Indian Overseas Bank",
	"IDIB": "Indian Bank",
	"CBIN": "Central Bank of India",
	"BKID": "Bank of India",
	"MAHB": "Bank of Maharashtra",
	"PSIB": "Punjab & Sind Bank",
	"UCOB": "UCO Bank",
	"FDRL": "Federal Bank",
	"SCBL": "Standard Chartered Bank",
	"HSBC": "HSBC",
	"CITI": "Citibank",
	"RATN": "RBL Bank",
	"KVBL": "Karur Vysya Bank",
	"SIBL": "South Indian Bank",
	"CSBK": "CSB Bank",
	"DCBL": "DCB Bank",
	"TMBL": "Tamilnad Mercantile Bank",
	"BAND": "Bandhan Bank",
	"AUBL": "AU Small Finance Bank",
	"ESFB": "Equitas Small Finance Bank",
	"UJJV": "Ujjivan Small Finance Bank",
	"AIRP": "Airtel Payments Bank",
	"PYTM": "Paytm Payments Bank",
	"IPOS": "India Post Payments Bank",
	"JAKA": "Jammu & Kashmir Bank",
	"SVCB": "SVC Co-operative Bank",
	"DEUT": "Deutsche Bank",
	"DBSS": "DBS Bank",
	"IBKL": "IDBI Bank",
}

// ValidateIFSC validates an Indian Financial System Code (IFSC) according to RBI guidelines:
// - Format: exactly 11 characters
// - Characters 1-4: 4 uppercase letters representing bank
// - Character 5: Reserved digit '0'
// - Characters 6-11: 6 alphanumeric characters representing branch
// Zero heap allocations.
func ValidateIFSC(s string) error {
	if len(s) != 11 {
		return ErrInvalidIFSCLength
	}

	// First 4 characters must be uppercase alphabetic [A-Z]
	for i := 0; i < 4; i++ {
		c := s[i]
		if c < 'A' || c > 'Z' {
			return ErrInvalidIFSCBankCode
		}
	}

	// 5th character must be '0'
	if s[4] != '0' {
		return ErrInvalidIFSCFifthChar
	}

	// Last 6 characters must be alphanumeric [0-9A-Z]
	for i := 5; i < 11; i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'A' || c > 'Z') {
			return ErrInvalidIFSCBranchCode
		}
	}

	return nil
}

// IsValidIFSC reports whether the given string is a valid IFSC code.
func IsValidIFSC(s string) bool {
	return ValidateIFSC(s) == nil
}

// BankCode extracts the 4-letter bank code prefix from an IFSC code.
func BankCode(ifsc string) string {
	if len(ifsc) >= 4 {
		return ifsc[:4]
	}
	return ""
}

// BranchCode extracts the 6-character branch code suffix from an IFSC code.
func BranchCode(ifsc string) string {
	if len(ifsc) == 11 {
		return ifsc[5:]
	}
	return ""
}

// BankNameFromIFSC returns the full bank name for major Indian banks, or an empty string if unrecognized.
func BankNameFromIFSC(ifsc string) string {
	code := BankCode(ifsc)
	if code == "" {
		return ""
	}
	return knownBanks[code]
}

package fintechin

import (
	"errors"
)

var (
	// ErrInvalidGSTINLength indicates the GSTIN does not have exactly 15 characters.
	ErrInvalidGSTINLength = errors.New("gstin: number must be exactly 15 characters")
	// ErrInvalidGSTINFormat indicates the GSTIN contains invalid characters or does not meet structural rules.
	ErrInvalidGSTINFormat = errors.New("gstin: invalid format, expected 2 digits state + 10 PAN + 1 entity + 'Z' + 1 check character")
	// ErrInvalidStateCode indicates the 2-digit state code is not recognized by GSTN.
	ErrInvalidStateCode = errors.New("gstin: invalid state code")
	// ErrInvalidGSTINChecksum indicates the 15th character does not match the official Mod-36 checksum.
	ErrInvalidGSTINChecksum = errors.New("gstin: invalid checksum character")
)

// gstStateCodes maps official 2-digit Indian State/UT GST codes (01-38, 97, 99).
var gstStateCodes = map[string]string{
	"01": "Jammu and Kashmir",
	"02": "Himachal Pradesh",
	"03": "Punjab",
	"04": "Chandigarh",
	"05": "Uttarakhand",
	"06": "Haryana",
	"07": "Delhi",
	"08": "Rajasthan",
	"09": "Uttar Pradesh",
	"10": "Bihar",
	"11": "Sikkim",
	"12": "Arunachal Pradesh",
	"13": "Nagaland",
	"14": "Manipur",
	"15": "Mizoram",
	"16": "Tripura",
	"17": "Meghalaya",
	"18": "Assam",
	"19": "West Bengal",
	"20": "Jharkhand",
	"21": "Odisha",
	"22": "Chhattisgarh",
	"23": "Madhya Pradesh",
	"24": "Gujarat",
	"25": "Daman and Diu",
	"26": "Dadra and Nagar Haveli and Daman and Diu",
	"27": "Maharashtra",
	"28": "Andhra Pradesh (Old)",
	"29": "Karnataka",
	"30": "Goa",
	"31": "Lakshadweep",
	"32": "Kerala",
	"33": "Tamil Nadu",
	"34": "Puducherry",
	"35": "Andaman and Nicobar Islands",
	"36": "Telangana",
	"37": "Andhra Pradesh",
	"38": "Ladakh",
	"97": "Other Territory",
	"99": "Centre Jurisdiction",
}

// StateName returns the Indian State or Union Territory name for a 2-digit GST state code.
func StateName(stateCode string) (string, error) {
	name, ok := gstStateCodes[stateCode]
	if !ok {
		return "", ErrInvalidStateCode
	}
	return name, nil
}

// StateCode extracts the 2-digit state code prefix from a GSTIN.
func StateCode(gstin string) string {
	if len(gstin) >= 2 {
		return gstin[:2]
	}
	return ""
}

// ExtractPAN extracts the embedded 10-character PAN from a GSTIN and validates it.
// Returns the PAN if valid, or an empty string if the GSTIN or embedded PAN is invalid.
func ExtractPAN(gstin string) string {
	if len(gstin) != 15 {
		return ""
	}
	pan := gstin[2:12]
	if err := ValidatePAN(pan); err != nil {
		return ""
	}
	return pan
}

// CalculateGSTINChecksum calculates the official GSTN Mod-36 checksum character for the first 14 characters.
// Employs a weighted Luhn Mod-36 algorithm with alternating factors 1 and 2.
// Zero heap allocations.
func CalculateGSTINChecksum(first14 string) (byte, error) {
	if len(first14) != 14 {
		return 0, ErrInvalidGSTINLength
	}

	sum := 0
	for i := 0; i < 14; i++ {
		c := first14[i]
		var val int
		switch {
		case c >= '0' && c <= '9':
			val = int(c - '0')
		case c >= 'A' && c <= 'Z':
			val = int(c - 'A' + 10)
		default:
			return 0, ErrInvalidGSTINFormat
		}

		factor := 1
		if i%2 != 0 {
			factor = 2
		}

		prod := val * factor
		term := (prod / 36) + (prod % 36)
		sum += term
	}

	checkVal := (36 - (sum % 36)) % 36
	if checkVal < 10 {
		return byte('0' + checkVal), nil
	}
	return byte('A' + (checkVal - 10)), nil
}

// ValidateGSTIN validates a 15-character Goods and Services Tax Identification Number (GSTIN).
// Checks:
// 1. Length must be exactly 15 characters
// 2. First 2 characters must be a recognized GST State Code (01-38, 97, 99)
// 3. Characters 3-12 must be a valid PAN
// 4. Character 13 must be alphanumeric (entity registration number)
// 5. Character 14 must be 'Z'
// 6. Character 15 must match the official GSTN Mod-36 checksum
// Zero heap allocations.
func ValidateGSTIN(s string) error {
	if len(s) != 15 {
		return ErrInvalidGSTINLength
	}

	// 1. State code check
	stateCode := s[:2]
	if _, ok := gstStateCodes[stateCode]; !ok {
		return ErrInvalidStateCode
	}

	// 2. PAN check
	pan := s[2:12]
	if err := ValidatePAN(pan); err != nil {
		return err
	}

	// 3. Entity code check (alphanumeric 0-9, A-Z)
	c13 := s[12]
	if (c13 < '0' || c13 > '9') && (c13 < 'A' || c13 > 'Z') {
		return ErrInvalidGSTINFormat
	}

	// 4. 14th character must be 'Z'
	if s[13] != 'Z' {
		return ErrInvalidGSTINFormat
	}

	// 5. 15th character checksum verification
	expectedCheck, err := CalculateGSTINChecksum(s[:14])
	if err != nil {
		return err
	}

	if s[14] != expectedCheck {
		return ErrInvalidGSTINChecksum
	}

	return nil
}

// IsValidGSTIN reports whether the given string is a valid GSTIN.
func IsValidGSTIN(s string) bool {
	return ValidateGSTIN(s) == nil
}

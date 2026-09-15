package fintechin

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	// ErrInvalidPANLength indicates the PAN string does not have length 10.
	ErrInvalidPANLength = errors.New("pan: number must be exactly 10 characters")
	// ErrInvalidPANFormat indicates the PAN does not conform to [A-Z]{5}[0-9]{4}[A-Z]{1}.
	ErrInvalidPANFormat = errors.New("pan: invalid format, must match [A-Z]{5}[0-9]{4}[A-Z]{1}")
	// ErrInvalidPANEntityType indicates the 4th character of PAN is not an acknowledged entity code.
	ErrInvalidPANEntityType = errors.New("pan: invalid entity type character at 4th position")
)

// PANEntityType represents the taxpayer entity type classified by the 4th character of the PAN.
type PANEntityType string

const (
	EntityTypeIndividual                PANEntityType = "Individual"
	EntityTypeCompany                   PANEntityType = "Company"
	EntityTypeHUF                       PANEntityType = "HUF"
	EntityTypeFirm                      PANEntityType = "Firm/LLP"
	EntityTypeAOP                       PANEntityType = "AOP"
	EntityTypeTrust                     PANEntityType = "Trust"
	EntityTypeBOI                       PANEntityType = "BOI"
	EntityTypeLocalAuthority            PANEntityType = "Local Authority"
	EntityTypeArtificialJuridicalPerson PANEntityType = "Artificial Juridical Person"
	EntityTypeGovernmentAgency          PANEntityType = "Government Agency"

	// Convenient aliases
	EntityIndividual                = EntityTypeIndividual
	EntityCompany                   = EntityTypeCompany
	EntityHUF                       = EntityTypeHUF
	EntityFirm                      = EntityTypeFirm
	EntityAOP                       = EntityTypeAOP
	EntityTrust                     = EntityTypeTrust
	EntityBOI                       = EntityTypeBOI
	EntityLocalAuthority            = EntityTypeLocalAuthority
	EntityArtificialJuridicalPerson = EntityTypeArtificialJuridicalPerson
	EntityGovernmentAgency          = EntityTypeGovernmentAgency
)

// ValidatePAN validates an Indian Permanent Account Number (PAN) according to Income Tax Department rules.
// Must match format: [A-Z]{5}[0-9]{4}[A-Z]{1}.
// Zero heap allocations.
func ValidatePAN(s string) error {
	if len(s) != 10 {
		return ErrInvalidPANLength
	}

	// First 5 characters must be uppercase alphabetic [A-Z]
	for i := 0; i < 5; i++ {
		c := s[i]
		if c < 'A' || c > 'Z' {
			return ErrInvalidPANFormat
		}
	}

	// Next 4 characters must be numeric digits [0-9]
	for i := 5; i < 9; i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return ErrInvalidPANFormat
		}
	}

	// Last character must be uppercase alphabetic [A-Z]
	if s[9] < 'A' || s[9] > 'Z' {
		return ErrInvalidPANFormat
	}

	return nil
}

// IsValidPAN reports whether the given string is a valid PAN.
func IsValidPAN(s string) bool {
	return ValidatePAN(s) == nil
}

// EntityType decodes the entity type from the 4th character of the PAN.
// 'P' -> Individual
// 'C' -> Company
// 'H' -> HUF
// 'F' -> Firm/LLP
// 'A' -> AOP
// 'T' -> Trust
// 'B' -> BOI
// 'L' -> Local Authority
// 'J' -> Artificial Juridical Person
// 'G' -> Government Agency
func EntityType(pan string) (PANEntityType, error) {
	if err := ValidatePAN(pan); err != nil {
		return "", err
	}

	switch pan[3] {
	case 'P':
		return EntityTypeIndividual, nil
	case 'C':
		return EntityTypeCompany, nil
	case 'H':
		return EntityTypeHUF, nil
	case 'F':
		return EntityTypeFirm, nil
	case 'A':
		return EntityTypeAOP, nil
	case 'T':
		return EntityTypeTrust, nil
	case 'B':
		return EntityTypeBOI, nil
	case 'L':
		return EntityTypeLocalAuthority, nil
	case 'J':
		return EntityTypeArtificialJuridicalPerson, nil
	case 'G':
		return EntityTypeGovernmentAgency, nil
	default:
		return "", ErrInvalidPANEntityType
	}
}

// EntityTypeFromPAN is an alias for EntityType.
func EntityTypeFromPAN(pan string) (PANEntityType, error) {
	return EntityType(pan)
}

// MatchesSurname checks if the 5th character of the PAN matches the first letter of the surname / entity name.
// Matches case-insensitively.
func MatchesSurname(pan, surname string) bool {
	if err := ValidatePAN(pan); err != nil {
		return false
	}
	trimmed := strings.TrimSpace(surname)
	if trimmed == "" {
		return false
	}

	r, _ := utf8.DecodeRuneInString(trimmed)
	return unicode.ToUpper(r) == rune(pan[4])
}

// MaskPAN returns the masked PAN representation "XXXXX1234X" preserving the 4 numeric digits.
func MaskPAN(pan string) string {
	if len(pan) == 10 {
		return "XXXXX" + pan[5:9] + "X"
	}
	// Fallback for non-standard lengths
	if len(pan) < 5 {
		return strings.Repeat("X", len(pan))
	}
	return "XXXXX" + strings.Repeat("X", len(pan)-5)
}

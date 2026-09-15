package fintechin

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// FormatINR formats integer paise into Indian Numbering Format (e.g. "12,34,567.89" or "-50,000.00").
func FormatINR(paise int64) string {
	sign := ""
	p := paise
	if p < 0 {
		sign = "-"
		p = -p
	}

	rupees := p / 100
	paisePart := p % 100

	rupeesStr := strconv.FormatInt(rupees, 10)
	var formattedRupees string

	if len(rupeesStr) <= 3 {
		formattedRupees = rupeesStr
	} else {
		last3 := rupeesStr[len(rupeesStr)-3:]
		remaining := rupeesStr[:len(rupeesStr)-3]

		var chunks []string
		for len(remaining) > 2 {
			chunks = append([]string{remaining[len(remaining)-2:]}, chunks...)
			remaining = remaining[:len(remaining)-2]
		}
		if remaining != "" {
			chunks = append([]string{remaining}, chunks...)
		}
		formattedRupees = strings.Join(chunks, ",") + "," + last3
	}

	return fmt.Sprintf("%s%s.%02d", sign, formattedRupees, paisePart)
}

// FormatINRSymbol formats integer paise with the Indian Rupee symbol (e.g. "₹12,34,567.89", "-₹50,000.00").
func FormatINRSymbol(paise int64) string {
	if paise < 0 {
		return "-₹" + FormatINR(-paise)
	}
	return "₹" + FormatINR(paise)
}

// ParseINR parses an Indian Rupee string representation into Money.
// Supports strings with or without currency symbols ("₹", "Rs.", "Rs", "INR"),
// commas, negative signs, and decimals (e.g. "₹12,34,567.89", "12,34,567.89", "1234567.89", "-₹50,000.00").
func ParseINR(s string) (Money, error) {
	cleaned, isNegative, err := stripCurrencyAndSigns(s)
	if err != nil {
		return Money{}, err
	}

	// Remove all commas
	cleaned = strings.ReplaceAll(cleaned, ",", "")

	// Split integer rupees and fractional paise
	parts := strings.Split(cleaned, ".")
	if len(parts) > 2 {
		return Money{}, fmt.Errorf("fintechin: multiple decimal points in %q", s)
	}

	var rupees int64
	if parts[0] != "" {
		for _, r := range parts[0] {
			if !unicode.IsDigit(r) {
				return Money{}, fmt.Errorf("fintechin: invalid character in amount %q", s)
			}
		}
		r, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return Money{}, fmt.Errorf("fintechin: overflow or invalid rupees in %q: %w", s, err)
		}
		rupees = r
	}

	var paise int64
	if len(parts) == 2 && parts[1] != "" {
		p, err := parsePaiseFraction(parts[1], s)
		if err != nil {
			return Money{}, err
		}
		paise = p
	}

	totalPaise := rupees*100 + paise
	if isNegative {
		totalPaise = -totalPaise
	}

	return NewMoney(totalPaise), nil
}

func stripCurrencyAndSigns(raw string) (cleaned string, isNeg bool, err error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", false, errors.New("fintechin: empty amount string")
	}

	// Check for accounting parentheses e.g. "(1,234.50)"
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		isNeg = true
		s = strings.TrimSpace(s[1 : len(s)-1])
	}

	cleaned = s
	hasSign := false
	hasCurrency := false
	for {
		trimmed := strings.TrimSpace(cleaned)
		if strings.HasPrefix(trimmed, "-") {
			if hasSign {
				return "", false, fmt.Errorf("fintechin: multiple signs in %q", raw)
			}
			hasSign = true
			isNeg = true
			cleaned = trimmed[1:]
			continue
		}
		if strings.HasPrefix(trimmed, "+") {
			if hasSign {
				return "", false, fmt.Errorf("fintechin: multiple signs in %q", raw)
			}
			hasSign = true
			cleaned = trimmed[1:]
			continue
		}
		if strings.HasPrefix(trimmed, "₹") {
			if hasCurrency {
				return "", false, fmt.Errorf("fintechin: duplicate currency symbol in %q", raw)
			}
			hasCurrency = true
			cleaned = trimmed[len("₹"):]
			continue
		}
		upper := strings.ToUpper(trimmed)
		if strings.HasPrefix(upper, "RS.") {
			if hasCurrency {
				return "", false, fmt.Errorf("fintechin: duplicate currency symbol in %q", raw)
			}
			hasCurrency = true
			cleaned = trimmed[3:]
			continue
		}
		if strings.HasPrefix(upper, "RS") {
			if hasCurrency {
				return "", false, fmt.Errorf("fintechin: duplicate currency symbol in %q", raw)
			}
			hasCurrency = true
			cleaned = trimmed[2:]
			continue
		}
		if strings.HasPrefix(upper, "INR") {
			if hasCurrency {
				return "", false, fmt.Errorf("fintechin: duplicate currency symbol in %q", raw)
			}
			hasCurrency = true
			cleaned = trimmed[3:]
			continue
		}
		break
	}

	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return "", false, fmt.Errorf("fintechin: invalid amount string: %q", raw)
	}
	return cleaned, isNeg, nil
}

func parsePaiseFraction(paiseStr string, raw string) (int64, error) {
	for _, r := range paiseStr {
		if !unicode.IsDigit(r) {
			return 0, fmt.Errorf("fintechin: invalid character in paise %q", raw)
		}
	}
	switch len(paiseStr) {
	case 1:
		return int64(paiseStr[0]-'0') * 10, nil
	case 2:
		p, err := strconv.ParseInt(paiseStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("fintechin: invalid paise in %q: %w", raw, err)
		}
		return p, nil
	default:
		// For fractional paise more than 2 decimals, round to nearest paisa
		pFloat, err := strconv.ParseFloat("0."+paiseStr, 64)
		if err != nil {
			return 0, fmt.Errorf("fintechin: invalid paise fractional %q: %w", raw, err)
		}
		return int64(pFloat*100.0 + 0.5), nil
	}
}

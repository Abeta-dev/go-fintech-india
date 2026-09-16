package fintechin

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// AppendINR appends the Indian Numbering Format of paise to dst and returns the extended buffer.
// Performs zero heap allocations when dst has sufficient capacity.
func AppendINR(dst []byte, paise int64) []byte {
	var buf [48]byte
	i := len(buf)

	var u uint64
	neg := paise < 0
	if neg {
		u = uint64(-paise)
	} else {
		u = uint64(paise)
	}

	paisePart := u % 100
	rupees := u / 100

	i--
	buf[i] = byte('0' + paisePart%10)
	i--
	buf[i] = byte('0' + paisePart/10)
	i--
	buf[i] = '.'

	if rupees == 0 {
		i--
		buf[i] = '0'
	} else {
		digitCount := 0
		groupLimit := 3
		for rupees > 0 {
			if digitCount == groupLimit {
				i--
				buf[i] = ','
				digitCount = 0
				groupLimit = 2
			}
			i--
			buf[i] = byte('0' + rupees%10)
			rupees /= 10
			digitCount++
		}
	}

	if neg {
		i--
		buf[i] = '-'
	}

	return append(dst, buf[i:]...)
}

// FormatINR formats integer paise into Indian Numbering Format (e.g. "12,34,567.89" or "-50,000.00").
// Optimized to perform only a single heap allocation for the returned string.
func FormatINR(paise int64) string {
	var buf [48]byte
	res := AppendINR(buf[:0], paise)
	return string(res)
}

// AppendINRSymbol appends the Indian Numbering Format with Rupee symbol ("₹") to dst.
// Performs zero heap allocations when dst has sufficient capacity.
func AppendINRSymbol(dst []byte, paise int64) []byte {
	var buf [48]byte
	i := len(buf)

	var u uint64
	neg := paise < 0
	if neg {
		u = uint64(-paise)
	} else {
		u = uint64(paise)
	}

	paisePart := u % 100
	rupees := u / 100

	i--
	buf[i] = byte('0' + paisePart%10)
	i--
	buf[i] = byte('0' + paisePart/10)
	i--
	buf[i] = '.'

	if rupees == 0 {
		i--
		buf[i] = '0'
	} else {
		digitCount := 0
		groupLimit := 3
		for rupees > 0 {
			if digitCount == groupLimit {
				i--
				buf[i] = ','
				digitCount = 0
				groupLimit = 2
			}
			i--
			buf[i] = byte('0' + rupees%10)
			rupees /= 10
			digitCount++
		}
	}

	// Prepend Rupee symbol "₹" (UTF-8: \u20b9, 3 bytes: 0xE2, 0x82, 0xB9)
	i -= 3
	copy(buf[i:i+3], "\u20b9")

	if neg {
		i--
		buf[i] = '-'
	}

	return append(dst, buf[i:]...)
}

// FormatINRSymbol formats integer paise with the Indian Rupee symbol (e.g. "₹12,34,567.89", "-₹50,000.00").
// Optimized to perform only a single heap allocation for the returned string.
func FormatINRSymbol(paise int64) string {
	var buf [48]byte
	res := AppendINRSymbol(buf[:0], paise)
	return string(res)
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

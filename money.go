package fintechin

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
)

// Money represents a monetary amount in Indian Rupees stored as integer paise.
// Exact integer arithmetic is guaranteed: 1 INR = 100 paise.
type Money struct {
	paise int64
}

// NewMoney creates a new Money value from raw integer paise.
func NewMoney(paise int64) Money {
	return Money{paise: paise}
}

// NewMoneyFromRupees creates a new Money value from whole rupees.
func NewMoneyFromRupees(rupees int64) Money {
	return Money{paise: rupees * 100}
}

// NewMoneyFromFloat creates a new Money value from a float amount in rupees,
// rounding to the nearest paisa (half away from zero).
func NewMoneyFromFloat(amount float64) Money {
	return Money{paise: int64(math.Round(amount * 100.0))}
}

// Paise returns the amount in paise.
func (m Money) Paise() int64 {
	return m.paise
}

// Rupees returns the whole rupees component (truncated toward zero).
func (m Money) Rupees() int64 {
	return m.paise / 100
}

// Float64 returns the monetary amount as float64 rupees.
func (m Money) Float64() float64 {
	return float64(m.paise) / 100.0
}

// IsZero reports whether the amount is zero.
func (m Money) IsZero() bool {
	return m.paise == 0
}

// IsPositive reports whether the amount is greater than zero.
func (m Money) IsPositive() bool {
	return m.paise > 0
}

// IsNegative reports whether the amount is less than zero.
func (m Money) IsNegative() bool {
	return m.paise < 0
}

// Abs returns the absolute value of Money.
func (m Money) Abs() Money {
	if m.paise < 0 {
		return Money{paise: -m.paise}
	}
	return m
}

// Negate returns the negated Money value.
func (m Money) Negate() Money {
	return Money{paise: -m.paise}
}

// Add returns m + other.
func (m Money) Add(other Money) Money {
	return Money{paise: m.paise + other.paise}
}

// Sub returns m - other.
func (m Money) Sub(other Money) Money {
	return Money{paise: m.paise - other.paise}
}

// Mul returns m * factor.
func (m Money) Mul(factor int64) Money {
	return Money{paise: m.paise * factor}
}

// Split splits m into n parts as evenly as possible, distributing remainder
// paise to the earlier elements so that the sum of parts always equals m.
func (m Money) Split(n int) ([]Money, error) {
	if n <= 0 {
		return nil, errors.New("fintechin: split count must be positive")
	}

	base := m.paise / int64(n)
	rem := m.paise % int64(n)

	parts := make([]Money, n)
	for i := 0; i < n; i++ {
		parts[i] = Money{paise: base}
	}

	if rem > 0 {
		for i := 0; i < int(rem); i++ {
			parts[i].paise++
		}
	} else if rem < 0 {
		for i := 0; i < int(-rem); i++ {
			parts[i].paise--
		}
	}

	return parts, nil
}

// Allocate divides m across proportional weights (ratios) without any remainder loss.
// Remainder paise are distributed to elements with the largest fractional parts (Hare-Niemeyer method).
func (m Money) Allocate(ratios ...int) ([]Money, error) {
	if len(ratios) == 0 {
		return nil, errors.New("fintechin: at least one ratio required")
	}

	var totalWeight int64
	for _, r := range ratios {
		if r < 0 {
			return nil, errors.New("fintechin: ratio cannot be negative")
		}
		totalWeight += int64(r)
	}
	if totalWeight <= 0 {
		return nil, errors.New("fintechin: sum of ratios must be greater than zero")
	}

	n := len(ratios)
	results := make([]Money, n)
	var allocatedSum int64

	type remInfo struct {
		index int
		rem   int64
	}
	remainders := make([]remInfo, n)

	for i, r := range ratios {
		product := m.paise * int64(r)
		base := product / totalWeight
		rem := product % totalWeight
		if rem < 0 {
			rem = -rem
		}
		results[i] = Money{paise: base}
		allocatedSum += base
		remainders[i] = remInfo{index: i, rem: rem}
	}

	leftover := m.paise - allocatedSum
	if leftover != 0 {
		// Sort by remainder descending, preserve order on tie
		sort.SliceStable(remainders, func(i, j int) bool {
			return remainders[i].rem > remainders[j].rem
		})

		step := int64(1)
		count := int(leftover)
		if leftover < 0 {
			step = -1
			count = int(-leftover)
		}

		for i := 0; i < count; i++ {
			idx := remainders[i%n].index
			results[idx].paise += step
		}
	}

	return results, nil
}

// String returns formatted INR with two decimal places (e.g. "123.45" or "-123.45").
func (m Money) String() string {
	sign := ""
	p := m.paise
	if p < 0 {
		sign = "-"
		p = -p
	}
	return fmt.Sprintf("%s%d.%02d", sign, p/100, p%100)
}

// Value implements database/sql/driver.Valuer, returning the integer paise.
func (m Money) Value() (driver.Value, error) {
	return m.paise, nil
}

// Scan implements database/sql.Scanner, supporting int64, int32, []byte, string, float64, int.
func (m *Money) Scan(src any) error {
	if src == nil {
		m.paise = 0
		return nil
	}

	switch v := src.(type) {
	case int64:
		m.paise = v
		return nil
	case int32:
		m.paise = int64(v)
		return nil
	case int:
		m.paise = int64(v)
		return nil
	case float64:
		*m = NewMoneyFromFloat(v)
		return nil
	case []byte:
		parsed, err := ParseINR(string(v))
		if err != nil {
			return fmt.Errorf("fintechin: failed to scan Money from []byte: %w", err)
		}
		*m = parsed
		return nil
	case string:
		parsed, err := ParseINR(v)
		if err != nil {
			return fmt.Errorf("fintechin: failed to scan Money from string: %w", err)
		}
		*m = parsed
		return nil
	default:
		return fmt.Errorf("fintechin: cannot scan %T into Money", src)
	}
}

// MarshalJSON implements json.Marshaler, emitting {"paise":12345,"inr":"123.45"}.
func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Paise int64  `json:"paise"`
		INR   string `json:"inr"`
	}{
		Paise: m.paise,
		INR:   m.String(),
	})
}

// UnmarshalJSON implements json.Unmarshaler, supporting {"paise":...}, raw integer paise, float rupees, or string.
func (m *Money) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		m.paise = 0
		return nil
	}

	// 1. JSON Object: {"paise": 12345, "inr": "123.45"}
	if data[0] == '{' {
		var obj struct {
			Paise *int64          `json:"paise"`
			INR   json.RawMessage `json:"inr"`
		}
		if err := json.Unmarshal(data, &obj); err != nil {
			return fmt.Errorf("fintechin: invalid Money json object: %w", err)
		}
		if obj.Paise != nil {
			m.paise = *obj.Paise
			return nil
		}
		if len(obj.INR) > 0 {
			var inrStr string
			if err := json.Unmarshal(obj.INR, &inrStr); err == nil {
				parsed, err := ParseINR(inrStr)
				if err != nil {
					return err
				}
				*m = parsed
				return nil
			}
			var inrFloat float64
			if err := json.Unmarshal(obj.INR, &inrFloat); err == nil {
				*m = NewMoneyFromFloat(inrFloat)
				return nil
			}
		}
		return errors.New("fintechin: money json object missing 'paise' and 'inr' fields")
	}

	// 2. JSON String: "123.45", "₹12,34,567.89", etc.
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("fintechin: invalid Money json string: %w", err)
		}
		parsed, err := ParseINR(s)
		if err != nil {
			return err
		}
		*m = parsed
		return nil
	}

	// 3. Raw number: float rupees (contains '.') or integer paise
	if bytes.ContainsRune(data, '.') {
		var f float64
		if err := json.Unmarshal(data, &f); err != nil {
			return fmt.Errorf("fintechin: invalid Money float json: %w", err)
		}
		*m = NewMoneyFromFloat(f)
		return nil
	}

	var i int64
	if err := json.Unmarshal(data, &i); err != nil {
		return fmt.Errorf("fintechin: invalid Money int json: %w", err)
	}
	m.paise = i
	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (m Money) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (m *Money) UnmarshalText(text []byte) error {
	parsed, err := ParseINR(string(text))
	if err != nil {
		return err
	}
	*m = parsed
	return nil
}

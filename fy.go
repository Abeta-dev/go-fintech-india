package fintechin

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FinancialYear represents an Indian Financial Year starting on April 1 and ending on March 31.
// The Year field represents the starting calendar year (e.g. 2024 for FY 2024-25).
type FinancialYear struct {
	Year int
}

// Label returns the Indian financial year label (e.g. "FY 2024-25").
func (fy FinancialYear) Label() string {
	nextTwoDigits := (fy.Year + 1) % 100
	return fmt.Sprintf("FY %d-%02d", fy.Year, nextTwoDigits)
}

// AssessmentYear returns the corresponding Indian assessment year label (e.g. "AY 2025-26").
func (fy FinancialYear) AssessmentYear() string {
	ayStart := fy.Year + 1
	ayNextTwoDigits := (ayStart + 1) % 100
	return fmt.Sprintf("AY %d-%02d", ayStart, ayNextTwoDigits)
}

// StartDate returns April 1 00:00:00 UTC of the financial year.
func (fy FinancialYear) StartDate() time.Time {
	return time.Date(fy.Year, time.April, 1, 0, 0, 0, 0, time.UTC)
}

// EndDate returns March 31 00:00:00 UTC of the subsequent calendar year.
func (fy FinancialYear) EndDate() time.Time {
	return time.Date(fy.Year+1, time.March, 31, 0, 0, 0, 0, time.UTC)
}

// Quarter returns the Indian financial year quarter (1 to 4) for a given date:
// Q1: Apr-Jun, Q2: Jul-Sep, Q3: Oct-Dec, Q4: Jan-Mar.
func (fy FinancialYear) Quarter(t time.Time) int {
	return Quarter(t)
}

// Contains reports whether the given timestamp falls within this financial year.
func (fy FinancialYear) Contains(t time.Time) bool {
	date := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	start := fy.StartDate()
	end := fy.EndDate()
	return (date.Equal(start) || date.After(start)) && (date.Equal(end) || date.Before(end))
}

// AdvanceTaxDueDates returns the 4 statutory advance tax installments for this financial year.
func (fy FinancialYear) AdvanceTaxDueDates() []AdvanceTaxInstallment {
	return AdvanceTaxDueDates(fy)
}

// Quarter returns the Indian financial quarter (1 to 4) for a given date:
// Q1: Apr-Jun, Q2: Jul-Sep, Q3: Oct-Dec, Q4: Jan-Mar.
func Quarter(t time.Time) int {
	switch t.Month() {
	case time.April, time.May, time.June:
		return 1
	case time.July, time.August, time.September:
		return 2
	case time.October, time.November, time.December:
		return 3
	case time.January, time.February, time.March:
		return 4
	}
	return 0
}

// FYFromDate computes the Indian Financial Year for the given time.
func FYFromDate(t time.Time) FinancialYear {
	year := t.Year()
	if t.Month() < time.April {
		year--
	}
	return FinancialYear{Year: year}
}

// CurrentFY returns the current Indian Financial Year based on system time.
func CurrentFY() FinancialYear {
	return FYFromDate(time.Now())
}

// ParseFY parses a financial year string into a FinancialYear.
// Supported formats include "FY 2024-25", "FY2024-25", "2024-25", "2024-2025", "AY 2025-26", and "2024".
func ParseFY(s string) (FinancialYear, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return FinancialYear{}, errors.New("fintechin: empty financial year string")
	}

	upper := strings.ToUpper(s)
	isAY := false
	if strings.HasPrefix(upper, "FY") {
		s = strings.TrimSpace(s[2:])
	} else if strings.HasPrefix(upper, "AY") {
		isAY = true
		s = strings.TrimSpace(s[2:])
	}

	delimiter := ""
	if strings.Contains(s, "-") {
		delimiter = "-"
	} else if strings.Contains(s, "/") {
		delimiter = "/"
	}

	var startYear int
	if delimiter != "" {
		parts := strings.Split(s, delimiter)
		if len(parts) != 2 {
			return FinancialYear{}, fmt.Errorf("fintechin: invalid financial year format: %q", s)
		}
		y1, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return FinancialYear{}, fmt.Errorf("fintechin: invalid start year: %w", err)
		}
		y2Str := strings.TrimSpace(parts[1])
		y2, err := strconv.Atoi(y2Str)
		if err != nil {
			return FinancialYear{}, fmt.Errorf("fintechin: invalid end year: %w", err)
		}

		switch len(y2Str) {
		case 2:
			expectedEnd := (y1 + 1) % 100
			if y2 != expectedEnd {
				return FinancialYear{}, fmt.Errorf("fintechin: inconsistent financial year %d-%02d (expected %d-%02d)", y1, y2, y1, expectedEnd)
			}
		case 4:
			if y2 != y1+1 {
				return FinancialYear{}, fmt.Errorf("fintechin: inconsistent financial year %d-%d (expected %d-%d)", y1, y2, y1, y1+1)
			}
		default:
			return FinancialYear{}, fmt.Errorf("fintechin: invalid end year digits: %q", y2Str)
		}
		startYear = y1
	} else {
		y, err := strconv.Atoi(s)
		if err != nil {
			return FinancialYear{}, fmt.Errorf("fintechin: invalid year format: %w", err)
		}
		startYear = y
	}

	if isAY {
		startYear--
	}

	return FinancialYear{Year: startYear}, nil
}

// AdvanceTaxInstallment represents an Indian statutory advance tax installment.
type AdvanceTaxInstallment struct {
	Quarter       int       // 1, 2, 3, or 4
	DueDate       time.Time // Statutory payment deadline
	PercentageCum int       // Cumulative percentage due (15, 45, 75, 100)
}

// AdvanceTaxDueDates returns the 4 statutory advance tax installments for corporate/non-corporate taxpayers
// under Section 211 of the Income Tax Act:
// - June 15: 15%
// - September 15: 45%
// - December 15: 75%
// - March 15: 100%
func AdvanceTaxDueDates(fy FinancialYear) []AdvanceTaxInstallment {
	return []AdvanceTaxInstallment{
		{
			Quarter:       1,
			DueDate:       time.Date(fy.Year, time.June, 15, 0, 0, 0, 0, time.UTC),
			PercentageCum: 15,
		},
		{
			Quarter:       2,
			DueDate:       time.Date(fy.Year, time.September, 15, 0, 0, 0, 0, time.UTC),
			PercentageCum: 45,
		},
		{
			Quarter:       3,
			DueDate:       time.Date(fy.Year, time.December, 15, 0, 0, 0, 0, time.UTC),
			PercentageCum: 75,
		},
		{
			Quarter:       4,
			DueDate:       time.Date(fy.Year+1, time.March, 15, 0, 0, 0, 0, time.UTC),
			PercentageCum: 100,
		},
	}
}

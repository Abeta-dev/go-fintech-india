package fintechin

import (
	"testing"
	"time"
)

func TestFinancialYearProperties(t *testing.T) {
	fy := FinancialYear{Year: 2024}

	if got := fy.Label(); got != "FY 2024-25" {
		t.Fatalf("fy.Label() = %q, want 'FY 2024-25'", got)
	}
	if got := fy.AssessmentYear(); got != "AY 2025-26" {
		t.Fatalf("fy.AssessmentYear() = %q, want 'AY 2025-26'", got)
	}

	start := fy.StartDate()
	if start.Year() != 2024 || start.Month() != time.April || start.Day() != 1 {
		t.Fatalf("expected StartDate 2024-04-01, got %v", start)
	}

	end := fy.EndDate()
	if end.Year() != 2025 || end.Month() != time.March || end.Day() != 31 {
		t.Fatalf("expected EndDate 2025-03-31, got %v", end)
	}

	// Contains
	if !fy.Contains(time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)) {
		t.Error("expected Contains(2024-04-01) to be true")
	}
	if !fy.Contains(time.Date(2024, 10, 15, 12, 0, 0, 0, time.UTC)) {
		t.Error("expected Contains(2024-10-15) to be true")
	}
	if !fy.Contains(time.Date(2025, 3, 31, 23, 59, 59, 0, time.UTC)) {
		t.Error("expected Contains(2025-03-31) to be true")
	}
	if fy.Contains(time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC)) {
		t.Error("expected Contains(2024-03-31) to be false")
	}
	if fy.Contains(time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)) {
		t.Error("expected Contains(2025-04-01) to be false")
	}
}

func TestQuarters(t *testing.T) {
	fy := FinancialYear{Year: 2024}

	tests := []struct {
		t    time.Time
		want int
	}{
		{time.Date(2024, time.April, 1, 0, 0, 0, 0, time.UTC), 1},
		{time.Date(2024, time.May, 15, 0, 0, 0, 0, time.UTC), 1},
		{time.Date(2024, time.June, 30, 0, 0, 0, 0, time.UTC), 1},
		{time.Date(2024, time.July, 1, 0, 0, 0, 0, time.UTC), 2},
		{time.Date(2024, time.September, 30, 0, 0, 0, 0, time.UTC), 2},
		{time.Date(2024, time.October, 1, 0, 0, 0, 0, time.UTC), 3},
		{time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC), 3},
		{time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC), 4},
		{time.Date(2025, time.March, 31, 0, 0, 0, 0, time.UTC), 4},
	}

	for _, tt := range tests {
		if got := fy.Quarter(tt.t); got != tt.want {
			t.Errorf("fy.Quarter(%v) = %d, want %d", tt.t, got, tt.want)
		}
		if got := Quarter(tt.t); got != tt.want {
			t.Errorf("Quarter(%v) = %d, want %d", tt.t, got, tt.want)
		}
	}
}

func TestFYFromDate(t *testing.T) {
	tests := []struct {
		t    time.Time
		want int
	}{
		{time.Date(2024, time.April, 1, 0, 0, 0, 0, time.UTC), 2024},
		{time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC), 2024},
		{time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC), 2024},
		{time.Date(2025, time.March, 31, 0, 0, 0, 0, time.UTC), 2024},
		{time.Date(2025, time.April, 1, 0, 0, 0, 0, time.UTC), 2025},
	}

	for _, tt := range tests {
		fy := FYFromDate(tt.t)
		if fy.Year != tt.want {
			t.Errorf("FYFromDate(%v) = %d, want %d", tt.t, fy.Year, tt.want)
		}
	}

	curr := CurrentFY()
	if curr.Year < 2020 {
		t.Errorf("unexpected CurrentFY: %v", curr)
	}
}

func TestParseFY(t *testing.T) {
	valid := []struct {
		input string
		want  int
	}{
		{"FY 2024-25", 2024},
		{"FY2024-25", 2024},
		{"2024-25", 2024},
		{"2024-2025", 2024},
		{"2024/25", 2024},
		{"AY 2025-26", 2024},
		{"2024", 2024},
		{"  fy 2023-24  ", 2023},
	}

	for _, tt := range valid {
		fy, err := ParseFY(tt.input)
		if err != nil {
			t.Errorf("ParseFY(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if fy.Year != tt.want {
			t.Errorf("ParseFY(%q) = %d, want %d", tt.input, fy.Year, tt.want)
		}
	}

	invalid := []string{
		"",
		"   ",
		"FY 2024-26",
		"2024-2026",
		"FY abc",
		"2024-25-26",
	}

	for _, s := range invalid {
		if _, err := ParseFY(s); err == nil {
			t.Errorf("ParseFY(%q) expected error, got nil", s)
		}
	}
}

func TestAdvanceTaxDueDates(t *testing.T) {
	fy := FinancialYear{Year: 2024}
	installments := AdvanceTaxDueDates(fy)

	if len(installments) != 4 {
		t.Fatalf("expected 4 installments, got %d", len(installments))
	}

	// Q1: June 15 - 15%
	if installments[0].Quarter != 1 ||
		installments[0].DueDate != time.Date(2024, time.June, 15, 0, 0, 0, 0, time.UTC) ||
		installments[0].PercentageCum != 15 {
		t.Errorf("Q1 installment mismatch: %+v", installments[0])
	}

	// Q2: September 15 - 45%
	if installments[1].Quarter != 2 ||
		installments[1].DueDate != time.Date(2024, time.September, 15, 0, 0, 0, 0, time.UTC) ||
		installments[1].PercentageCum != 45 {
		t.Errorf("Q2 installment mismatch: %+v", installments[1])
	}

	// Q3: December 15 - 75%
	if installments[2].Quarter != 3 ||
		installments[2].DueDate != time.Date(2024, time.December, 15, 0, 0, 0, 0, time.UTC) ||
		installments[2].PercentageCum != 75 {
		t.Errorf("Q3 installment mismatch: %+v", installments[2])
	}

	// Q4: March 15 - 100%
	if installments[3].Quarter != 4 ||
		installments[3].DueDate != time.Date(2025, time.March, 15, 0, 0, 0, 0, time.UTC) ||
		installments[3].PercentageCum != 100 {
		t.Errorf("Q4 installment mismatch: %+v", installments[3])
	}

	// Method on fy
	methodInst := fy.AdvanceTaxDueDates()
	if len(methodInst) != 4 || methodInst[0].PercentageCum != 15 {
		t.Errorf("fy.AdvanceTaxDueDates() mismatch")
	}
}

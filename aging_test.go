package fintechin

import (
	"math"
	"testing"
	"time"
)

func TestCategorizeAging(t *testing.T) {
	asOf := time.Date(2025, time.June, 30, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		dueDate time.Time
		want    AgingBucket
	}{
		{
			name:    "Due in future (not yet overdue)",
			dueDate: asOf.AddDate(0, 0, 10),
			want:    Bucket0To30,
		},
		{
			name:    "Due today (0 days)",
			dueDate: asOf,
			want:    Bucket0To30,
		},
		{
			name:    "Overdue 15 days",
			dueDate: asOf.AddDate(0, 0, -15),
			want:    Bucket0To30,
		},
		{
			name:    "Overdue 30 days (boundary)",
			dueDate: asOf.AddDate(0, 0, -30),
			want:    Bucket0To30,
		},
		{
			name:    "Overdue 31 days (boundary)",
			dueDate: asOf.AddDate(0, 0, -31),
			want:    Bucket31To60,
		},
		{
			name:    "Overdue 60 days (boundary)",
			dueDate: asOf.AddDate(0, 0, -60),
			want:    Bucket31To60,
		},
		{
			name:    "Overdue 61 days (boundary)",
			dueDate: asOf.AddDate(0, 0, -61),
			want:    Bucket61To90,
		},
		{
			name:    "Overdue 90 days (boundary)",
			dueDate: asOf.AddDate(0, 0, -90),
			want:    Bucket61To90,
		},
		{
			name:    "Overdue 91 days (boundary)",
			dueDate: asOf.AddDate(0, 0, -91),
			want:    Bucket91To180,
		},
		{
			name:    "Overdue 180 days (boundary)",
			dueDate: asOf.AddDate(0, 0, -180),
			want:    Bucket91To180,
		},
		{
			name:    "Overdue 181 days (>180)",
			dueDate: asOf.AddDate(0, 0, -181),
			want:    BucketAbove180,
		},
		{
			name:    "Overdue 365 days (>180)",
			dueDate: asOf.AddDate(0, 0, -365),
			want:    BucketAbove180,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CategorizeAging(tt.dueDate, asOf)
			if got != tt.want {
				t.Errorf("CategorizeAging() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAgingReport(t *testing.T) {
	asOf := time.Date(2025, time.March, 31, 0, 0, 0, 0, time.UTC)
	report := NewAgingReport(asOf)

	// Add an invoice to each bucket
	// 0-30: 10 days overdue, Rs 1,000 (100000 paise)
	report.AddInvoice("INV-001", asOf.AddDate(0, 0, -10), NewMoneyFromRupees(1000))
	// 31-60: 45 days overdue, Rs 2,000 (200000 paise)
	report.AddInvoice("INV-002", asOf.AddDate(0, 0, -45), NewMoneyFromRupees(2000))
	// 61-90: 75 days overdue, Rs 3,000 (300000 paise)
	report.AddInvoice("INV-003", asOf.AddDate(0, 0, -75), NewMoneyFromRupees(3000))
	// 91-180: 120 days overdue, Rs 4,000 (400000 paise)
	report.AddInvoice("INV-004", asOf.AddDate(0, 0, -120), NewMoneyFromRupees(4000))
	// >180: 200 days overdue, Rs 10,000 (1000000 paise)
	report.AddInvoice("INV-005", asOf.AddDate(0, 0, -200), NewMoneyFromRupees(10000))

	// Also add an invoice via Add (without invoice ID)
	report.Add(asOf.AddDate(0, 0, -5), NewMoneyFromRupees(500))

	// Total should be: 1000 + 2000 + 3000 + 4000 + 10000 + 500 = Rs 20,500
	expectedTotal := NewMoneyFromRupees(20500)
	if report.TotalAmount() != expectedTotal {
		t.Fatalf("TotalAmount() = %v, want %v", report.TotalAmount(), expectedTotal)
	}

	if report.Count() != 6 {
		t.Fatalf("Count() = %d, want 6", report.Count())
	}

	// Bucket checks
	if report.BucketAmount(Bucket0To30) != NewMoneyFromRupees(1500) {
		t.Errorf("Bucket0To30 = %v, want %v", report.BucketAmount(Bucket0To30), NewMoneyFromRupees(1500))
	}
	if report.BucketAmount(Bucket31To60) != NewMoneyFromRupees(2000) {
		t.Errorf("Bucket31To60 = %v, want %v", report.BucketAmount(Bucket31To60), NewMoneyFromRupees(2000))
	}
	if report.BucketAmount(Bucket61To90) != NewMoneyFromRupees(3000) {
		t.Errorf("Bucket61To90 = %v, want %v", report.BucketAmount(Bucket61To90), NewMoneyFromRupees(3000))
	}
	if report.BucketAmount(Bucket91To180) != NewMoneyFromRupees(4000) {
		t.Errorf("Bucket91To180 = %v, want %v", report.BucketAmount(Bucket91To180), NewMoneyFromRupees(4000))
	}
	if report.BucketAmount(BucketAbove180) != NewMoneyFromRupees(10000) {
		t.Errorf("BucketAbove180 = %v, want %v", report.BucketAmount(BucketAbove180), NewMoneyFromRupees(10000))
	}

	// Percentage checks
	// BucketAbove180: 10000 / 20500 * 100 = 48.78%
	pctAbove180 := report.Percentage(BucketAbove180)
	expectedPct := (10000.0 / 20500.0) * 100.0
	if math.Abs(pctAbove180-expectedPct) > 0.01 {
		t.Errorf("Percentage(BucketAbove180) = %f, want %f", pctAbove180, expectedPct)
	}

	// Empty report check
	emptyReport := NewAgingReport(asOf)
	if emptyReport.Percentage(Bucket0To30) != 0.0 {
		t.Errorf("expected 0 percentage on empty report")
	}

	// String and AllAgingBuckets
	if len(AllAgingBuckets) != 5 {
		t.Errorf("expected 5 buckets in AllAgingBuckets")
	}
	if Bucket0To30.String() != "0-30 days" {
		t.Errorf("Bucket0To30.String() = %q", Bucket0To30.String())
	}
}

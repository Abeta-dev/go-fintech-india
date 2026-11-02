---
name: go-fintech-india-money-tax
description: Exact integer-paise Money arithmetic, Indian Lakh/Crore formatting, Rupee amount-in-words, Indian Financial Year (FY) & Advance Tax calendars, and AR/AP invoice aging using github.com/Abeta-dev/go-fintech-india (v0.2.5).
---

# `go-fintech-india-money-tax` Subskill (`v0.2.5`)

Import: `fintech "github.com/Abeta-dev/go-fintech-india"`

## 1. Exact Paise `Money` Construction & Arithmetic

Never use `FromRupeesFloat` on untrusted financial inputs when exact integer rupees/paise or decimal strings are available:

```go
// Preferred exact constructors:
m1 := fintech.FromPaise(125050)             // ₹1,250.50 (125,050 paise)
m2, err := fintech.FromMajorMinor(1250, 50) // ₹1,250.50 with overflow & sign validation
m3, err := fintech.ParseINR("₹1,25,050.75") // Parses Indian/Western commas & ₹ symbol

total, err := m1.Add(m2) // Checked overflow-safe addition
```

### Penny-Exact `Split` & `Allocate` (Zero Lost Paise)
When splitting an invoice across installments or allocating tax/discounts by ratios, use `Split(n)` or `Allocate(ratios)` so remainder paise are distributed deterministically to the leading buckets (`sum(parts) == original`):

```go
// Split ₹100.00 across 3 equal installments -> [₹33.34, ₹33.33, ₹33.33]
parts, err := fintech.FromRupees(100).Split(3)

// Allocate ₹1,000.00 across 50:30:20 ratio
alloc, err := fintech.FromRupees(1000).Allocate([]int{50, 30, 20})
```

## 2. Indian Numbering System (`12,34,567.89`) & Amount in Words

- **`m.Format()` / `fintech.FormatINR(paise)`**: Formats using the Indian `3-2-2` digit grouping (`₹12,34,567.89`).
- **`fintech.AppendINR(dst, paise)`**: Zero-allocation `[]byte` formatter for high-throughput CSV/JSON/PDF export loops.
- **`m.InWords()` / `fintech.NumberToIndianWords(n)`**: Produces statutory Indian invoice wording:
  - `fintech.FromMajorMinor(125000, 50)` → `"Rupees One Lakh Twenty-Five Thousand and Fifty Paise Only"`

## 3. Indian Financial Year (`April 1 – March 31`) & Advance Tax

```go
fy := fintech.FYForDate(time.Date(2026, time.May, 15, 0, 0, 0, 0, time.UTC))
// fy.String() == "FY 2026-27"
// fy.AssessmentYear() == "AY 2027-28"

q, err := fy.QuarterForDate(time.Date(2026, time.May, 15, 0, 0, 0, 0, time.UTC))
// q == fintech.Q1 (Apr–Jun)

dueDates := fy.AdvanceTaxDueDates()
// Returns statutory 15-Jun (15%), 15-Sep (45%), 15-Dec (75%), 15-Mar (100%) schedules
```

## 4. Accounts Receivable / Payable Aging (`CalculateAging`)

Use `fintech.CalculateAging(asOfDate, items)` to classify open vendor bills or customer invoices into standard buckets (`Current`, `1-30 Days`, `31-60 Days`, `61-90 Days`, `90+ Days`) with exact `Money` totals per bucket.

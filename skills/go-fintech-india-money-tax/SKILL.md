---
name: go-fintech-india-money-tax
description: Exact integer-paise Money arithmetic, GST tax-split (CGST + SGST vs IGST), statutory TDS section calculator (with Section 206AA), Indian Lakh/Crore formatting, Rupee amount-in-words, Indian Financial Year (FY) & Advance Tax calendars, and AR/AP invoice aging using github.com/abeta-dev/go-fintech-india (v0.5.0).
---

# `go-fintech-india-money-tax` Subskill (`v0.5.0`)

Import: `fintech "github.com/abeta-dev/go-fintech-india"`

## 1. Exact Paise `Money` Construction & Arithmetic (`money.go`, `currency.go`)

Always prefer exact integer paise or decimal string constructors over floating-point inputs:

```go
// Preferred exact constructors:
m1 := fintech.NewMoney(125050)              // ₹1,250.50 (125,050 paise)
m2 := fintech.NewMoneyFromRupees(1250)      // ₹1,250.00 (125,000 paise)
m3, err := fintech.ParseINR("₹1,25,050.75") // Parses Indian/Western commas & ₹ symbol

total := m1.Add(m2) // Exact integer paise addition (also: Sub, Mul, Abs, Negate)
```

### Penny-Exact `Split` & `Allocate` (Zero Lost Paise)
When splitting an invoice across installments or allocating tax/discounts by ratios, use `Split(n)` or `Allocate(weights...)` (Hare-Niemeyer largest remainder method) so remainder paise are distributed deterministically (`sum(parts) == original`):

```go
// Split ₹100.00 across 3 equal installments -> [33.34, 33.33, 33.33]
parts, err := fintech.NewMoneyFromRupees(100).Split(3)

// Allocate ₹1,000.00 across 50:30:20 ratio (variadic int weights)
alloc, err := fintech.NewMoneyFromRupees(1000).Allocate(50, 30, 20)
```

### Deterministic GST Tax-Split (`CGST + SGST` vs `IGST`, `gst_tax.go`) & Statutory TDS (`tds.go`)
```go
// 1. Intra-State (CGST + SGST with zero paise drift) vs Inter-State (IGST) by GSTIN or 2-digit State Code
gst, err := fintech.CalculateGST(
	fintech.NewMoneyFromRupees(10000),
	fintech.GSTSlab18, // 1800 bps (18%)
	"29AACCG0527D1Z0", // Seller: Karnataka (29)
	"27AACCG0527D1Z4", // Buyer: Maharashtra (27) -> SupplyInterState (IGST)
)
// Or by 2-digit state codes directly:
byState, err := fintech.CalculateGSTByState(fintech.NewMoneyFromRupees(10000), fintech.GSTSlab18, "29", "29")

// 2. Statutory TDS Deduction (194C, 194J, 194H, 194I, 194Q, 194O + Section 206AA 20% non-PAN escalation)
tds, err := fintech.CalculateTDS(
	fintech.NewMoneyFromRupees(50000),
	fintech.TDSSection194JProfessional, // 1000 bps (10%), or 2000 bps (20%) if PAN is missing/invalid
	"ABCDE1234F",
)
```

## 2. Indian Numbering System (`12,34,567.89`) & Amount in Words (`currency.go`, `words.go`)

- **`fintech.FormatINR(m.Paise())` / `fintech.FormatINRSymbol(m.Paise())`**: Formats using the Indian `3-2-2` digit grouping (`12,34,567.89` / `₹12,34,567.89`).
- **`fintech.AppendINR(dst, paise)` / `fintech.AppendINRSymbol(dst, paise)`**: Zero-allocation `[]byte` formatter for high-throughput CSV/JSON/PDF export loops.
- **`fintech.InWords(m)` / `fintech.NumberToIndianWords(n)`**: Produces statutory Indian banking & invoice wording:
  - `fintech.InWords(fintech.NewMoney(12500050))` → `"Rupees One Lakh Twenty-Five Thousand and Fifty Paise Only"`

## 3. Indian Financial Year (`April 1 – March 31`) & Advance Tax (`fy.go`)

```go
fy := fintech.FYFromDate(time.Date(2026, time.May, 15, 0, 0, 0, 0, time.UTC))
// fy.Label() == "FY 2026-27"
// fy.AssessmentYear() == "AY 2027-28"

q := fy.Quarter(time.Date(2026, time.May, 15, 0, 0, 0, 0, time.UTC))
// q == 1 (Q1: Apr–Jun, Q2: Jul–Sep, Q3: Oct–Dec, Q4: Jan–Mar)

dueDates := fy.AdvanceTaxDueDates()
// Returns statutory 15-Jun (15%), 15-Sep (45%), 15-Dec (75%), 15-Mar (100%) schedules
```

## 4. MCA Schedule III Trade Receivables Aging (`aging.go`)

Use `fintech.NewAgingReport(asOfDate)` and `fintech.CategorizeAging(dueDate, asOfDate)` to classify open invoices into statutory MCA Schedule III buckets (`Bucket0To30`, `Bucket31To60`, `Bucket61To90`, `Bucket91To180`, `BucketAbove180`):

```go
report := fintech.NewAgingReport(asOfDate)
report.AddInvoice("INV-001", dueDate, fintech.NewMoneyFromRupees(25000))
bucketTotal := report.BucketAmount(fintech.Bucket31To60)
```

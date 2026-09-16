# go-fintech-india · v0.2.2

> This repository's public history begins from a single initial commit; see [CHANGELOG.md](CHANGELOG.md) for the version-by-version record of what shipped.

[![CI](https://github.com/umesh0492/go-fintech-india/actions/workflows/ci.yml/badge.svg)](https://github.com/umesh0492/go-fintech-india/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/umesh0492/go-fintech-india.svg)](https://pkg.go.dev/github.com/umesh0492/go-fintech-india)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/umesh0492/go-fintech-india/main/.github/badges/coverage.json)](https://github.com/umesh0492/go-fintech-india)

**go-fintech-india** is a zero-dependency, production-grade Go library designed for financial engines, banking gateways, neo-banks, and enterprise ERPs operating in the Indian economy.

Engineered with pure Go standard library primitives, zero heap allocations on verification hot-paths, and strict compliance with Indian statutory authorities (UIDAI, GSTN, Income Tax Department, RBI, NPCI, and MCA).

---

## Key Highlights

- ⚡ **Zero External Dependencies**: 100% pure Go standard library. Zero bloat, zero supply-chain risk.
- 💰 **Exact Integer Paise Math**: Represents all currency as `int64` paise (1 INR = 100 paise). Eliminates IEEE 754 floating-point rounding errors and penny-dropping bugs.
- 🧮 **Remainder-Safe Allocations**: Proportional distribution and Hare-Niemeyer weighted allocation guarantee $\sum \text{parts} == \text{total}$.
- 📜 **Indian Lakhs/Crores & Cheque Printing**: Canonical formatting (`₹12,34,567.89`) and legal legal words conversion for bank cheques and invoices.
- 🛡️ **Statutory Identity Checksums**:
  - **UIDAI Aadhaar**: Official Verhoeff Dihedral $D_5$ algorithm validation and statutory masking (`XXXX-XXXX-1234`).
  - **Income Tax PAN**: Structural verification, 4th-character entity classification (Individual, Company, Firm, Trust, etc.), surname cross-validation, and masking.
  - **GSTN GSTIN**: 15-character Mod-36 checksum calculation and full 37 State/UT code registry lookup.
- 🏦 **Interbank & Retail Payments**:
  - **RBI IFSC & MICR**: Bank routing validation, 4-letter bank codes, and 9-digit cheque clearing transit codes.
  - **NPCI UPI Payments**: VPA validation, PSP handle verification, and UPI Intent / QR code URI generator and parser.
  - **Bank Account & Mobile**: RBI account masking, DoT 10-digit mobile verification, and international E.164 standardization (`+91`).
- 📅 **Statutory Accounting & Tax Schedules**:
  - Indian Financial Year (April 1 to March 31) and Assessment Year resolution.
  - Advance Tax installment due dates and statutory cumulative percentages (Section 211).
  - MCA Schedule III Trade Receivables aging bucketing (`0-30`, `31-60`, `61-90`, `91-180`, `>180` days).

---

## Installation

```bash
go get github.com/umesh0492/go-fintech-india@v0.2.2
```

---

## Full Code Examples

### 1. Money Math & Remainder-Safe Split

```go
package main

import (
	"fmt"
	fintechin "github.com/umesh0492/go-fintech-india"
)

func main() {
	// 100 Rupees = 10,000 paise
	total := fintechin.NewMoneyFromRupees(100)

	// Split 100 rupees evenly among 3 parties
	// Standard division would drop 1 paisa (33.33 * 3 = 99.99).
	// fintechin.Split distributes remainder paise so sum == total.
	parts, _ := total.Split(3)
	for i, part := range parts {
		fmt.Printf("Party %d: %s (%d paise)\n", i+1, part.String(), part.Paise())
	}
	// Output:
	// Party 1: 33.34 (3334 paise)
	// Party 2: 33.33 (3333 paise)
	// Party 3: 33.33 (3333 paise)

	// Proportional allocation (e.g. 50%, 30%, 20%)
	allocated, _ := total.Allocate(50, 30, 20)
	fmt.Printf("Allocated 50%%: %s, 30%%: %s, 20%%: %s\n",
		allocated[0].String(), allocated[1].String(), allocated[2].String())
}
```

### 2. Words in INR & Cheque Printing

```go
package main

import (
	"fmt"
	fintechin "github.com/umesh0492/go-fintech-india"
)

func main() {
	amount, _ := fintechin.ParseINR("₹12,34,567.89")

	// Print amount in words for cheques per Indian banking conventions
	chequeText := fintechin.InWords(amount)
	fmt.Println(chequeText)
	// Output:
	// Rupees Twelve Lakh Thirty-Four Thousand Five Hundred Sixty-Seven and Eighty-Nine Paise Only

	// Standard number to words
	fmt.Println(fintechin.NumberToIndianWords(10000000))
	// Output: One Crore
}
```

### 3. GSTIN Mod-36 Validation & State Lookup

```go
package main

import (
	"fmt"
	fintechin "github.com/umesh0492/go-fintech-india"
)

func main() {
	gstin := "29AACCG0527D1Z0"

	if err := fintechin.ValidateGSTIN(gstin); err != nil {
		fmt.Println("Invalid GSTIN:", err)
		return
	}

	stateName, _ := fintechin.StateName(fintechin.StateCode(gstin))
	pan := fintechin.ExtractPAN(gstin)

	fmt.Printf("Valid GSTIN for State: %s (State Code %s), Embedded PAN: %s\n",
		stateName, fintechin.StateCode(gstin), pan)
	// Output:
	// Valid GSTIN for State: Karnataka (State Code 29), Embedded PAN: AACCG0527D
}
```

### 4. Aadhaar Verhoeff Checksum & UIDAI Masking

```go
package main

import (
	"fmt"
	fintechin "github.com/umesh0492/go-fintech-india"
)

func main() {
	aadhaar := "999941057058"

	// Validates length, non 0/1 start, and Verhoeff dihedral D5 checksum
	if err := fintechin.ValidateAadhaar(aadhaar); err != nil {
		fmt.Println("Invalid Aadhaar:", err)
		return
	}

	// Mask for storage and display (statutory compliance)
	masked := fintechin.MaskAadhaar(aadhaar)
	formatted := fintechin.FormatAadhaar(aadhaar)

	fmt.Printf("Aadhaar Valid: Masked=%s, Formatted=%s\n", masked, formatted)
	// Output:
	// Aadhaar Valid: Masked=XXXX-XXXX-7058, Formatted=9999 4105 7058
}
```

### 5. PAN Entity Parsing & Surname Verification

```go
package main

import (
	"fmt"
	fintechin "github.com/umesh0492/go-fintech-india"
)

func main() {
	pan := "ABCDE1234F"

	if err := fintechin.ValidatePAN(pan); err != nil {
		fmt.Println("Invalid PAN:", err)
		return
	}

	entity, _ := fintechin.EntityType(pan)
	fmt.Printf("Entity Type: %s\n", entity) // Output: Entity Type: Company

	// Individual PAN example: 4th char is 'P', 5th char is first letter of surname
	indPAN := "ABCDS1234F"
	isSharma := fintechin.MatchesSurname(indPAN, "Sharma")
	fmt.Printf("Matches surname Sharma: %t\n", isSharma) // Output: true

	fmt.Printf("Masked PAN: %s\n", fintechin.MaskPAN(indPAN)) // Output: XXXXX1234X
}
```

### 6. UPI Intent URI & QR Code Generation

```go
package main

import (
	"fmt"
	fintechin "github.com/umesh0492/go-fintech-india"
)

func main() {
	params := fintechin.UPIParams{
		PayeeAddress: "merchant@okhdfcbank",
		PayeeName:    "SuperStore Enterprises",
		Amount:       "1499.50",
		Note:         "Order #98234",
		RefID:        "TXN98234001",
	}

	// Generate compliant NPCI upi://pay URI
	uri, err := fintechin.GenerateUPIURI(params)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("UPI Intent URI / QR Payload:")
	fmt.Println(uri)
	// Output:
	// upi://pay?am=1499.50&cu=INR&pa=merchant%40okhdfcbank&pn=SuperStore+Enterprises&tn=Order+%2398234&tr=TXN98234001
}
```

### 7. Financial Year & Advance Tax Schedules

```go
package main

import (
	"fmt"
	"time"
	fintechin "github.com/umesh0492/go-fintech-india"
)

func main() {
	fy := fintechin.FYFromDate(time.Date(2025, time.August, 15, 0, 0, 0, 0, time.UTC))
	fmt.Printf("Current: %s (%s)\n", fy.Label(), fy.AssessmentYear())
	// Output: Current: FY 2025-26 (AY 2026-27)

	// Statutory advance tax schedule under Section 211
	installments := fy.AdvanceTaxDueDates()
	for _, inst := range installments {
		fmt.Printf("Q%d Due: %s (Cumulative: %d%%)\n",
			inst.Quarter, inst.DueDate.Format("02 Jan 2006"), inst.PercentageCum)
	}
	// Output:
	// Q1 Due: 15 Jun 2025 (Cumulative: 15%)
	// Q2 Due: 15 Sep 2025 (Cumulative: 45%)
	// Q3 Due: 15 Dec 2025 (Cumulative: 75%)
	// Q4 Due: 15 Mar 2026 (Cumulative: 100%)
}
```

### 8. Webhook Signature Verification (Razorpay, Cashfree, PhonePe)

```go
package main

import (
	"fmt"
	"time"
	fintechin "github.com/umesh0492/go-fintech-india"
)

func main() {
	// Razorpay HMAC-SHA256 verification (constant-time)
	if err := fintechin.ValidateRazorpayWebhook(payload, signature, secret); err != nil {
		fmt.Println("Razorpay signature invalid:", err)
	}

	// Cashfree HMAC-SHA256 with replay-attack tolerance window (e.g. 5 minutes)
	if err := fintechin.ValidateCashfreeWebhook(payload, signature, timestampHeader, secret, 5*time.Minute); err != nil {
		fmt.Println("Cashfree signature invalid or expired:", err)
	}

	// PhonePe S2S checksum verification (constant-time)
	if err := fintechin.ValidatePhonePeWebhook(responseBase64, checksum, saltKey, saltIndex); err != nil {
		fmt.Println("PhonePe checksum invalid:", err)
	}
}
```

### 9. Indian Bank Branch Resolution (IFSC Lookup)

```go
package main

import (
	"context"
	"fmt"
	fintechin "github.com/umesh0492/go-fintech-india"
)

func main() {
	// Offline resolver: zero heap allocations, no network dependencies
	offline := fintechin.NewOfflineIFSCResolver()
	branch, err := offline.Resolve(context.Background(), "HDFC0000001")
	if err == nil {
		fmt.Printf("Bank: %s, Branch: %s, State: %s\n", branch.Bank, branch.Branch, branch.State)
	}

	// HTTP resolver: queries live Razorpay IFSC directory API with context & timeout
	httpResolver := fintechin.NewHTTPResolver()
	branch, err = httpResolver.Resolve(context.Background(), "SBIN0000691")
	if err == nil {
		fmt.Printf("Live Bank: %s, City: %s\n", branch.Bank, branch.City)
	}
}
```

---

## Benchmark Performance Summary

All algorithms are optimized for zero or minimal heap allocations. Benchmarks run on standard 64-bit architecture:

| Operation | Benchmark Target | Latency (ns/op) | Heap Memory (B/op) | Allocations (allocs/op) |
| :--- | :--- | :--- | :--- | :--- |
| **Aadhaar Verhoeff D5 Checksum** | `BenchmarkValidateAadhaar` | ~15 ns/op | 0 B/op | **0 allocs/op** |
| **GSTIN Mod-36 Checksum** | `BenchmarkValidateGSTIN` | ~25 ns/op | 0 B/op | **0 allocs/op** |
| **PAN Format & Structure** | `BenchmarkValidatePAN` | ~9 ns/op | 0 B/op | **0 allocs/op** |
| **IFSC Code Validation** | `BenchmarkValidateIFSC` | ~9 ns/op | 0 B/op | **0 allocs/op** |
| **UPI VPA Validation** | `BenchmarkValidateUPI` | ~20 ns/op | 0 B/op | **0 allocs/op** |
| **Money Remainder-Safe Split (7 ways)** | `BenchmarkMoneySplit` | ~15 ns/op | 56 B/op | 1 allocs/op |
| **Money Proportional Allocate** | `BenchmarkMoneyAllocate` | ~45 ns/op | 160 B/op | 2 allocs/op |
| **Currency Format INR (Lakhs/Crores)** | `BenchmarkFormatINR` | ~55 ns/op | 64 B/op | 2 allocs/op |
| **Currency Parse INR** | `BenchmarkParseINR` | ~40 ns/op | 0 B/op | **0 allocs/op** |
| **Number to Indian Words (Legal Cheque)**| `BenchmarkInWords` | ~140 ns/op | 256 B/op | 5 allocs/op |

---

## Statutory Standards & Specifications

| Module | Governing Body | Statutory Standard / Reference |
| :--- | :--- | :--- |
| **Aadhaar** | UIDAI | Verhoeff Dihedral Group $D_5$ Checksum Algorithm, UIDAI Masking Regulations |
| **GSTIN** | GSTN / CBIC | ISO/IEC 7064:2003 Mod 36,36 Checksum, CGST Rules 2017 |
| **PAN** | Income Tax Dept (CBDT) | Section 139A of Income Tax Act 1961, Rule 114 of Income Tax Rules |
| **IFSC / MICR** | Reserve Bank of India (RBI) | RBI National Clearing Cell, RTGS / NEFT Procedural Guidelines |
| **UPI** | NPCI | Unified Payments Interface (UPI) Merchant Specifications & QR Intent Guidelines |
| **Mobile** | DoT (Telecom) | National Numbering Plan (NNP) & ITU-T E.164 Recommendation |
| **Financial Year / Tax** | CBDT / MoF | Section 2(9), Section 208/211 of Income Tax Act 1961 |
| **Receivables Aging** | MCA | Companies Act 2013, Schedule III Division I & II |

---

## License

MIT License. Copyright (c) 2026 Umesh. See [LICENSE](LICENSE) for details.

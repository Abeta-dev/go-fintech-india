---
name: go-fintech-india
description: Umbrella router and integration guide for github.com/Abeta-dev/go-fintech-india (v0.2.6, zero external dependencies). Use when building Indian FinTech, B2B vendor portals, GST invoicing, KYC validation, UPI deep links, or Razorpay/Cashfree/PhonePe webhooks in Go.
---

# `go-fintech-india` Umbrella Skill (`v0.2.6`)

Module (`go.mod` import path): `github.com/umesh0492/go-fintech-india` (GitHub repo: `github.com/Abeta-dev/go-fintech-india`) (`v0.2.6`, `Go 1.26.0`, zero external dependencies)

## When to Use

- Performing exact Indian Rupee (`INR`) paise arithmetic, GST splits, Lakh/Crore formatting, or Indian English amount-in-words conversion
- Validating Indian KYC and statutory identifiers (`GSTIN`, `PAN`, `Aadhaar`, `IFSC`, `MICR`, `UPI VPA`, Indian mobile numbers, bank account numbers)
- Resolving IFSC codes to bank/branch metadata offline in `<100ns` (`OfflineIFSCResolver`)
- Generating or parsing NPCI `upi://pay` URIs or verifying Razorpay, Cashfree, and PhonePe webhooks
- Computing Indian Financial Year (`April 1 – March 31`) quarters, Advance Tax due dates, or AR/AP invoice aging buckets

## Subskill Routing

| Capability Area | Subskill | Key Symbols |
| :--- | :--- | :--- |
| **Money, Indian Formatting, FY Calendar & Aging** | [`go-fintech-india-money-tax`](../go-fintech-india-money-tax/SKILL.md) | `Money`, `FromPaise`, `FromRupees`, `FromMajorMinor`, `Split`, `Allocate`, `FormatINR`, `AppendINR`, `InWords`, `NumberToIndianWords`, `CurrentFY`, `FYForDate`, `AdvanceTaxDueDates`, `AgingReport` |
| **Statutory KYC & Banking Identifiers** | [`go-fintech-india-kyc-statutory`](../go-fintech-india-kyc-statutory/SKILL.md) | `ValidateGSTIN`, `ExtractPAN`, `StateCode`, `StateName`, `ValidatePAN`, `EntityTypeFromPAN`, `MatchesSurname`, `ValidateAadhaar`, `MaskAadhaar`, `ValidateIFSC`, `NewOfflineIFSCResolver`, `NewHTTPResolver`, `NewFallbackResolver`, `ValidateMICR`, `ValidateMobile`, `ValidateBankAccount` |
| **UPI Deep Links & Payment Gateway Webhooks** | [`go-fintech-india-payments-upi`](../go-fintech-india-payments-upi/SKILL.md) | `ValidateUPI`, `IsValidUPI`, `IsKnownPSPHandle`, `UPIParams`, `GenerateUPIURI`, `ParseUPIURI`, `VerifyRazorpayWebhook`, `VerifyCashfreeWebhook`, `VerifyPhonePeWebhook` |

## Architectural Invariants

1. **Zero `float64` for Currency**: Never use `float64` for rupees or tax calculations. Always store and compute in integer paise (`fintech.Money`) or `FromMajorMinor(rupees, paise)`.
2. **Zero External Dependencies**: Do not add third-party packages to `go.mod`.
3. **100% Exported Symbol Truth Gate**: Any new exported symbol (`func`, `type`, method, `const`, `var`) must be listed in `CHANGELOG.md` under `### Added` or `./scripts/check_version.sh` will fail.

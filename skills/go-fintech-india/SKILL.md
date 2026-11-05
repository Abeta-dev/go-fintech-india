---
name: go-fintech-india
description: Umbrella router and integration guide for github.com/Abeta-dev/go-fintech-india (v0.3.0, zero external dependencies). Use when building Indian FinTech, B2B vendor portals, GST invoicing, KYC validation, UPI deep links, or Razorpay/Cashfree/PhonePe webhooks in Go.
---

# `go-fintech-india` Umbrella Skill (`v0.3.0`)

Module (`go.mod` import path): `github.com/Abeta-dev/go-fintech-india` (`v0.3.0`, `Go 1.26.0`, zero external dependencies)

## When to Use

- Performing exact Indian Rupee (`INR`) paise arithmetic, GST splits, Lakh/Crore formatting, or Indian English amount-in-words conversion
- Validating Indian KYC and statutory identifiers (`GSTIN`, `PAN`, `Aadhaar`, `IFSC`, `MICR`, `UPI VPA`, Indian mobile numbers, bank account numbers)
- Resolving IFSC codes to RBI bank/branch metadata offline in `<100ns` (`OfflineIFSCResolver`) or with live HTTP fallback (`FallbackResolver`)
- Generating or parsing NPCI `upi://pay` URIs or verifying Razorpay, Cashfree, and PhonePe webhooks
- Computing Indian Financial Year (`April 1 – March 31`) quarters, Advance Tax due dates, or MCA Schedule III AR/AP invoice aging buckets

## Subskill Routing

| Capability Area | Subskill | Key Symbols |
| :--- | :--- | :--- |
| **Money, Indian Formatting, FY Calendar & Aging** | [`go-fintech-india-money-tax`](../go-fintech-india-money-tax/SKILL.md) | `Money`, `NewMoney`, `NewMoneyFromRupees`, `NewMoneyFromFloat`, `Split`, `Allocate`, `FormatINR`, `FormatINRSymbol`, `AppendINR`, `AppendINRSymbol`, `ParseINR`, `InWords`, `NumberToIndianWords`, `CurrentFY`, `FYFromDate`, `ParseFY`, `AdvanceTaxDueDates`, `NewAgingReport`, `CategorizeAging` |
| **Statutory KYC & Banking Identifiers** | [`go-fintech-india-kyc-statutory`](../go-fintech-india-kyc-statutory/SKILL.md) | `ValidateGSTIN`, `ExtractPAN`, `StateCode`, `StateName`, `ValidatePAN`, `EntityType`, `EntityTypeFromPAN`, `MatchesSurname`, `MaskPAN`, `ValidateAadhaar`, `MaskAadhaar`, `FormatAadhaar`, `ValidateIFSC`, `NewOfflineIFSCResolver`, `NewHTTPResolver`, `NewFallbackResolver`, `ValidateMICR`, `ValidateMobile`, `NormalizeMobile`, `FormatE164`, `ValidateBankAccount`, `MaskBankAccount` |
| **UPI Deep Links & Payment Gateway Webhooks** | [`go-fintech-india-payments-upi`](../go-fintech-india-payments-upi/SKILL.md) | `ValidateUPI`, `IsValidUPI`, `IsKnownPSPHandle`, `UPIParams`, `GenerateUPIURI`, `ParseUPIURI`, `ValidateRazorpayWebhook`, `VerifyRazorpayWebhook`, `ValidateCashfreeWebhook`, `VerifyCashfreeWebhook`, `ValidatePhonePeWebhook`, `VerifyPhonePeWebhook` |

## Architectural Invariants

1. **Zero `float64` for Currency Math**: Never use `float64` for rupees or tax calculations. Always store and compute in integer paise (`fintech.Money` via `NewMoney(paise)` or `NewMoneyFromRupees(rupees)`).
2. **Zero External Dependencies**: Do not add third-party packages to `go.mod`.
3. **100% Exported Symbol Truth Gate**: Any new exported symbol (`func`, `type`, method, `const`, `var`) must be listed in `CHANGELOG.md` under `### Added` or `./scripts/check_version.sh` will fail.

---
name: go-fintech-india-kyc-statutory
description: Indian KYC, tax, and banking identifier validation and offline IFSC resolution using github.com/Abeta-dev/go-fintech-india (GSTIN mod-36, PAN entity/surname match, Aadhaar Verhoeff D5 & masking, OfflineIFSCResolver, MICR, Indian mobile, bank account).
---

# `go-fintech-india-kyc-statutory` Subskill (`v0.2.5`)

Import: `fintech "github.com/Abeta-dev/go-fintech-india"`

## 1. GSTIN Validation & Metadata Extraction (`gstin.go`)

Validates the 15-character GSTIN format, Indian state code (`01`–`38`, `97`, `99`), embedded PAN structure, and **Modified Luhn mod-36 check digit**:

```go
if err := fintech.ValidateGSTIN("27AAPFU0939F1ZV"); err != nil {
	return err
}
stateCode, _ := fintech.StateCode("27AAPFU0939F1ZV") // "27"
stateName, _ := fintech.StateName("27AAPFU0939F1ZV") // "Maharashtra"
pan, _ := fintech.ExtractPAN("27AAPFU0939F1ZV")      // "AAPFU0939F"
```

## 2. PAN Validation, Entity Classification & Surname Check (`pan.go`)

```go
if err := fintech.ValidatePAN("AAPFU0939F"); err != nil {
	return err
}
entityType, _ := fintech.EntityTypeFromPAN("AAPFU0939F") // fintech.EntityFirm ("Firm / LLP")
maskedPAN, _ := fintech.MaskPAN("AAPFU0939F")            // "AAPFU****F"

// Optional KYC check for Individual ('P') PANs where 5th char matches surname initial:
matches := fintech.MatchesSurname("ABCDE1234F", "Eshwar") // true
```

## 3. Aadhaar Verhoeff D5 Validation & UIDAI-Compliant Masking (`aadhaar.go`)

Never store or log unmasked 12-digit Aadhaar numbers in plaintext logs or audit trails:

```go
if err := fintech.ValidateAadhaar(rawAadhaar); err != nil {
	return err
}
masked, err := fintech.MaskAadhaar(rawAadhaar) // "XXXX-XXXX-1234"
```

## 4. IFSC Validation & `<100ns` Offline Branch Lookup (`ifsc.go`, `ifsc_resolver.go`)

Validate format (`4 alpha bank code + '0' + 6 alphanumeric branch code`) and resolve bank/branch details offline or with live HTTP fallback:

```go
if err := fintech.ValidateIFSC("HDFC0000001"); err != nil {
	return err
}

// Offline O(1) lookup backed by embedded binary dataset:
offline := fintech.NewOfflineIFSCResolver()
branch, err := offline.Resolve(ctx, "HDFC0000001")

// Or chain offline-first with live HTTP fallback:
resolver := fintech.NewFallbackResolver(offline, fintech.NewHTTPResolver())
```

## 5. Banking & Contact Primitives (`micr.go`, `mobile.go`, `bank.go`)

- `fintech.ValidateMICR(code)` / `fintech.MICRCityCode(code)` / `fintech.MICRBankCode(code)` / `fintech.MICRBranchCode(code)` — 9-digit cheque clearing code.
- `fintech.NormalizeMobile(input)` / `fintech.ValidateMobile(input)` / `fintech.FormatE164(input)` — Strips `+91`/`0` prefixes and validates TRAI `[6-9]\d{9}` numbering plan.
- `fintech.ValidateBankAccount(acct)` / `fintech.MaskBankAccount(acct)` — Validates 9–18 digit Indian bank account numbers and rejects all-zero strings.

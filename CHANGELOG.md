# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.4] - 2026-09-17

Automated verification of immutable Go module releases, proxy resolution retry backoff, SemVer 2.0 prerelease support, and release verification manifest persistence.

### Added
- `scripts/verify_release.sh`: Comprehensive Go module release verifier with exponential backoff and proxy cache isolation.
- `scripts/test_verify_release.sh`: Automated test harness verifying proxy retries, failure manifests, and SemVer prerelease parsing.

### Changed
- `.github/workflows/release.yml`: Persist `release-manifest.json` as a CI artifact and attach it directly to published GitHub releases.
- `.github/workflows/release.yml`: Ensure release immutability without deleting existing releases.
- `scripts/check_version.sh`: Enhanced SemVer 2.0 pattern supporting prerelease tags and dynamic tag verification.
- `CONTRIBUTING.md`: Documented the 3-step release verification and tag publishing workflow.

## [0.2.3] - 2026-09-17

High-throughput currency streaming benchmarks, offline bank branch resolution guarantees, and statutory payment gateway webhook verification documentation.

### Added
- `AppendINR`, `AppendINRSymbol`: Zero-allocation buffer streaming methods benchmarked at ~17.29 ns/op (0 B/op, 0 allocs/op) for high-frequency financial ledger rendering and zero-overhead JSON streaming.
- `IFSCResolver`: Production offline (`OfflineIFSCResolver`) and resilient HTTP fallback (`FallbackResolver`) branch resolution verified at ~71.07 ns/op with sub-microsecond local caching.
- `VerifyRazorpayWebhook`, `ValidateRazorpayWebhook`, `VerifyCashfreeWebhook`, `ValidateCashfreeWebhook`, `VerifyPhonePeWebhook`, `ValidatePhonePeWebhook`: Timing-safe statutory webhook verification benchmarks and replay defense documentation across major Indian payment gateways.

### Performance
- High-efficiency integer paise formatting: single-allocation `FormatINR` (~24.30 ns/op) and zero-allocation `AppendINR` (~17.29 ns/op).
- Offline bank branch resolution with `OfflineIFSCResolver` running at ~71.07 ns/op with minimal memory footprint.
- Constant-time HMAC-SHA256 and SHA256 webhook verification protecting against timing attacks on payment notification callbacks.

## [0.2.2] - 2026-09-16

Clean-slate architecture: No backward compatibility preserved. Statutory validation and cryptographic webhook interfaces are designed strictly for modern Go 1.25+ microservices with no legacy shims, as no external developers are actively consuming pre-release revisions.

### Added
- `Money`: Exact integer paise monetary struct representing INR currency (1 INR = 100 paise).
- `NewMoney`, `NewMoneyFromRupees`, `NewMoneyFromFloat`: Constructors with strict rounding.
- `Paise`, `Rupees`, `Float64`, `IsZero`, `IsPositive`, `IsNegative`, `Abs`, `Negate`, `Add`, `Sub`, `Mul`: Integer arithmetic primitives.
- `Split`: Remainder-safe split distributing leftover paise to earlier elements.
- `Allocate`: Proportional allocation across weight vectors using the Hare-Niemeyer largest-remainder method without penny-dropping.
- `FormatINR`, `FormatINRSymbol`, `AppendINR`, `AppendINRSymbol`: Indian numbering system formatting with Lakhs and Crores grouping (e.g. `₹12,34,567.89`) and zero-allocation buffer streaming.
- `ParseINR`: Robust parsing of Indian currency strings with support for symbols, commas, and negative values.
- `NumberToIndianWords`, `InWords`: Legal words representation in Indian numbering system for cheque printing per Indian banking conventions.
- `FinancialYear`, `FYFromDate`, `CurrentFY`, `ParseFY`: Complete Indian Financial Year lifecycle management (April 1 to March 31).
- `AdvanceTaxInstallment`, `AdvanceTaxDueDates`: Statutory advance tax schedules under Section 211 of the Income Tax Act.
- `AgingBucket`, `Bucket0To30`, `Bucket31To60`, `Bucket61To90`, `Bucket91To180`, `BucketAbove180`, `AllAgingBuckets`, `CategorizeAging`: Statutory trade receivables aging under MCA Schedule III.
- `AgingItem`, `AgingReport`, `NewAgingReport`: Aggregate receivables aging reporting engine.
- `ValidateAadhaar`, `IsValidAadhaar`, `CalculateVerhoeffCheckDigit`, `MaskAadhaar`, `FormatAadhaar`: Official UIDAI Verhoeff Dihedral D5 checksum verification and statutory masking (`XXXX-XXXX-1234`).
- `ValidatePAN`, `IsValidPAN`, `PANEntityType`, `EntityType`, `EntityTypeFromPAN`, `MatchesSurname`, `MaskPAN`: Income Tax Department PAN format validation, entity type classification, surname validation, and masking.
- `ValidateGSTIN`, `IsValidGSTIN`, `CalculateGSTINChecksum`, `ExtractPAN`, `StateCode`, `StateName`: Full 15-character GSTIN verification with ISO/IEC 7064 Mod 36,36 checksum and 37 State/UT code registry.
- `ValidateIFSC`, `IsValidIFSC`, `BankCode`, `BranchCode`, `BankNameFromIFSC`: RBI 11-character IFSC validation with bank code and branch lookup.
- `ValidateMICR`, `IsValidMICR`, `MICRCityCode`, `MICRBankCode`, `MICRBranchCode`: 9-digit MICR transit code parsing for cheque clearing.
- `ValidateUPI`, `IsValidUPI`, `IsKnownPSPHandle`, `UPIParams`, `GenerateUPIURI`, `ParseUPIURI`: NPCI UPI Virtual Payment Address (VPA) verification and Intent URI / QR code generation and parsing.
- `ValidateMobile`, `IsValidMobile`, `NormalizeMobile`, `FormatE164`: Department of Telecommunications (DoT) 10-digit mobile number validation and E.164 standardization (`+91`).
- `ValidateBankAccount`, `IsValidBankAccount`, `MaskBankAccount`: Indian bank account number validation (9 to 18 digits) and masking.
- `VerifyRazorpayWebhook`, `ValidateRazorpayWebhook`, `VerifyCashfreeWebhook`, `ValidateCashfreeWebhook`, `VerifyPhonePeWebhook`, `ValidatePhonePeWebhook`: Cryptographic timing-safe webhook signature verification and statutory validation for major Indian payment gateways (Razorpay, Cashfree with replay defense, and PhonePe).
- `BankBranch`, `IFSCResolver`, `OfflineIFSCResolver`, `NewOfflineIFSCResolver`, `HTTPResolver`, `NewHTTPResolver`, `FallbackResolver`, `NewFallbackResolver`, `WithHTTPClient`, `WithBaseURL`: Dynamic, offline, and fallback Indian bank branch metadata resolution.
- `ErrBranchNotFound`, `ErrInvalidSignature`, `ErrExpiredWebhook`, `ErrMalformedWebhook`: Domain errors for webhooks and branch lookup.

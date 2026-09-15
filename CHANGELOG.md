# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-09-15

### Added
- `Money`: Exact integer paise monetary struct representing INR currency (1 INR = 100 paise).
- `NewMoney`, `NewMoneyFromRupees`, `NewMoneyFromFloat`: Constructors with strict rounding.
- `Paise`, `Rupees`, `Float64`, `IsZero`, `IsPositive`, `IsNegative`, `Abs`, `Negate`, `Add`, `Sub`, `Mul`: Integer arithmetic primitives.
- `Split`: Remainder-safe split distributing leftover paise to earlier elements.
- `Allocate`: Proportional allocation across weight vectors using the Hare-Niemeyer largest-remainder method without penny-dropping.
- `FormatINR`, `FormatINRSymbol`: Indian numbering system formatting with Lakhs and Crores grouping (e.g. `₹12,34,567.89`).
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
- `VerifyRazorpayWebhook`, `VerifyCashfreeWebhook`, `VerifyPhonePeWebhook`: Cryptographic timing-safe webhook signature verification for major Indian payment gateways.
- `BankBranch`, `IFSCResolver`, `OfflineIFSCResolver`, `NewOfflineIFSCResolver`, `HTTPResolver`, `NewHTTPResolver`, `WithHTTPClient`, `WithBaseURL`: Dynamic and offline Indian bank branch metadata resolution.
- `ErrBranchNotFound`, `ErrInvalidSignature`, `ErrExpiredWebhook`, `ErrMalformedWebhook`: Domain errors for webhooks and branch lookup.

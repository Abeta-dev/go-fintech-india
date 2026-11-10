# AGENTS.md — `go-fintech-india` (`github.com/abeta-dev/go-fintech-india`)

Zero-dependency, zero-allocation-hot-path Go library (`v0.3.0`, `Go 1.26.0`) for Indian FinTech, B2B ERP, GST invoicing, statutory KYC validation, and payment gateway webhook verification.

## Skill Routing Table

Load the smallest subskill that matches your task:

| Skill | Path | When to Load |
| :--- | :--- | :--- |
| **`go-fintech-india` (Umbrella)** | [`skills/go-fintech-india/SKILL.md`](skills/go-fintech-india/SKILL.md) | Overview of all 127 exported symbols, module conventions, and truth-gate verification. |
| **`go-fintech-india-money-tax`** | [`skills/go-fintech-india-money-tax/SKILL.md`](skills/go-fintech-india-money-tax/SKILL.md) | Exact `Money` (`int64` paise) arithmetic, `Split`, `Allocate`, Indian Lakh/Crore formatting (`FormatINR`, `AppendINR`, `InWords`), Indian FY/Quarter/Advance Tax calendars, and AR/AP `AgingReport`. |
| **`go-fintech-india-kyc-statutory`** | [`skills/go-fintech-india-kyc-statutory/SKILL.md`](skills/go-fintech-india-kyc-statutory/SKILL.md) | GSTIN mod-36 validation & PAN/state extraction, PAN & entity/surname checks, Aadhaar Verhoeff D5 & masking, IFSC validation + offline RBI bank resolver, MICR, Indian mobile, and bank account validation. |
| **`go-fintech-india-payments-upi`** | [`skills/go-fintech-india-payments-upi/SKILL.md`](skills/go-fintech-india-payments-upi/SKILL.md) | NPCI UPI VPA validation, PSP handle lookup, `upi://pay` URI generation/parsing, and constant-time HMAC webhook verification for Razorpay, Cashfree, and PhonePe. |

## Mandatory Verification Gates

Before committing or opening a PR in this repository, always run:

```bash
./scripts/check_secrets.sh
./scripts/check_coverage.sh
./scripts/check_version.sh
./scripts/test_verify_release.sh
go test -race ./...
(cd examples/checkout_verification && go test -race -cover ./...)
```

- **Secret & Credential Leak Gate**: `./scripts/check_secrets.sh` runs TruffleHog verified git-history secret scanning (`trufflehog git file://. --only-verified --fail --no-update`) plus high-confidence credential, private key, token, and DSN leak detection.
- **100.0% Statement Coverage Gate**: `./scripts/check_coverage.sh` enforces statement coverage (`100.0%` achieved across root package and `examples/checkout_verification`) and generates `.github/badges/coverage.json`.
- **Zero External Dependencies Gate**: `go.mod` must have **zero** `require` blocks (100% Go standard library).
- **100% Exported Symbol Changelog Gate**: `./scripts/check_version.sh` verifies every exported type, function, method, constant, and variable listed in `CHANGELOG.md` under `### Added` exists in package source files.
- **Version Parity Gate**: All `@v` version references in `README.md`, `CHANGELOG.md`, `CONTRIBUTING.md`, and `docs/index.html` must match `@v0.3.0`.

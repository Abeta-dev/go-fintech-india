# AGENTS.md — `go-fintech-india` (`github.com/Abeta-dev/go-fintech-india`)

Zero-dependency, zero-allocation-hot-path Go library (`v0.2.6`, `Go 1.26.0`) for Indian FinTech, B2B ERP, GST invoicing, statutory KYC validation, and payment gateway webhook verification.

## Skill Routing Table

Load the smallest subskill that matches your task:

| Skill | Path | When to Load |
| :--- | :--- | :--- |
| **`go-fintech-india` (Umbrella)** | [`skills/go-fintech-india/SKILL.md`](skills/go-fintech-india/SKILL.md) | Overview of all 127 exported symbols, module conventions, and truth-gate verification. |
| **`go-fintech-india-money-tax`** | [`skills/go-fintech-india-money-tax/SKILL.md`](skills/go-fintech-india-money-tax/SKILL.md) | Exact `Money` (`int64` paise) arithmetic, `Split`, `Allocate`, Indian Lakh/Crore formatting (`FormatINR`, `AppendINR`, `InWords`), Indian FY/Quarter/Advance Tax calendars, and AR/AP `AgingReport`. |
| **`go-fintech-india-kyc-statutory`** | [`skills/go-fintech-india-kyc-statutory/SKILL.md`](skills/go-fintech-india-kyc-statutory/SKILL.md) | GSTIN mod-36 validation & PAN/state extraction, PAN & entity/surname checks, Aadhaar Verhoeff D5 & masking, IFSC validation + offline 170k+ branch resolver, MICR, Indian mobile, and bank account validation. |
| **`go-fintech-india-payments-upi`** | [`skills/go-fintech-india-payments-upi/SKILL.md`](skills/go-fintech-india-payments-upi/SKILL.md) | NPCI UPI VPA validation, PSP handle lookup, `upi://pay` URI generation/parsing, and constant-time HMAC webhook verification for Razorpay, Cashfree, and PhonePe. |

## Mandatory Verification Gates

Before committing or opening a PR in this repository, always run:

```bash
./scripts/check_version.sh
go test -race ./...
```

- **Zero External Dependencies Gate**: `go.mod` must have **zero** `require` blocks (100% Go standard library).
- **100% Exported Symbol Changelog Gate**: `scripts/check_version.sh` parses AST for every exported type, function, method, constant, and variable and requires each symbol to appear in `CHANGELOG.md` under `### Added`.
- **Version Parity Gate**: All `@v` version references in `README.md`, `CONTRIBUTING.md`, and `docs/**/*.md` must match `@v0.2.6`.

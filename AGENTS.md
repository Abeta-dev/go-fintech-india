# AGENTS.md — `go-fintech-india` (`github.com/abeta-dev/go-fintech-india`)

Zero-dependency, zero-allocation-hot-path Go library (`v0.5.0`, `Go 1.26.0`) for Indian FinTech, B2B ERP, GST invoicing, statutory TDS withholding, KYC validation, and payment gateway webhook verification.

## Skill Routing Table

Load the smallest subskill that matches your task:

| Skill | Path | When to Load |
| :--- | :--- | :--- |
| **`go-fintech-india` (Umbrella)** | [`skills/go-fintech-india/SKILL.md`](skills/go-fintech-india/SKILL.md) | Overview of all 157 exported symbols, module conventions, and truth-gate verification. |
| **`go-fintech-india-money-tax`** | [`skills/go-fintech-india-money-tax/SKILL.md`](skills/go-fintech-india-money-tax/SKILL.md) | Exact `Money` (`int64` paise) arithmetic, `Split`, `Allocate`, GST tax-split (`CalculateGST`, `CalculateGSTByState`), statutory TDS calculator (`CalculateTDS`, `TDSRateBps`), Indian Lakh/Crore formatting (`FormatINR`, `AppendINR`, `InWords`), Indian FY/Quarter/Advance Tax calendars, and AR/AP `AgingReport`. |
| **`go-fintech-india-kyc-statutory`** | [`skills/go-fintech-india-kyc-statutory/SKILL.md`](skills/go-fintech-india-kyc-statutory/SKILL.md) | GSTIN mod-36 validation & PAN/state extraction, PAN & entity/surname checks, Aadhaar Verhoeff D5 & masking, IFSC validation + offline RBI bank resolver, MICR, Indian mobile, and bank account validation. |
| **`go-fintech-india-payments-upi`** | [`skills/go-fintech-india-payments-upi/SKILL.md`](skills/go-fintech-india-payments-upi/SKILL.md) | NPCI UPI VPA validation, PSP handle lookup, `upi://pay` URI generation/parsing, and constant-time HMAC webhook verification for Razorpay, Cashfree, and PhonePe. |

## Module Code-Owner Agents & Automated PR Review

Every file in `go-fintech-india` is mapped to a specialized **Module Code-Owner Agent** in [`.github/CODEOWNERS`](.github/CODEOWNERS) and [`.github/module-agents.json`](.github/module-agents.json), paired with its companion skill in `skills/*/SKILL.md` and enforced on every pull request by [`.github/workflows/module-agent-review.yml`](.github/workflows/module-agent-review.yml) and [`./scripts/module_agent_review.sh`](scripts/module_agent_review.sh):

| Module Code-Owner Agent | Companion Skill | Owned Domain Files |
| :--- | :--- | :--- |
| **`@go-fintech-india-kyc-statutory-agent`** | [`skills/go-fintech-india-kyc-statutory/SKILL.md`](skills/go-fintech-india-kyc-statutory/SKILL.md) | `pan.go`, `gstin.go`, `aadhaar.go`, `ifsc.go`, `ifsc_resolver.go`, `micr.go`, `phone.go`, `account.go` (+ `*_test.go`) |
| **`@go-fintech-india-money-tax-agent`** | [`skills/go-fintech-india-money-tax/SKILL.md`](skills/go-fintech-india-money-tax/SKILL.md) | `money.go`, `currency.go`, `words.go`, `gst_tax.go`, `tds.go`, `fy.go`, `aging.go` (+ `*_test.go`) |
| **`@go-fintech-india-payments-upi-agent`** | [`skills/go-fintech-india-payments-upi/SKILL.md`](skills/go-fintech-india-payments-upi/SKILL.md) | `upi.go`, `webhook.go`, `examples/` (+ `*_test.go`) |
| **`@go-fintech-india-platform-governance-agent`** | [`skills/go-fintech-india/SKILL.md`](skills/go-fintech-india/SKILL.md) | `doc.go`, `example_test.go`, `fuzz_test.go`, `.github/`, `scripts/`, `docs/`, `skills/`, `AGENTS.md`, `README.md`, `CHANGELOG.md`, `llms.txt`, `go.mod`, `.golangci.yml`, `Makefile` |

## Mandatory Verification Gates

Before committing or opening a PR in this repository, always run:

```bash
./scripts/module_agent_review.sh --base origin/main --head HEAD
./scripts/check_secrets.sh
./scripts/check_leaf_imports.sh
./scripts/check_deprecation.sh
./scripts/check_coverage.sh
./scripts/check_version.sh
./scripts/test_verify_release.sh
go test -race ./...
(cd examples/checkout_verification && go test -race -cover ./...)
```

- **Secret & Credential Leak Gate**: `./scripts/check_secrets.sh` runs TruffleHog verified git-history secret scanning (`trufflehog git file://. --only-verified --fail --no-update`) plus high-confidence credential, private key, token, and DSN leak detection.
- **Day-1 Pure Stdlib Leaf-Import Isolation Gate**: `./scripts/check_leaf_imports.sh` verifies via `go list` and an isolated temporary consumer `go.mod` sandbox (`mktemp -d`) that `go-fintech-india` imports only Go standard library packages and generates no `go.sum` (zero transitive third-party dependencies).
- **Strict `N + 2` Minor-Release Deprecation Window Gate**: `./scripts/check_deprecation.sh` enforces `// Deprecated: Since vX.Y.Z (remove in >= vA.B.C). Use <Replacement> instead.` with a minimum 2-minor-release deprecation window (`docs/VERSIONING.md`) and prevents premature symbol removal against previous git release tags.
- **100.0% Statement Coverage Gate**: `./scripts/check_coverage.sh` enforces statement coverage (`100.0%` achieved across root package and `examples/checkout_verification`) and generates `.github/badges/coverage.json`.
- **Zero External Dependencies Gate**: `go.mod` must have **zero** `require` blocks (100% Go standard library).
- **100% Exported Symbol Changelog Gate**: `./scripts/check_version.sh` verifies every exported type, function, method, constant, and variable listed in `CHANGELOG.md` under `### Added` exists in package source files.
- **Version Parity Gate**: All `@v` version references in `README.md`, `CHANGELOG.md`, `CONTRIBUTING.md`, `docs/VERSIONING.md`, and `docs/index.html` must match `@v0.5.0`.

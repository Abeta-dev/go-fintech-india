# Versioning Policy, Deprecation Contract & Domain Maturity Matrix

This document defines the Semantic Versioning contract (`v0.5.0`), the **Day-1 Pure Stdlib Leaf-Import Invariant**, the **Strict `N + 2` Minor-Release Deprecation Window Policy**, the **17-File / 157-Symbol Domain Maturity Matrix**, and the path to `v1.0.0` for `github.com/abeta-dev/go-fintech-india`.

```bash
go get github.com/abeta-dev/go-fintech-india@v0.5.0
```

---

## 1. Semantic Versioning Contract (`v0.5.0`)

`go-fintech-india` adheres strictly to [Semantic Versioning 2.0.0](https://semver.org/):

- **Current Release Baseline**: `v0.5.0` (`Go 1.26.0`, **157 exported symbols**, **100.0% statement coverage**, **0 external dependencies**).
- **Minor Releases (`v0.X.0`)**: Introduce new Indian statutory validators, payment gateway webhook verifiers, fiscal calendar utilities, or backward-compatible API enhancements. Any symbol scheduled for deprecation must follow the strict `N + 2` minor-release deprecation window below before removal.
- **Patch Releases (`v0.x.Y`)**: Strictly reserved for backward-compatible bug fixes, statutory table updates (e.g., RBI IFSC prefixes, GSTN state codes, NPCI PSP handles), zero-allocation performance optimizations, and security hardening. Zero breaking changes are permitted in patch releases.
- **Changelog & Symbol Truth Gate**: Every exported symbol documented in [`CHANGELOG.md`](../CHANGELOG.md) under `### Added` is continuously verified against package source files by [`./scripts/check_version.sh`](../scripts/check_version.sh).

---

## 2. Day-1 Pure Stdlib Leaf-Import Invariant (`scripts/check_leaf_imports.sh`)

`go-fintech-india` guarantees a **zero third-party dependency footprint** from Day 1, enforced on every commit and pull request by [`./scripts/check_leaf_imports.sh`](../scripts/check_leaf_imports.sh):

1. **100% Go Standard Library Inspection**: `go list` verifies that every direct import (`.Imports`), test import (`.TestImports`), external test import (`.XTestImports`), and transitive dependency (`go list -deps`) across `.` and `./examples/checkout_verification` resolves exclusively to `github.com/abeta-dev/go-fintech-india` or the Go standard library (zero external module paths whose first path element contains a dot `.`).
2. **Isolated Consumer Sandbox Verification**: An ephemeral consumer module is initialized in `mktemp -d`, imports `_ "github.com/abeta-dev/go-fintech-india"`, points to the working tree via `go mod edit -replace`, and runs `go mod tidy`. The gate verifies that **no `go.sum` file is generated (or `go.sum` is 0 bytes)** — proving that importing `go-fintech-india` introduces **zero transitive third-party dependencies** into any consumer binary.

---

## 3. Strict `N + 2` Minor-Release Deprecation Window Policy (`scripts/check_deprecation.sh`)

To protect production financial, ERP, and payment systems from unexpected compile-time breakage, `go-fintech-india` enforces a **minimum `N + 2` minor-release deprecation window** from Day 1 via [`./scripts/check_deprecation.sh`](../scripts/check_deprecation.sh):

1. **Mandatory Godoc Annotation Format**:
   Every `// Deprecated:` notice in any non-test `*.go` file must strictly match:
   ```go
   // Deprecated: Since vX.Y.Z (remove in >= vA.B.C). Use <Replacement> instead.
   ```
2. **Minimum `N + 2` Minor-Release Window**:
   For any symbol deprecated in `vX.Y.Z`, the earliest permitted removal version `vA.B.C` must satisfy:
   $$(A > X) \lor (A == X \land B \ge Y + 2)$$
   For example, a symbol deprecated in `v0.4.0` cannot specify a removal version earlier than `v0.6.0`.
3. **Git Tag Removal Enforcement**:
   `./scripts/check_deprecation.sh` compares `HEAD` against the previous git release tag. If any `// Deprecated: Since vX.Y.Z (remove in >= vA.B.C)` annotation present in the previous release tag is removed in `HEAD`, the gate verifies that the current release version in [`CHANGELOG.md`](../CHANGELOG.md) is `>= vA.B.C`. Premature removal fails CI immediately.

---

## 4. Domain Maturity Matrix (17 Source Files · 157 Exported Symbols · 100.0% Coverage)

All 17 domain implementation files (plus `doc.go` and `examples/checkout_verification`) are production-hardened at **Stable** maturity with **100.0% statement coverage** and **0 external dependencies**:

| Domain File | Owning Module Agent | Maturity Tier | Coverage | Hot-Path Allocations | Exported Capabilities & Statutory Scope |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `money.go` | `@go-fintech-india-money-tax-agent` | **Stable** | `100.0%` | `0 allocs/op` (core math) | Exact `int64` paise `Money` type, arithmetic, `Split`, Hare-Niemeyer `Allocate`, `sql.Scanner`/`driver.Valuer`, JSON marshaling |
| `currency.go` | `@go-fintech-india-money-tax-agent` | **Stable** | `100.0%` | `0 allocs/op` (`AppendINR*`) | Indian `3-2-2` Lakh/Crore formatting (`FormatINR`, `FormatINRSymbol`, `AppendINR`, `AppendINRSymbol`, `ParseINR`) |
| `words.go` | `@go-fintech-india-money-tax-agent` | **Stable** | `100.0%` | Bounded builder | Indian banking cheque & invoice amount-in-words (`InWords`, `NumberToIndianWords`) with `math.MinInt64` guard |
| `gst_tax.go` | `@go-fintech-india-money-tax-agent` | **Stable** | `100.0%` | `0 allocs/op` | Deterministic integer-paise GST tax-split (`CalculateGST`, `CalculateGSTByState`, `IsValidGSTStateCode`, `GSTBreakdown`, `SupplyIntraState` `CGST + SGST` vs `SupplyInterState` `IGST`) |
| `tds.go` | `@go-fintech-india-money-tax-agent` | **Stable** | `100.0%` | `0 allocs/op` | Statutory Income Tax TDS section calculator (`CalculateTDS`, `TDSRateBps`, `TDSBreakdown`, Sections `194C`/`194J`/`194H`/`194I`/`194Q`/`194O` + Section 206AA `20%` non-PAN escalation) |
| `fy.go` | `@go-fintech-india-money-tax-agent` | **Stable** | `100.0%` | `0 allocs/op` | Indian Financial Year (`April 1 – March 31`), Assessment Year, Q1–Q4 quarters, and Section 211 Advance Tax schedules |
| `aging.go` | `@go-fintech-india-money-tax-agent` | **Stable** | `100.0%` | Deterministic slices | MCA Schedule III Trade Receivables aging (`0-30`, `31-60`, `61-90`, `91-180`, `>180` days) and `AgingReport` |
| `pan.go` | `@go-fintech-india-kyc-statutory-agent` | **Stable** | `100.0%` | `0 allocs/op` | Income Tax 10-char PAN validation, 4th-char `PANEntityType` classification, surname check, and `MaskPAN` |
| `gstin.go` | `@go-fintech-india-kyc-statutory-agent` | **Stable** | `100.0%` | `0 allocs/op` | GSTN 15-char GSTIN ISO/IEC 7064 Mod 36,36 checksum verification, 37 State/UT codes (`01`–`38`, `97`, `99`), PAN extraction |
| `aadhaar.go` | `@go-fintech-india-kyc-statutory-agent` | **Stable** | `100.0%` | `0 allocs/op` | UIDAI 12-digit Aadhaar Verhoeff Dihedral $D_5$ checksum validation, `MaskAadhaar` (`XXXX-XXXX-1234`), `FormatAadhaar` |
| `ifsc.go` | `@go-fintech-india-kyc-statutory-agent` | **Stable** | `100.0%` | `0 allocs/op` | RBI 11-char IFSC validation, bank/branch code extraction, and offline RBI bank name lookup (`BankNameFromIFSC`) |
| `ifsc_resolver.go` | `@go-fintech-india-kyc-statutory-agent` | **Stable** | `100.0%` | `~71 ns/op` (offline) | `IFSCResolver` interface, zero-network `OfflineIFSCResolver`, context-aware `HTTPResolver`, and `FallbackResolver` |
| `micr.go` | `@go-fintech-india-kyc-statutory-agent` | **Stable** | `100.0%` | `0 allocs/op` | RBI 9-digit cheque clearing MICR validation and city/bank/branch code extraction |
| `phone.go` | `@go-fintech-india-kyc-statutory-agent` | **Stable** | `100.0%` | `0 allocs/op` | DoT 10-digit Indian mobile validation (`6`–`9` start digit), `NormalizeMobile`, and E.164 formatting (`+91XXXXXXXXXX`) |
| `account.go` | `@go-fintech-india-kyc-statutory-agent` | **Stable** | `100.0%` | `0 allocs/op` | RBI 9–18 digit bank account validation, all-zeros rejection, and compliant masking (`MaskBankAccount`) |
| `upi.go` | `@go-fintech-india-payments-upi-agent` | **Stable** | `100.0%` | Zero-injection parser | NPCI UPI VPA validation, PSP handle registry, and `upi://pay` deep-link intent URI generator & parser |
| `webhook.go` | `@go-fintech-india-payments-upi-agent` | **Stable** | `100.0%` | Constant-time HMAC | Timing-safe HMAC-SHA256 / SHA256 webhook signature and replay-window verifiers for Razorpay, Cashfree, and PhonePe |

---

## 5. Path to `v1.0.0`

Cutting `v1.0.0` establishes a permanent backward-compatibility guarantee under the Go module compatibility promise. Before tagging `v1.0.0`, all of the following gates must remain green across consecutive minor releases:

1. **Zero Breaking Changes Without `N + 2` Deprecation**: Complete adherence to `./scripts/check_deprecation.sh` across at least two consecutive minor release cycles with zero unannounced symbol removals.
2. **Pure Stdlib Leaf-Import Invariant**: 100% pass rate on `./scripts/check_leaf_imports.sh`, ensuring zero external dependencies in `go.mod` and zero transitive `go.sum` pollution for downstream consumers.
3. **100.0% Statement Coverage & Fuzzing Resilience**: Continuous `100.0%` statement coverage enforced by `./scripts/check_coverage.sh` across the root package and `examples/checkout_verification`, paired with zero-crash fuzz testing (`fuzz_test.go`) and race detection (`go test -race ./...`).
4. **Multi-Agent Governance & Truth Gates**: 100% pass rate across `./scripts/check_secrets.sh`, `./scripts/check_version.sh`, and `./scripts/module_agent_review.sh`.

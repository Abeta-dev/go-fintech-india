## Summary
<!-- Provide a concise summary of the statutory, financial, or engineering changes introduced by this PR. -->

## Related Issue
<!-- Fixes #(issue) or Relates to #(issue) -->

## Type of Change
- [ ] `feat`: New Indian statutory/fintech primitive or verification helper
- [ ] `fix`: Arithmetic, checksum, or boundary-condition fix
- [ ] `perf`: Zero-allocation hot-path optimization with benchmark evidence
- [ ] `docs` / `skills`: Documentation, `docs/index.html`, `llms.txt`, or `skills/*/SKILL.md` update
- [ ] `ci` / `chore`: Workflow, security gate, or release automation update

## Domain Pillar(s) Affected
- [ ] **Money & INR Formatting**: `money.go` (`int64` paise arithmetic, `Allocate`, `Split`) · `currency.go` (Lakhs/Crores, `AppendINR`) · `words.go` (`InWords` cheque printing)
- [ ] **GST & Direct Tax (`TDS`/`TCS`)**: `gst.go` (CGST/SGST/IGST/Cess, HSN/SAC) · `tax.go` (Sections `194C`/`194J`/`194H`/`194I`/`194Q` & `206C(1H)`)
- [ ] **Statutory KYC & Banking Checksums**: `aadhaar.go` (UIDAI Verhoeff $D_5$) · `pan.go` (Entity classification) · `gstin.go` (Mod-36, 37 States/UTs) · `ifsc.go` (RBI IFSC) · `micr.go` · `account.go` · `phone.go` (E.164)
- [ ] **UPI & Payment Gateway Webhooks**: `upi.go` (NPCI UPI intent URI/QR, VPA) · `webhook.go` (Razorpay, Cashfree, PhonePe HMAC-SHA256 verification)
- [ ] **Fiscal Calendar & Aging**: `fy.go` (Indian FY, Advance Tax quarters) · `aging.go` (MCA Schedule III buckets)
- [ ] **Reference Example, Docs & CI**: `examples/checkout_verification` · `docs/index.html` · `llms.txt` · `skills/` · `.github/workflows/`

## Mandatory Quality, Statutory & Security Gates
- [ ] **Zero External Dependencies**: `go.mod` strictly contains zero third-party runtime dependencies (100% Go standard library).
- [ ] `./scripts/check_secrets.sh` passed with **0 leaked credentials or verified TruffleHog findings**.
- [ ] `./scripts/check_coverage.sh` passed with **100.0% statement coverage** across both the root module and `examples/checkout_verification`.
- [ ] `./scripts/check_version.sh` and `./scripts/test_verify_release.sh` passed (`CHANGELOG.md`, `README.md`, `docs/index.html`, `gofmt`, and exported symbol inventory synchronized).
- [ ] `go test -race ./...` and `(cd examples/checkout_verification && go test -race -cover ./...)` passed with **0 data races**.
- [ ] `golangci-lint run ./...` passed with **0 issues**.
- [ ] **Statutory & Boundary Rigor**: Covered `math.MinInt64`/`math.MaxInt64` overflow edges, invalid checksums, and constant-time HMAC comparisons; added Godoc and `Example*()` tests for new exports.
- [ ] **Documentation & AI Context Sync**: Updated `README.md`, `CHANGELOG.md`, `docs/index.html`, `docs/llms.txt` (`llms.txt`), `AGENTS.md`, and `skills/*/SKILL.md` where applicable.

## Description
<!-- Provide a concise summary of the changes introduced by this PR. -->

## Related Issue
<!-- Fixes #(issue) or Relates to #(issue) -->

## Module(s) Affected
- [ ] `money` (Paise arithmetic, Split, Allocate, SQL/JSON)
- [ ] `currency` (Lakhs/Crores formatting & parsing)
- [ ] `words` (Rupees to Words, Cheque Printing)
- [ ] `fy` (Financial Year, Advance Tax Schedule)
- [ ] `aging` (MCA Schedule III Receivables Aging)
- [ ] `aadhaar` (Verhoeff D5 Checksum, Masking)
- [ ] `pan` (Entity Parsing, Format, Masking)
- [ ] `gstin` (Mod-36 Checksum, State Code Lookup)
- [ ] `ifsc` (RBI IFSC Validation, Bank Code)
- [ ] `micr` (MICR Transit Code Parsing)
- [ ] `upi` (NPCI UPI Intent URI/QR, VPA Validation)
- [ ] `phone` (Mobile Validation, E.164 Formatting)
- [ ] `account` (Bank Account Validation, Masking)
- [ ] CI / Automation / Documentation

## Quality & Governance Checklist
- [ ] `make fmt-check` passed with 0 unformatted files (`gofmt -l .`).
- [ ] `make lint` passed with 0 warnings or errors (`golangci-lint run ./...`).
- [ ] `make test` passed all unit tests.
- [ ] `make race` passed with 0 data races (`go test -race ./...`).
- [ ] `./scripts/check_version.sh` passed with exit code 0.
- [ ] `./scripts/check_coverage.sh` passed with statement coverage $\ge 90\%$.
- [ ] Godoc comments added for all newly exported identifiers.
- [ ] Test cases cover all boundary conditions, checksum edge-cases, and invalid inputs.
- [ ] Zero external dependencies maintained (standard library only).

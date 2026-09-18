# Contributing to go-fintech-india

Thank you for your interest in contributing to `go-fintech-india`! We welcome community contributions, bug reports, optimizations, and feature enhancements.

This project is a high-reliability financial, tax, and identity toolkit designed for Indian transactional backends. To ensure the highest levels of correctness and security, all contributions must adhere to the standards outlined below.

---

## Core Principles

1. **Zero External Dependencies**:
   - The entire package strictly uses the Go standard library. No third-party modules or external dependencies may be introduced.
2. **Integer Paise Financial Arithmetic**:
   - Never represent currency values using raw floating-point types (`float32`/`float64`).
   - All monetary math must use `int64` paise representations (1 INR = 100 paise) via `fintechin.Money`.
   - Remainder-safe allocations and splits must guarantee exact conservation of value.
3. **Statutory Integrity**:
   - Checksums, masks, and parsing logic must strictly conform to official statutory specifications (UIDAI Verhoeff D5, GSTN Mod-36, RBI IFSC/MICR, NPCI UPI, MCA Schedule III).
4. **Zero-Allocation & High Performance**:
   - Hot paths (checksum verification, formatting, parsing) should minimize heap allocations and pass microbenchmarks.
5. **Quality Truth Gates**:
   - Statement coverage must remain $\ge 90\%$ at all times.
   - Code must pass `golangci-lint`, the Go race detector (`-race`), and `gofmt -l`.

---

## Development Workflow

### Prerequisites
- Go 1.24+ installed
- `golangci-lint` (v1.60+ or v2.x)

### Local Setup
```bash
git clone https://github.com/Abeta-dev/go-fintech-india.git
cd go-fintech-india
```

### Essential Make Commands
```bash
# Verify formatting, run linter, run tests with race detector, and verify coverage gates
make all

# Format all code
make fmt

# Check formatting without modifying
make fmt-check

# Run unit tests
make test

# Run tests with race detector
make race

# Run benchmarks with allocation statistics
make bench

# Run linter
make lint

# Run coverage gate script (enforces >= 90% statement coverage)
make coverage
```

### Release verification

A release tag is immutable. Complete the release in this order:

1. Update the version references, changelog, and release checks in a reviewed release commit; run the local quality gates.
2. Create and push the immutable `vX.Y.Z` (or valid SemVer prerelease such as `vX.Y.Z-rc.1`) tag that points to that commit. The workflow cannot verify a tag or its Go-proxy artifact before this prerequisite exists.
3. Let the tag-triggered release workflow run. It exports the tag before checking documentation, verifies the local/remote tag objects, retries fresh Go-proxy resolution with bounded exponential backoff, writes `release-manifest.json`, and uploads it both as a workflow artifact and a GitHub release asset.

For an already-pushed tag, use the same checkout and run:

```bash
GIT_TAG=vX.Y.Z ./scripts/verify_release.sh vX.Y.Z
```

The verifier records module sums and source archive hashes in the ignored manifest. The workflow builds an isolated consumer using the published module; it never replaces that dependency with the checkout. Re-running a release only replaces that release's verification manifest asset and never deletes or recreates the GitHub release.

---

## Testing & Quality Checklist

Before submitting a pull request, ensure:
- [ ] `make all` passes with zero errors and zero warnings.
- [ ] `make test-race` passes without any data races.
- [ ] `./scripts/check_version.sh` passes successfully.
- [ ] `./scripts/check_coverage.sh` passes with $\ge 90\%$ statement coverage.
- [ ] All newly introduced exported types, functions, and methods have godoc comments.
- [ ] Test table cases cover all boundary conditions, invalid inputs, edge-cases, and checksum failures.
- [ ] Any public API changes are documented in `CHANGELOG.md` under `### Added` or `### Changed`.

---

## Pull Request Guidelines

1. **Descriptive Commits**: Write clear, conventional commit messages (e.g. `feat(gstin): add Mod-36 checksum validator`).
2. **Atomic Changes**: Keep PRs scoped to a single logical improvement or fix.
3. **Issue Reference**: Link the PR to a related GitHub issue (e.g. `Fixes #12`).
4. **CI Compliance**: All GitHub Actions CI checks must pass before review.

Thank you for helping build high-reliability fintech infrastructure for India!

# Release Procedure & Guardrails

This document establishes the mandatory, step-by-step release procedure for `go-fintech-india` (`github.com/abeta-dev/go-fintech-india`).

---

## 1. Immutable Release Principles

Go modules rely on public proxy and checksum services (`proxy.golang.org` and `sum.golang.org`).

- **Never Overwrite or Move Tags**: Once a Git tag is pushed to remote, Go module proxies immediately cache the source archive and its cryptographic checksum (`h1:...`). Retagging the same version name to a different commit will break downstream builds with checksum mismatch errors.
- **Monotonic Releases Only**: If a release fails or needs corrections, a new monotonic SemVer version (e.g. `v0.3.1`) must be published.
- **Annotated Tags Only**: All release tags must be created as annotated Git tags (`git tag -a vX.Y.Z -m "Release vX.Y.Z"`). Never use lightweight tags for official releases.

---

## 2. Release Guardrails & Automated Verification Gates

To prevent releasing unverified code, leaked credentials, or out-of-sync documentation, multiple verification gates are enforced locally and in CI:

1. **Secret & Credential Leak Gate** (`./scripts/check_secrets.sh`):
   - Scans tracked files for forbidden key/certificate/`.env` files, PEM private key headers, cloud/VCS/payment gateway tokens, and embedded database DSN credentials.
   - Runs TruffleHog verified git-history secret scanning (`trufflehog git file://. --only-verified --fail --no-update`).
2. **Statement Coverage Gate** (`./scripts/check_coverage.sh`):
   - Enforces statement coverage (`100.0%` across the repository and `examples/checkout_verification`) and updates `.github/badges/coverage.json`.
3. **Version, Formatting & Exported Symbol Truth Gate** (`./scripts/check_version.sh`):
   - Verifies exact SemVer parity across `CHANGELOG.md`, `README.md` (header and `go get` snippet), `docs/index.html`, and `GIT_TAG` (when set).
   - Verifies `gofmt` formatting across all `.go` files.
   - Verifies that every exported Go symbol listed under `### Added` in `CHANGELOG.md` exists in non-test package source files.
   - Invokes `./scripts/check_secrets.sh`.
4. **Immutable Release Verifier & Test Harness** (`./scripts/verify_release.sh` & `./scripts/test_verify_release.sh`):
   - Verifies local and remote tag objects match `HEAD`.
   - Resolves `github.com/abeta-dev/go-fintech-india@vX.Y.Z` from the Go module proxy using an isolated `GOMODCACHE` with bounded exponential backoff.
   - Records `module_sum`, `go_mod_sum`, `source_archive_sha256`, and `go_mod_sha256` into `release-manifest.json`.

---

## 3. Step-by-Step Release Checklist

### Step 1: Prepare Release PR on a Feature Branch
1. Create a release branch from `main`:
   ```bash
   git checkout -b release/vX.Y.Z
   ```
2. Update documentation and release metadata:
   - **`CHANGELOG.md`**: Add `## [X.Y.Z] - YYYY-MM-DD` at the top of released versions. Document all changes under `Added`, `Changed`, or `Fixed`.
   - **`README.md`**: Update header `# go-fintech-india · vX.Y.Z` and `go get github.com/abeta-dev/go-fintech-india@vX.Y.Z`.
   - **`docs/index.html`**: Update version badges, metadata, and `go get github.com/abeta-dev/go-fintech-india@vX.Y.Z`.
   - **`AGENTS.md` & `skills/*/SKILL.md`**: Synchronize version references and copy updated skills to `~/.gemini/config/skills/`.
   - **`SECURITY.md`**: Verify that the active version series is marked **Active / Current**.
   - **`scripts/check_version.sh`**: Update `DEFAULT_EXPECTED_VER="X.Y.Z"`.
3. Run all local verification gates:
   ```bash
   ./scripts/check_secrets.sh
   ./scripts/check_coverage.sh
   ./scripts/check_version.sh
   ./scripts/test_verify_release.sh
   go test -race ./...
   (cd examples/checkout_verification && go test -race -cover ./...)
   rm -f coverage.out
   ```
4. Commit and push the branch, then open a Pull Request into `main` and wait for all 7 CI checks (`CI Quality Gate`) to pass.

### Step 2: Merge PR & Create Immutable Annotated Tag on `main`
1. Merge the Release PR into `main`.
2. Switch to `main` and pull the merged commit:
   ```bash
   git checkout main
   git pull upstream main
   ```
3. Create and push the annotated SemVer tag:
   ```bash
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   git push upstream vX.Y.Z
   ```

### Step 3: Verify CI Release Workflow & Manifest
- The tag push triggers `.github/workflows/release.yml`, which runs `./scripts/verify_release.sh vX.Y.Z`, builds an isolated consumer via `go get github.com/abeta-dev/go-fintech-india@vX.Y.Z`, and publishes the GitHub Release with `release-manifest.json` attached.
- To re-verify an already-pushed tag locally:
  ```bash
  GIT_TAG=vX.Y.Z ./scripts/verify_release.sh vX.Y.Z
  ```

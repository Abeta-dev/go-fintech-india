#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# Version, Symbol & Documentation Truth Gate (go-fintech-india)
# Guarantees that:
#   1. README.md, CHANGELOG.md, and git tags never drift out of sync.
#   2. Code formatting is clean across all files (gofmt).
#   3. Module import tags strictly match @v<EXPECTED_VER>.
#   4. Every exported symbol documented in CHANGELOG.md under "### Added" exists.
# ==============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

DEFAULT_EXPECTED_VER="0.2.3"
GIT_TAG_REF="${GIT_TAG:-}"
if [ -z "$GIT_TAG_REF" ]; then
  GIT_TAG_REF=$(git describe --tags --exact-match 2>/dev/null || true)
fi

# A release workflow supplies GIT_TAG. Retain the checkout's documented version
# for ordinary contributor checks, but let tagged prereleases verify their exact
# SemVer version rather than forcing a stable tag.
EXPECTED_VER="${DEFAULT_EXPECTED_VER}"
if [ -n "$GIT_TAG_REF" ]; then
  EXPECTED_VER="${GIT_TAG_REF#v}"
fi
SEMVER_PATTERN='[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?'

CHANGELOG_VER=$(grep -E "^## \[${SEMVER_PATTERN}\]" CHANGELOG.md | head -n1 | sed -E "s/^## \[(${SEMVER_PATTERN})\].*/\1/")
README_HEADER_VER=$(grep -E "^# go-fintech-india · v${SEMVER_PATTERN}" README.md | head -n1 | sed -E "s/^# go-fintech-india · v(${SEMVER_PATTERN}).*/\1/")
README_GET_VER=$(grep -E "go get github.com/umesh0492/go-fintech-india@v${SEMVER_PATTERN}" README.md | head -n1 | sed -E "s/.*go-fintech-india@v(${SEMVER_PATTERN}).*/\1/")

echo "========================================================"
echo "🔒 Verifying Version Synchronization (go-fintech-india)"
echo "   - Expected Version: v$EXPECTED_VER"
echo "   - CHANGELOG.md:     v$CHANGELOG_VER"
echo "   - README.md Header: v$README_HEADER_VER"
echo "   - README.md go get: v$README_GET_VER"

if [ -n "$GIT_TAG_REF" ]; then
  echo "   - Git Tag:          $GIT_TAG_REF"
fi
echo "========================================================"

if [ "$CHANGELOG_VER" != "$EXPECTED_VER" ]; then
  echo "❌ Error: CHANGELOG.md version (v$CHANGELOG_VER) does not match expected version (v$EXPECTED_VER)"
  exit 1
fi

if [ "$CHANGELOG_VER" != "$README_HEADER_VER" ]; then
  echo "❌ Error: Version mismatch between CHANGELOG.md (v$CHANGELOG_VER) and README.md header (v$README_HEADER_VER)"
  exit 1
fi

if [ "$CHANGELOG_VER" != "$README_GET_VER" ]; then
  echo "❌ Error: Version mismatch between CHANGELOG.md (v$CHANGELOG_VER) and README.md go get (v$README_GET_VER)"
  exit 1
fi

if [ -n "$GIT_TAG_REF" ]; then
  STRIPPED_TAG=$(echo "$GIT_TAG_REF" | sed -E 's/^v//')
  if [ "$STRIPPED_TAG" != "$CHANGELOG_VER" ]; then
    echo "❌ Error: Git tag ($GIT_TAG_REF) does not match CHANGELOG.md version (v$CHANGELOG_VER)"
    exit 1
  fi
  echo "✅ Git tag matches repository version ($GIT_TAG_REF)."
fi

echo "========================================================"
echo "🎨 Verifying Code Formatting (gofmt)"
echo "========================================================"

GO_FILES=$(git ls-files '*.go' 2>/dev/null || true)
if [ -z "$GO_FILES" ]; then
  GO_FILES=$(find . -maxdepth 1 -name '*.go' -not -name '.*')
fi

if [ -n "$GO_FILES" ]; then
  UNFORMATTED=$(gofmt -l $GO_FILES)
else
  UNFORMATTED=""
fi

if [ -n "$UNFORMATTED" ]; then
  echo "❌ Error: Unformatted files detected by gofmt:"
  echo "$UNFORMATTED"
  echo "Run 'make fmt' to fix."
  exit 1
fi
echo "✅ Code formatting is clean across all Go files."

echo "========================================================"
echo "🔍 Verifying CHANGELOG [### Added] Exported Symbols"
echo "========================================================"

NON_TEST_GO_FILES=$(find . -maxdepth 1 -name '*.go' -not -name '*_test.go' -not -name '.*')
CHANGELOG_FILE="CHANGELOG.md"
IN_ADDED=0
FAILED_SYMBOLS=0
CHECKED_SYMBOLS=0

while IFS= read -r line; do
  if [[ "${line}" =~ ^"### Added" ]]; then
    IN_ADDED=1
    continue
  fi
  if [ "${IN_ADDED}" -eq 1 ] && [[ "${line}" =~ ^"### " || "${line}" =~ ^"## " ]]; then
    IN_ADDED=0
  fi
  if [ "${IN_ADDED}" -eq 1 ] && [[ "${line}" =~ ^"- " ]]; then
    # Extract all symbols in backticks
    RAW_SYMBOLS=$(echo "${line}" | grep -oE '`[^`]+`' | tr -d '`' || true)
    for sym in ${RAW_SYMBOLS}; do
      [ -z "${sym}" ] && continue

      # Filter out standard non-exported or external references
      if [[ "${sym}" =~ "/" || "${sym}" =~ ^"http" || "${sym}" =~ ^"sql." || "${sym}" =~ ^"driver." || "${sym}" =~ ^"json." ]]; then
        continue
      fi
      if [[ "${sym}" =~ "-" || "${sym}" =~ "=" || "${sym}" =~ "@" || "${sym}" =~ " " ]]; then
        continue
      fi
      # Only check capitalized exported Go symbols
      if [[ ! "${sym}" =~ ^[A-Z] ]]; then
        continue
      fi

      CHECKED_SYMBOLS=$((CHECKED_SYMBOLS + 1))

      # Search strictly in non-test Go source files
      if ! grep -q -w "${sym}" $NON_TEST_GO_FILES 2>/dev/null; then
        echo "❌ Error: Exported symbol '${sym}' referenced in CHANGELOG.md does not exist in package source files"
        FAILED_SYMBOLS=1
      else
        echo "   - ${sym}: verified ✅"
      fi
    done
  fi
done < "${CHANGELOG_FILE}"

if [ "${FAILED_SYMBOLS}" -ne 0 ]; then
  echo "❌ Error: CHANGELOG symbol verification failed!"
  exit 1
fi
echo "✅ All CHANGELOG exported symbols exist in the repository (${CHECKED_SYMBOLS} symbols verified)."

echo "========================================================"
echo "✅ All versions, format, and symbols are strictly verified!"
echo "========================================================"

#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# Strict Deprecation Window Gate (N + 2 Minor Releases from Day 1)
#
# Guarantees that:
#   1. All non-test *.go files are scanned for '// Deprecated:' annotations.
#   2. Every '// Deprecated:' line strictly matches:
#      // Deprecated: Since vX.Y.Z (remove in >= vA.B.C). Use <Replacement> instead.
#   3. The removal version satisfies (A > X) || (A == X && B >= Y + 2)
#      (minimum N + 2 minor releases deprecation window).
#   4. Any '// Deprecated: Since vX.Y.Z (remove in >= vA.B.C)' annotation present
#      in the previous release git tag that was removed in HEAD is only removed
#      when the current version in CHANGELOG.md is >= vA.B.C.
# ==============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

SEMVER_REGEX='[0-9]+\.[0-9]+\.[0-9]+'
CURRENT_VER=$(grep -E "^## \[${SEMVER_REGEX}" CHANGELOG.md | head -n1 | sed -E "s/^## \[(${SEMVER_REGEX})\].*/\1/")

if [ -z "${CURRENT_VER}" ]; then
  echo "❌ Error: Could not determine current version from CHANGELOG.md"
  exit 1
fi

IFS='.' read -r CUR_MAJ CUR_MIN CUR_PAT <<< "${CURRENT_VER}"

echo "========================================================"
echo "⏳ Verifying Strict N+2 Minor-Release Deprecation Gate"
echo "   Current CHANGELOG Version: v${CURRENT_VER}"
echo "   Required Format: // Deprecated: Since vX.Y.Z (remove in >= vA.B.C). Use <Replacement> instead."
echo "========================================================"

NON_TEST_GO_FILES=$(find . -name '*.go' ! -name '*_test.go' ! -path './.git/*' | sort)

FAILURES=0
DEPRECATED_COUNT=0
DEPRECATION_PATTERN='^// Deprecated: Since v([0-9]+)\.([0-9]+)\.([0-9]+) \(remove in >= v([0-9]+)\.([0-9]+)\.([0-9]+)\)\. Use .+ instead\.$'

for go_file in ${NON_TEST_GO_FILES}; do
  MATCHES=$(grep -nE '^[[:space:]]*//[[:space:]]*Deprecated:' "${go_file}" || true)
  [ -z "${MATCHES}" ] && continue

  while IFS= read -r match_line; do
    [ -z "${match_line}" ] && continue
    DEPRECATED_COUNT=$((DEPRECATED_COUNT + 1))

    LINENO_PART="${match_line%%:*}"
    RAW_COMMENT="${match_line#*:}"
    TRIMMED_COMMENT="$(echo "${RAW_COMMENT}" | sed -E 's/^[[:space:]]+//; s/[[:space:]]+$//')"

    if [[ ! "${TRIMMED_COMMENT}" =~ ${DEPRECATION_PATTERN} ]]; then
      echo "❌ Error: Malformed deprecation notice at ${go_file}:${LINENO_PART}"
      echo "   Found:    ${TRIMMED_COMMENT}"
      echo "   Expected: // Deprecated: Since vX.Y.Z (remove in >= vA.B.C). Use <Replacement> instead."
      FAILURES=$((FAILURES + 1))
      continue
    fi

    SINCE_X="${BASH_REMATCH[1]}"
    SINCE_Y="${BASH_REMATCH[2]}"
    SINCE_Z="${BASH_REMATCH[3]}"
    REM_A="${BASH_REMATCH[4]}"
    REM_B="${BASH_REMATCH[5]}"
    REM_C="${BASH_REMATCH[6]}"

    MIN_MINOR=$((SINCE_Y + 2))
    if ! (( REM_A > SINCE_X || (REM_A == SINCE_X && REM_B >= MIN_MINOR) )); then
      echo "❌ Error: Deprecation window violation at ${go_file}:${LINENO_PART}"
      echo "   Since v${SINCE_X}.${SINCE_Y}.${SINCE_Z} requires removal version >= v${SINCE_X}.${MIN_MINOR}.0 (N + 2 minor releases), got v${REM_A}.${REM_B}.${REM_C}"
      FAILURES=$((FAILURES + 1))
      continue
    fi

    echo "   - ${go_file}:${LINENO_PART}: valid N+2 window (Since v${SINCE_X}.${SINCE_Y}.${SINCE_Z} -> remove in >= v${REM_A}.${REM_B}.${REM_C}) ✅"
  done <<EOF
${MATCHES}
EOF
done

# 4. Check previous release tag in git (if available) for premature removal
PREV_TAG=$(git tag -l 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname 2>/dev/null | head -n1 || true)
if [ -n "${PREV_TAG}" ]; then
  echo "   - Checking against previous git release tag: ${PREV_TAG}"
  PREV_DEPRECATIONS=$(git grep -hE '^[[:space:]]*// Deprecated: Since v[0-9]+\.[0-9]+\.[0-9]+ \(remove in >= v[0-9]+\.[0-9]+\.[0-9]+\)' "${PREV_TAG}" -- '*.go' ':!*_test.go' 2>/dev/null || true)

  if [ -n "${PREV_DEPRECATIONS}" ]; then
    while IFS= read -r prev_raw; do
      [ -z "${prev_raw}" ] && continue
      PREV_TRIMMED="$(echo "${prev_raw}" | sed -E 's/^[[:space:]]+//; s/[[:space:]]+$//')"

      if ! grep -Fq "${PREV_TRIMMED}" ${NON_TEST_GO_FILES} 2>/dev/null; then
        if [[ "${PREV_TRIMMED}" =~ ${DEPRECATION_PATTERN} ]]; then
          REM_A="${BASH_REMATCH[4]}"
          REM_B="${BASH_REMATCH[5]}"
          REM_C="${BASH_REMATCH[6]}"

          if ! (( CUR_MAJ > REM_A || (CUR_MAJ == REM_A && CUR_MIN > REM_B) || (CUR_MAJ == REM_A && CUR_MIN == REM_B && CUR_PAT >= REM_C) )); then
            echo "❌ Error: Premature removal of deprecated symbol from ${PREV_TAG}!"
            echo "   Removed Notice:  ${PREV_TRIMMED}"
            echo "   Current Version: v${CURRENT_VER} (must be >= v${REM_A}.${REM_B}.${REM_C})"
            FAILURES=$((FAILURES + 1))
          else
            echo "   - Completed deprecation window removal allowed in v${CURRENT_VER} (>= v${REM_A}.${REM_B}.${REM_C}) ✅"
          fi
        fi
      fi
    done <<EOF
${PREV_DEPRECATIONS}
EOF
  fi
fi

if [ "${FAILURES}" -ne 0 ]; then
  echo "========================================================"
  echo "❌ Error: Deprecation window gate failed with ${FAILURES} violation(s)!"
  echo "========================================================"
  exit 1
fi

echo "   - Active deprecations scanned: ${DEPRECATED_COUNT} (0 violations)"
echo "========================================================"
echo "✅ Strict N+2 Minor-Release Deprecation Window Contract verified!"
echo "========================================================"

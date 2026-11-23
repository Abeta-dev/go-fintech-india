#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# Day-1 Pure Stdlib Leaf-Import Isolation Gate (go-fintech-india)
#
# Guarantees that:
#   1. Every direct, test, external-test, and transitive dependency across
#      `.` and `./examples/checkout_verification` is either
#      `github.com/abeta-dev/go-fintech-india` or a Go standard library package
#      (zero import paths whose first path element contains a dot `.`).
#   2. Importing `github.com/abeta-dev/go-fintech-india` in an isolated
#      temporary consumer `go.mod` sandbox produces zero `go.sum` entries
#      (no `go.sum` file or 0-byte `go.sum`), proving zero transitive
#      third-party dependencies are introduced into any consumer binary.
# ==============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

MODULE_PATH="github.com/abeta-dev/go-fintech-india"
PACKAGES=(".")
if [ -d "./examples/checkout_verification" ]; then
  PACKAGES+=("./examples/checkout_verification")
fi

echo "========================================================"
echo "🍃 Verifying Day-1 Pure Stdlib Leaf-Import Invariant"
echo "   Module:   ${MODULE_PATH}"
echo "   Packages: ${PACKAGES[*]}"
echo "========================================================"

# 1. Verify direct, test, and xtest imports via go list
DIRECT_IMPORTS=$(go list -f '{{join .Imports "\n"}}{{"\n"}}{{join .TestImports "\n"}}{{"\n"}}{{join .XTestImports "\n"}}' "${PACKAGES[@]}" | sed '/^$/d' | sort -u)
TRANSITIVE_DEPS=$(go list -deps "${PACKAGES[@]}" | sed '/^$/d' | sort -u)

COMBINED_IMPORTS=$(printf "%s\n%s\n" "${DIRECT_IMPORTS}" "${TRANSITIVE_DEPS}" | sed '/^$/d' | sort -u)

EXTERNAL_LEAKS=0
CHECKED_COUNT=0

while IFS= read -r pkg; do
  [ -z "${pkg}" ] && continue
  CHECKED_COUNT=$((CHECKED_COUNT + 1))

  if [ "${pkg}" = "${MODULE_PATH}" ] || [[ "${pkg}" == "${MODULE_PATH}/"* ]]; then
    continue
  fi

  FIRST_ELEM="${pkg%%/*}"
  if [[ "${FIRST_ELEM}" == *.* ]]; then
    echo "❌ Error: External non-stdlib import detected: '${pkg}'"
    EXTERNAL_LEAKS=$((EXTERNAL_LEAKS + 1))
  fi
done <<EOF
${COMBINED_IMPORTS}
EOF

if [ "${EXTERNAL_LEAKS}" -ne 0 ]; then
  echo "❌ Error: Found ${EXTERNAL_LEAKS} non-stdlib import path(s) violating the pure stdlib invariant!"
  exit 1
fi
echo "   - Direct, test & transitive imports (${CHECKED_COUNT} packages): 100% pure Go stdlib ✅"

# 2. Verify isolated consumer sandbox generates no go.sum (or 0-byte go.sum)
TMP_DIR=$(mktemp -d)
trap 'rm -rf "${TMP_DIR}"' EXIT

SANDBOX_DIR="${TMP_DIR}/consumer_sandbox"
mkdir -p "${SANDBOX_DIR}"

(
  cd "${SANDBOX_DIR}"
  go mod init "verify_leaf_fintech_india" >/dev/null 2>&1
  cat << 'EOF' > main.go
package main

import (
	_ "github.com/abeta-dev/go-fintech-india"
)

func main() {}
EOF
  go mod edit -replace "${MODULE_PATH}=${ROOT_DIR}" >/dev/null 2>&1
  if ! TIDY_ERR=$(go mod tidy 2>&1); then
    echo "❌ Error: 'go mod tidy' failed in consumer sandbox: ${TIDY_ERR}"
    exit 1
  fi

  if [ -s "go.sum" ]; then
    echo "❌ Error: Consumer sandbox generated a non-empty go.sum:"
    cat go.sum
    exit 1
  fi
)

echo "   - Isolated consumer sandbox: clean leaf import ✅ (no go.sum / 0 bytes, zero transitive dependencies)"
echo "========================================================"
echo "✅ Pure Stdlib Leaf-Import Invariant verified across all packages!"
echo "========================================================"

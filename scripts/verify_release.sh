#!/usr/bin/env bash
set -euo pipefail

# Verify that a tagged checkout is an immutable, resolvable Go module release and
# write the evidence needed to reproduce the published artifact.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

TAG="${1:-${GIT_TAG:-}}"
MANIFEST_PATH="${RELEASE_MANIFEST_PATH:-release-manifest.json}"
GO_CMD="${GO_CMD:-go}"
RETRY_ATTEMPTS="${RELEASE_PROXY_RETRY_ATTEMPTS:-4}"
RETRY_INITIAL_DELAY_SECONDS="${RELEASE_PROXY_RETRY_INITIAL_DELAY_SECONDS:-2}"
RETRY_MAX_DELAY_SECONDS="${RELEASE_PROXY_RETRY_MAX_DELAY_SECONDS:-30}"
SLEEP_CMD="${RELEASE_PROXY_SLEEP_COMMAND:-sleep}"
ATTEMPT_MODULE_CACHE_DIR=""
DOWNLOAD_ERROR=""
VERIFICATION_STATUS="failed"
FAILURE_STAGE="initialization"

write_manifest() {
  cat >"${MANIFEST_PATH}" <<EOF
{
  "schema_version": 1,
  "verification_status": "${VERIFICATION_STATUS}",
  "failure_stage": "${FAILURE_STAGE}",
  "module": "${MODULE_PATH:-}",
  "version": "${TAG}",
  "tag_object": "${LOCAL_TAG_OBJECT:-}",
  "commit": "${LOCAL_COMMIT:-}",
  "remote_tag_object": "${REMOTE_TAG_OBJECT:-}",
  "remote_commit": "${REMOTE_COMMIT:-}",
  "module_sum": "${MODULE_SUM:-}",
  "go_mod_sum": "${MODULE_GO_MOD_SUM:-}",
  "source_archive_sha256": "${ARCHIVE_SHA256:-}",
  "go_mod_sha256": "${GO_MOD_SHA256:-}"
}
EOF
}

remove_attempt_module_cache() {
  if [ -n "${ATTEMPT_MODULE_CACHE_DIR}" ]; then
    # Downloaded Go toolchains are read-only in the module cache.
    chmod -R u+w "${ATTEMPT_MODULE_CACHE_DIR}" 2>/dev/null || true
    rm -rf "${ATTEMPT_MODULE_CACHE_DIR}" || true
    ATTEMPT_MODULE_CACHE_DIR=""
  fi
}

cleanup() {
  if [ -n "${DOWNLOAD_ERROR}" ]; then
    rm -f "${DOWNLOAD_ERROR}"
  fi
  remove_attempt_module_cache
}

on_exit() {
  local status=$?
  trap - EXIT
  cleanup
  if [ "${status}" -ne 0 ]; then
    VERIFICATION_STATUS="failed"
    if ! write_manifest; then
      echo "warning: could not write failed release manifest to ${MANIFEST_PATH}" >&2
    fi
  fi
  exit "${status}"
}
trap on_exit EXIT

if [[ ! "${TAG}" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$ ]]; then
  echo "usage: $0 vMAJOR.MINOR.PATCH[-PRERELEASE]" >&2
  exit 2
fi
if [[ ! "${RETRY_ATTEMPTS}" =~ ^[1-9][0-9]*$ ]] ||
  [[ ! "${RETRY_INITIAL_DELAY_SECONDS}" =~ ^[0-9]+$ ]] ||
  [[ ! "${RETRY_MAX_DELAY_SECONDS}" =~ ^[0-9]+$ ]]; then
  echo "release proxy retry settings must be non-negative integers and attempts must be at least one" >&2
  exit 2
fi

# check_version derives the expected documentation version from GIT_TAG when set.
export GIT_TAG="${TAG}"
FAILURE_STAGE="version"
./scripts/check_version.sh

FAILURE_STAGE="local_tag"
LOCAL_TAG_OBJECT=$(git rev-parse "refs/tags/${TAG}")
LOCAL_COMMIT=$(git rev-parse "${TAG}^{}")
HEAD_COMMIT=$(git rev-parse HEAD)
if [ "${LOCAL_COMMIT}" != "${HEAD_COMMIT}" ]; then
  echo "release tag ${TAG} resolves to ${LOCAL_COMMIT}, not HEAD ${HEAD_COMMIT}" >&2
  exit 1
fi

FAILURE_STAGE="remote_tag"
REMOTE_TAG_OBJECT=$(git ls-remote --tags origin "refs/tags/${TAG}" | awk '{print $1}')
REMOTE_COMMIT=$(git ls-remote --tags origin "refs/tags/${TAG}^{}" | awk '{print $1}')
if [ -z "${REMOTE_COMMIT}" ]; then
  REMOTE_COMMIT="${REMOTE_TAG_OBJECT}"
fi
if [ -z "${REMOTE_TAG_OBJECT}" ] || [ "${REMOTE_COMMIT}" != "${LOCAL_COMMIT}" ]; then
  echo "remote tag ${TAG} does not match local release commit" >&2
  exit 1
fi

FAILURE_STAGE="module_identity"
MODULE_PATH=$("${GO_CMD}" list -m -f '{{.Path}}')

FAILURE_STAGE="module_proxy"
# Force a fresh module-cache lookup for every attempt. A disposable GOMODCACHE
# prevents a cached negative lookup from masking a tag that has just reached the
# proxy, while the normal Go toolchain cache remains available.
DOWNLOAD_ERROR=$(mktemp)
DELAY_SECONDS="${RETRY_INITIAL_DELAY_SECONDS}"
for ((attempt = 1; attempt <= RETRY_ATTEMPTS; attempt++)); do
  ATTEMPT_MODULE_CACHE_DIR=$(mktemp -d)
  if DOWNLOAD_JSON=$(GOWORK=off GOMODCACHE="${ATTEMPT_MODULE_CACHE_DIR}" "${GO_CMD}" mod download -json "${MODULE_PATH}@${TAG}" 2>"${DOWNLOAD_ERROR}"); then
    remove_attempt_module_cache
    break
  fi
  remove_attempt_module_cache

  echo "Go module proxy resolution attempt ${attempt}/${RETRY_ATTEMPTS} for ${MODULE_PATH}@${TAG} failed:" >&2
  cat "${DOWNLOAD_ERROR}" >&2
  if [ "${attempt}" -eq "${RETRY_ATTEMPTS}" ]; then
    echo "Go module proxy resolution failed after ${RETRY_ATTEMPTS} attempts; final diagnostic is shown above." >&2
    exit 1
  fi

  echo "Retrying Go module proxy resolution in ${DELAY_SECONDS}s..." >&2
  "${SLEEP_CMD}" "${DELAY_SECONDS}"
  if [ "${DELAY_SECONDS}" -lt "${RETRY_MAX_DELAY_SECONDS}" ]; then
    DELAY_SECONDS=$((DELAY_SECONDS * 2))
    if [ "${DELAY_SECONDS}" -gt "${RETRY_MAX_DELAY_SECONDS}" ]; then
      DELAY_SECONDS="${RETRY_MAX_DELAY_SECONDS}"
    fi
  fi
done
rm -f "${DOWNLOAD_ERROR}"
DOWNLOAD_ERROR=""

MODULE_VERSION=$(printf '%s\n' "${DOWNLOAD_JSON}" | sed -n 's/^[[:space:]]*"Version": "\([^"]*\)",/\1/p' | head -n1)
MODULE_SUM=$(printf '%s\n' "${DOWNLOAD_JSON}" | sed -n 's/^[[:space:]]*"Sum": "\([^"]*\)",/\1/p' | head -n1)
MODULE_GO_MOD_SUM=$(printf '%s\n' "${DOWNLOAD_JSON}" | sed -n 's/^[[:space:]]*"GoModSum": "\([^"]*\)".*/\1/p' | head -n1)
if [ "${MODULE_VERSION}" != "${TAG}" ] || [ -z "${MODULE_SUM}" ] || [ -z "${MODULE_GO_MOD_SUM}" ]; then
  echo "module proxy did not resolve ${MODULE_PATH}@${TAG} with integrity sums" >&2
  exit 1
fi

FAILURE_STAGE="source_hash"
ARCHIVE_SHA256=$(git archive --format=tar "${TAG}" | sha256sum | awk '{print $1}')
GO_MOD_SHA256=$(git show "${TAG}:go.mod" | sha256sum | awk '{print $1}')

VERIFICATION_STATUS="verified"
FAILURE_STAGE=""
write_manifest
echo "release manifest verified and written to ${MANIFEST_PATH}"

#!/usr/bin/env bash
set -euo pipefail

# Verify that a tagged checkout is an immutable, resolvable Go module release and
# write the evidence needed to reproduce the published artifact.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

TAG="${1:-${GIT_TAG:-}}"
MANIFEST_PATH="${RELEASE_MANIFEST_PATH:-release-manifest.json}"

if [[ ! "${TAG}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "usage: $0 vMAJOR.MINOR.PATCH" >&2
  exit 2
fi

./scripts/check_version.sh

LOCAL_TAG_OBJECT=$(git rev-parse "refs/tags/${TAG}")
LOCAL_COMMIT=$(git rev-parse "${TAG}^{}")
HEAD_COMMIT=$(git rev-parse HEAD)
if [ "${LOCAL_COMMIT}" != "${HEAD_COMMIT}" ]; then
  echo "release tag ${TAG} resolves to ${LOCAL_COMMIT}, not HEAD ${HEAD_COMMIT}" >&2
  exit 1
fi

REMOTE_TAG_OBJECT=$(git ls-remote --tags origin "refs/tags/${TAG}" | awk '{print $1}')
REMOTE_COMMIT=$(git ls-remote --tags origin "refs/tags/${TAG}^{}" | awk '{print $1}')
if [ -z "${REMOTE_COMMIT}" ]; then
  REMOTE_COMMIT="${REMOTE_TAG_OBJECT}"
fi
if [ -z "${REMOTE_TAG_OBJECT}" ] || [ "${REMOTE_COMMIT}" != "${LOCAL_COMMIT}" ]; then
  echo "remote tag ${TAG} does not match local release commit" >&2
  exit 1
fi

MODULE_PATH=$(go list -m -f '{{.Path}}')
DOWNLOAD_JSON=$(GOWORK=off go mod download -json "${MODULE_PATH}@${TAG}")
MODULE_VERSION=$(printf '%s\n' "${DOWNLOAD_JSON}" | sed -n 's/^[[:space:]]*"Version": "\([^"]*\)",/\1/p' | head -n1)
MODULE_SUM=$(printf '%s\n' "${DOWNLOAD_JSON}" | sed -n 's/^[[:space:]]*"Sum": "\([^"]*\)",/\1/p' | head -n1)
MODULE_GO_MOD_SUM=$(printf '%s\n' "${DOWNLOAD_JSON}" | sed -n 's/^[[:space:]]*"GoModSum": "\([^"]*\)".*/\1/p' | head -n1)
if [ "${MODULE_VERSION}" != "${TAG}" ] || [ -z "${MODULE_SUM}" ] || [ -z "${MODULE_GO_MOD_SUM}" ]; then
  echo "module proxy did not resolve ${MODULE_PATH}@${TAG} with integrity sums" >&2
  exit 1
fi

ARCHIVE_SHA256=$(git archive --format=tar "${TAG}" | sha256sum | awk '{print $1}')
GO_MOD_SHA256=$(git show "${TAG}:go.mod" | sha256sum | awk '{print $1}')

cat >"${MANIFEST_PATH}" <<EOF
{
  "schema_version": 1,
  "module": "${MODULE_PATH}",
  "version": "${TAG}",
  "tag_object": "${LOCAL_TAG_OBJECT}",
  "commit": "${LOCAL_COMMIT}",
  "remote_tag_object": "${REMOTE_TAG_OBJECT}",
  "remote_commit": "${REMOTE_COMMIT}",
  "module_sum": "${MODULE_SUM}",
  "go_mod_sum": "${MODULE_GO_MOD_SUM}",
  "source_archive_sha256": "${ARCHIVE_SHA256}",
  "go_mod_sha256": "${GO_MOD_SHA256}"
}
EOF

echo "release manifest verified and written to ${MANIFEST_PATH}"

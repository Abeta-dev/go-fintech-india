#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

TMP_DIR=$(mktemp -d)
TEST_REPO="${TMP_DIR}/repo"
cleanup() {
  git worktree remove --force "${TEST_REPO}" >/dev/null 2>&1 || true
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

git worktree add --detach "${TEST_REPO}" v0.2.3 >/dev/null
cp scripts/verify_release.sh scripts/check_version.sh "${TEST_REPO}/scripts/"
chmod +x "${TEST_REPO}/scripts/verify_release.sh" "${TEST_REPO}/scripts/check_version.sh"

FAKE_GO="${TMP_DIR}/go"
STATE_FILE="${TMP_DIR}/attempts"
SLEEP_LOG="${TMP_DIR}/sleep.log"
MANIFEST_PATH="${TMP_DIR}/release-manifest.json"

cat >"${FAKE_GO}" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
if [ "$1" = "list" ]; then
  printf '%s\n' 'github.com/umesh0492/go-fintech-india'
  exit 0
fi
if [ "$1" = "clean" ] && [ "$2" = "-modcache" ]; then
  exit 0
fi
if [ "$1" = "mod" ] && [ "$2" = "download" ]; then
  attempt=$(cat "$FAKE_GO_STATE" 2>/dev/null || printf '0')
  attempt=$((attempt + 1))
  printf '%s\n' "$attempt" >"$FAKE_GO_STATE"
  if [ "$attempt" -lt 3 ]; then
    echo "synthetic proxy failure ${attempt}" >&2
    exit 1
  fi
  cat <<JSON
{
  "Path": "github.com/umesh0492/go-fintech-india",
  "Version": "v0.2.3",
  "Sum": "h1:synthetic-module-sum",
  "GoModSum": "h1:synthetic-go-mod-sum"
}
JSON
  exit 0
fi
echo "unexpected fake go invocation: $*" >&2
exit 2
EOF
chmod +x "${FAKE_GO}"

cat >"${TMP_DIR}/sleep" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$1" >>"$FAKE_SLEEP_LOG"
EOF
chmod +x "${TMP_DIR}/sleep"

output=$(cd "${TEST_REPO}" && \
  GO_CMD="${FAKE_GO}" \
  FAKE_GO_STATE="${STATE_FILE}" \
  FAKE_SLEEP_LOG="${SLEEP_LOG}" \
  RELEASE_PROXY_SLEEP_COMMAND="${TMP_DIR}/sleep" \
  RELEASE_PROXY_RETRY_ATTEMPTS=3 \
  RELEASE_PROXY_RETRY_INITIAL_DELAY_SECONDS=1 \
  RELEASE_PROXY_RETRY_MAX_DELAY_SECONDS=4 \
  RELEASE_MANIFEST_PATH="${MANIFEST_PATH}" \
  GIT_TAG=v0.2.3 \
  PATH="/usr/local/go/bin:${PATH}" \
  ./scripts/verify_release.sh v0.2.3 2>&1)
printf '%s\n' "${output}" | grep -F 'attempt 1/3'
printf '%s\n' "${output}" | grep -F 'attempt 2/3'
[ "$(cat "${STATE_FILE}")" = "3" ]
printf '1\n2\n' | cmp -s - "${SLEEP_LOG}"
grep -F '"verification_status": "verified"' "${MANIFEST_PATH}"
grep -F 'h1:synthetic-module-sum' "${MANIFEST_PATH}"

cat >"${TMP_DIR}/always-fail-go" <<'EOF'
#!/usr/bin/env bash
if [ "$1" = "list" ]; then
  printf '%s\n' 'github.com/umesh0492/go-fintech-india'
  exit 0
fi
if [ "$1" = "clean" ] && [ "$2" = "-modcache" ]; then
  exit 0
fi
echo 'synthetic final proxy diagnostic' >&2
exit 1
EOF
chmod +x "${TMP_DIR}/always-fail-go"

set +e
failure_output=$(cd "${TEST_REPO}" && \
  GO_CMD="${TMP_DIR}/always-fail-go" \
  RELEASE_PROXY_RETRY_ATTEMPTS=1 \
  RELEASE_MANIFEST_PATH="${TMP_DIR}/failed-release-manifest.json" \
  GIT_TAG=v0.2.3 \
  PATH="/usr/local/go/bin:${PATH}" \
  ./scripts/verify_release.sh v0.2.3 2>&1)
status=$?
set -e
[ "${status}" -ne 0 ]
printf '%s\n' "${failure_output}" | grep -F 'synthetic final proxy diagnostic'
printf '%s\n' "${failure_output}" | grep -F 'failed after 1 attempts'
grep -F '"verification_status": "failed"' "${TMP_DIR}/failed-release-manifest.json"
grep -F '"failure_stage": "module_proxy"' "${TMP_DIR}/failed-release-manifest.json"

set +e
prerelease_output=$(cd "${TEST_REPO}" && \
  PATH="/usr/local/go/bin:${PATH}" GIT_TAG=v0.2.3-rc.1 ./scripts/verify_release.sh v0.2.3-rc.1 2>&1)
status=$?
set -e
[ "${status}" -ne 0 ]
printf '%s\n' "${prerelease_output}" | grep -F 'Expected Version: v0.2.3-rc.1'
if printf '%s\n' "${prerelease_output}" | grep -F 'usage:' >/dev/null; then
  echo 'valid prerelease tag was rejected by parser' >&2
  exit 1
fi

printf '%s\n' 'verify_release.sh retry, failure-manifest, and prerelease parsing tests passed'

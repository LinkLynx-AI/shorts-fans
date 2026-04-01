#!/usr/bin/env bash
set -euo pipefail

UPSTREAM_REPO="zenyonedad-glitch/short-fans"
UPSTREAM_PATH="_project/short-fans/ssot"
REQUESTED_REF="${1:-main}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
TARGET_DIR="${REPO_ROOT}/docs/ssot"
STAGING_DIR="$(mktemp -d "${TMPDIR:-/tmp}/shorts-fans-ssot.XXXXXX")"

cleanup() {
  rm -rf "${STAGING_DIR}"
}
trap cleanup EXIT

if ! command -v gh >/dev/null 2>&1; then
  echo "gh command is required" >&2
  exit 1
fi

if ! command -v rsync >/dev/null 2>&1; then
  echo "rsync command is required" >&2
  exit 1
fi

RESOLVED_SHA="$(gh api "repos/${UPSTREAM_REPO}/commits/${REQUESTED_REF}" --jq '.sha')"

UPSTREAM_FILES=()
while IFS= read -r upstream_file; do
  UPSTREAM_FILES+=("${upstream_file}")
done < <(
  gh api "repos/${UPSTREAM_REPO}/git/trees/${RESOLVED_SHA}?recursive=1" \
    --jq ".tree[] | select(.type == \"blob\" and (.path | startswith(\"${UPSTREAM_PATH}/\"))) | .path"
)

if [ "${#UPSTREAM_FILES[@]}" -eq 0 ]; then
  echo "No files found under ${UPSTREAM_PATH} at ${RESOLVED_SHA}" >&2
  exit 1
fi

for upstream_file in "${UPSTREAM_FILES[@]}"; do
  relative_path="${upstream_file#${UPSTREAM_PATH}/}"
  destination_path="${STAGING_DIR}/${relative_path}"
  mkdir -p "$(dirname "${destination_path}")"

  gh api \
    -H "Accept: application/vnd.github.raw" \
    "repos/${UPSTREAM_REPO}/contents/${upstream_file}?ref=${RESOLVED_SHA}" \
    > "${destination_path}"
done

mkdir -p "${TARGET_DIR}"
rsync -a --delete \
  --exclude "LOCAL_INDEX.md" \
  --exclude "SOURCE.md" \
  "${STAGING_DIR}/" "${TARGET_DIR}/"

cat > "${TARGET_DIR}/SOURCE.md" <<EOF
# SSOT Source

- source repo: \`https://github.com/${UPSTREAM_REPO}\`
- source path: \`${UPSTREAM_PATH}\`
- source ref: \`${RESOLVED_SHA}\`
- imported at: \`$(date +%F)\`
- sync command: \`./scripts/sync_ssot.sh [ref]\`
- notes: \`docs/ssot\` is a vendored copy of the upstream SSOT subtree.
EOF

echo "Synced docs/ssot from ${UPSTREAM_REPO}@${RESOLVED_SHA}"

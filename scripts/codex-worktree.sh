#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  scripts/codex-worktree.sh <branch-name> [-- <codex-args...>]

Behavior:
  - Creates or reuses a git worktree at /codex/<branch-name>.
  - Launches codex inside the target worktree.

Examples:
  scripts/codex-worktree.sh feat/frontend-shell
  scripts/codex-worktree.sh feat/frontend-shell -- exec
EOF
}

require_command() {
  local command_name="$1"

  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "error: '$command_name' is required but was not found." >&2
    exit 1
  fi
}

sanitize_worktree_name() {
  printf '%s' "$1" \
    | tr '/:@ ' '----' \
    | tr -cd '[:alnum:]._-'
}

find_existing_worktree_for_branch() {
  local branch_name="$1"

  git worktree list --porcelain | awk -v target_branch="refs/heads/${branch_name}" '
    $1 == "worktree" {
      current_worktree = $2
      next
    }

    $1 == "branch" && $2 == target_branch {
      print current_worktree
      exit
    }
  '
}

require_command git
require_command codex

if [[ "${1:-}" == "-h" ]] || [[ "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ "${1:-}" == "" ]]; then
  usage
  exit 1
fi

branch_name="$1"
shift

if [[ "${1:-}" == "--" ]]; then
  shift
fi

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(git -C "${script_dir}/.." rev-parse --show-toplevel)"
base_ref="main"

if ! git show-ref --verify --quiet "refs/heads/${base_ref}"; then
  base_ref="$(git rev-parse --abbrev-ref HEAD)"
fi

worktree_name="$(sanitize_worktree_name "$branch_name")"

if [[ -z "$worktree_name" ]]; then
  echo "error: failed to derive a safe worktree identifier from branch '${branch_name}'." >&2
  exit 1
fi

target_worktree_path="/codex/${branch_name}"
existing_worktree_path="$(find_existing_worktree_for_branch "$branch_name")"

cd "$repo_root"

if [[ -n "$existing_worktree_path" ]]; then
  target_worktree_path="$existing_worktree_path"
elif [[ -e "$target_worktree_path" ]]; then
  echo "error: target path already exists and is not registered as a worktree: ${target_worktree_path}" >&2
  exit 1
elif git show-ref --verify --quiet "refs/heads/${branch_name}"; then
  mkdir -p "$(dirname "$target_worktree_path")"
  git worktree add "$target_worktree_path" "$branch_name"
else
  mkdir -p "$(dirname "$target_worktree_path")"
  git worktree add -b "$branch_name" "$target_worktree_path" "$base_ref"
fi

echo "worktree: ${target_worktree_path}"
echo "branch: ${branch_name}"

cd "$target_worktree_path"
exec codex "$@"

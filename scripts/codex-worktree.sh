#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  scripts/codex-worktree.sh <branch-name> [-- <codex-args...>]

Behavior:
  - Creates a git worktree under $HOME/.codex/worktrees by default.
  - The worktree directory name is derived from the branch name.
  - Launches codex with --ask-for-approval never, --sandbox <mode>, and --cd <worktree-path>.

Environment:
  CODEX_WORKTREE_ROOT  Destination root for created worktrees
                       (default: $HOME/.codex/worktrees)
  CODEX_SANDBOX_MODE   Sandbox mode passed to codex
                       (default: danger-full-access)
  CODEX_INHERIT_GH_TOKEN
                       If set to 1, export GH auth into the Codex session.
                       Resolves from GH_TOKEN, then GITHUB_TOKEN, then `gh auth token`.
                       If resolution fails, prints a warning and continues without token inheritance.
                       (default: 1)
  CODEX_EXTRA_WRITABLE_DIRS
                       Colon-separated extra writable directories.
                       Relative paths are resolved from the created worktree root.
                       Example: .agents:.codex

Examples:
  scripts/codex-worktree.sh feat/frontend-shell
  scripts/codex-worktree.sh feat/frontend-shell -- exec
  scripts/codex-worktree.sh feat/gh-pr
  CODEX_EXTRA_WRITABLE_DIRS=".agents:.codex" scripts/codex-worktree.sh feat/skill-edit
  CODEX_SANDBOX_MODE=workspace-write scripts/codex-worktree.sh feat/frontend-shell
EOF
}

require_command() {
  local command_name="$1"

  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "error: '$command_name' is required but was not found." >&2
    exit 1
  fi
}

warn() {
  echo "warning: $*" >&2
}

slugify() {
  local value="$1"

  value="$(printf '%s' "$value" | tr '[:upper:]' '[:lower:]')"
  value="$(printf '%s' "$value" | sed -E 's/[^a-z0-9._-]+/-/g; s/^-+//; s/-+$//; s/-{2,}/-/g')"

  printf '%s' "$value"
}

resolve_default_base_ref() {
  local repo_root="$1"
  local current_branch

  current_branch="$(git -C "$repo_root" symbolic-ref --quiet --short HEAD 2>/dev/null || true)"
  if [[ -n "$current_branch" ]]; then
    printf '%s\n' "$current_branch"
    return
  fi

  local origin_head

  origin_head="$(git -C "$repo_root" symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null || true)"
  if [[ -n "$origin_head" ]]; then
    printf '%s\n' "$origin_head"
    return
  fi

  if git -C "$repo_root" show-ref --verify --quiet refs/heads/main; then
    printf 'main\n'
    return
  fi

  if git -C "$repo_root" show-ref --verify --quiet refs/heads/master; then
    printf 'master\n'
    return
  fi

  git -C "$repo_root" rev-parse --short HEAD
}

resolve_gh_token() {
  local resolved_token=""

  if [[ -n "${GH_TOKEN:-}" ]]; then
    printf '%s\n' "$GH_TOKEN"
    return 0
  fi

  if [[ -n "${GITHUB_TOKEN:-}" ]]; then
    printf '%s\n' "$GITHUB_TOKEN"
    return 0
  fi

  if ! command -v gh >/dev/null 2>&1; then
    warn "gh is not available; continuing without GH token inheritance."
    return 0
  fi

  resolved_token="$(gh auth token 2>/dev/null || true)"

  if [[ -n "$resolved_token" ]]; then
    printf '%s\n' "$resolved_token"
    return 0
  fi

  warn "could not resolve GitHub auth token; continuing without GH token inheritance."
  return 0
}

resolve_shell_environment_policy_inherit_override() {
  local selective_inherit_override='shell_environment_policy.inherit=["GH_TOKEN","GITHUB_TOKEN"]'
  local inherit_all_override='shell_environment_policy.inherit=all'

  if codex -c "$selective_inherit_override" features list >/dev/null 2>&1; then
    printf '%s\n' "$selective_inherit_override"
    return 0
  fi

  # Newer Codex releases only accept enum values like core/all/none here.
  # Fall back to "all" so GH auth still reaches the Codex session.
  if codex -c "$inherit_all_override" features list >/dev/null 2>&1; then
    warn "current codex does not support selectively allowlisting GH_TOKEN/GITHUB_TOKEN; falling back to shell_environment_policy.inherit=all."
    printf '%s\n' "$inherit_all_override"
    return 0
  fi

  warn "could not determine shell_environment_policy compatibility; continuing without an explicit shell environment inherit override."
  return 0
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
base_ref="$(resolve_default_base_ref "$repo_root")"
worktree_name="$(slugify "${branch_name#codex/}")"

if [[ -z "$worktree_name" ]]; then
  echo "error: failed to derive a safe worktree identifier from branch '${branch_name}'." >&2
  exit 1
fi

if ! git -C "$repo_root" rev-parse --verify --quiet "${base_ref}^{commit}" >/dev/null; then
  echo "error: base ref not found: ${base_ref}" >&2
  exit 1
fi

if git -C "$repo_root" worktree list --porcelain | awk '/^branch / {print $2}' | grep -Fxq "refs/heads/${branch_name}"; then
  echo "error: branch is already checked out in another worktree: ${branch_name}" >&2
  exit 1
fi

worktree_root="${CODEX_WORKTREE_ROOT:-$HOME/.codex/worktrees}"
mkdir -p "$worktree_root"

session_dir="$(mktemp -d "${worktree_root%/}/${worktree_name}-XXXXXX")"
session_dir="$(cd "$session_dir" && pwd -P)"
repo_name="$(basename "$repo_root")"
target_worktree_path="${session_dir}/${repo_name}"

cd "$repo_root"

if git show-ref --verify --quiet "refs/heads/${branch_name}"; then
  git worktree add "$target_worktree_path" "$branch_name"
else
  git worktree add -b "$branch_name" "$target_worktree_path" "$base_ref"
fi

echo "worktree: ${target_worktree_path}"
echo "branch: ${branch_name}"
echo "base: ${base_ref}"

sandbox_mode="${CODEX_SANDBOX_MODE:-danger-full-access}"
inherit_gh_token="${CODEX_INHERIT_GH_TOKEN:-1}"

codex_args=(
  --ask-for-approval never
  --sandbox "$sandbox_mode"
  --cd "$target_worktree_path"
)

if [[ "$inherit_gh_token" == "1" ]]; then
  resolved_gh_token="$(resolve_gh_token)"

  if [[ -n "$resolved_gh_token" ]]; then
    export GH_TOKEN="$resolved_gh_token"
    export GITHUB_TOKEN="${GITHUB_TOKEN:-$resolved_gh_token}"
    shell_environment_policy_override="$(resolve_shell_environment_policy_inherit_override)"

    if [[ -n "$shell_environment_policy_override" ]]; then
      codex_args+=(-c "$shell_environment_policy_override")
    fi
  fi
fi

if [[ -n "${CODEX_EXTRA_WRITABLE_DIRS:-}" ]]; then
  old_ifs="$IFS"
  IFS=':'
  read -r -a extra_writable_dirs <<< "${CODEX_EXTRA_WRITABLE_DIRS}"
  IFS="$old_ifs"

  for extra_dir in "${extra_writable_dirs[@]}"; do
    [[ -z "$extra_dir" ]] && continue

    if [[ "$extra_dir" = /* ]]; then
      codex_args+=(--add-dir "$extra_dir")
    else
      codex_args+=(--add-dir "${target_worktree_path}/${extra_dir}")
    fi
  done
fi

exec codex "${codex_args[@]}" "$@"

SHELL := /bin/bash

.DEFAULT_GOAL := help

.PHONY: help codex codex-worktree

help:
	@printf '%s\n' \
		'Usage:' \
		'  make codex branch=<branch-name> [ARGS="..."]' \
		'  make codex-worktree branch=<branch-name> [ARGS="..."]' \
		'' \
		'Examples:' \
		'  make codex branch=feat/frontend-shell' \
		'  make codex branch=feat/frontend-shell ARGS="exec"'

codex: codex-worktree

codex-worktree:
	@if [[ -z "$(strip $(branch))" ]]; then \
		echo 'error: branch is required. Usage: make codex branch=<branch-name> [ARGS="..."]' >&2; \
		exit 1; \
	fi
	@if [[ -n "$(strip $(ARGS))" ]]; then \
		./scripts/codex-worktree.sh "$(branch)" -- $(ARGS); \
	else \
		./scripts/codex-worktree.sh "$(branch)"; \
	fi

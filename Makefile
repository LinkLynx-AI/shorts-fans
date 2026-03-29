SHELL := /bin/bash

.DEFAULT_GOAL := help

.PHONY: help codex codex-worktree backend-dev-up backend-dev-down backend-run backend-worker backend-migrate-up backend-migrate-down backend-generate backend-test backend-vet backend-fmt

BACKEND_DIR := backend
BACKEND_APP_ENV ?= development
BACKEND_API_ADDR ?= :8080
BACKEND_POSTGRES_DSN ?= postgres://shorts_fans:shorts_fans@localhost:5432/shorts_fans?sslmode=disable
BACKEND_REDIS_ADDR ?= localhost:6379
BACKEND_AWS_REGION ?=
BACKEND_SQS_QUEUE_URL ?=
SQLC_VERSION := v1.27.0

help:
	@printf '%s\n' \
		'Usage:' \
		'  make codex branch=<branch-name> [ARGS="..."]' \
		'  make codex-worktree branch=<branch-name> [ARGS="..."]' \
		'  make backend-dev-up' \
		'  make backend-run' \
		'  make backend-worker' \
		'' \
		'Examples:' \
		'  make codex branch=feat/frontend-shell' \
		'  make codex branch=feat/frontend-shell ARGS="exec"' \
		'  make backend-run'

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

backend-dev-up:
	docker compose up -d --wait postgres redis

backend-dev-down:
	docker compose down

backend-run:
	cd $(BACKEND_DIR) && \
		APP_ENV='$(BACKEND_APP_ENV)' \
		API_ADDR='$(BACKEND_API_ADDR)' \
		POSTGRES_DSN='$(BACKEND_POSTGRES_DSN)' \
		REDIS_ADDR='$(BACKEND_REDIS_ADDR)' \
		AWS_REGION='$(BACKEND_AWS_REGION)' \
		SQS_QUEUE_URL='$(BACKEND_SQS_QUEUE_URL)' \
		go run ./cmd/api

backend-worker:
	cd $(BACKEND_DIR) && \
		APP_ENV='$(BACKEND_APP_ENV)' \
		API_ADDR='$(BACKEND_API_ADDR)' \
		POSTGRES_DSN='$(BACKEND_POSTGRES_DSN)' \
		REDIS_ADDR='$(BACKEND_REDIS_ADDR)' \
		AWS_REGION='$(BACKEND_AWS_REGION)' \
		SQS_QUEUE_URL='$(BACKEND_SQS_QUEUE_URL)' \
		go run ./cmd/worker

backend-migrate-up:
	cd $(BACKEND_DIR) && \
		POSTGRES_DSN='$(BACKEND_POSTGRES_DSN)' \
		go run ./cmd/migrate up

backend-migrate-down:
	cd $(BACKEND_DIR) && \
		POSTGRES_DSN='$(BACKEND_POSTGRES_DSN)' \
		go run ./cmd/migrate down

backend-generate:
	cd $(BACKEND_DIR) && go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) generate

backend-test:
	cd $(BACKEND_DIR) && go test ./...

backend-vet:
	cd $(BACKEND_DIR) && go vet ./...

backend-fmt:
	cd $(BACKEND_DIR) && gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

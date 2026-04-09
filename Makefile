SHELL := /bin/bash

.DEFAULT_GOAL := help

.PHONY: help codex codex-worktree backend-dev-up backend-dev-down backend-run backend-worker backend-media-smoke backend-dev-seed backend-migrate-up backend-migrate-down backend-generate backend-schema backend-test backend-coverage-check backend-vet backend-fmt

BACKEND_DIR := backend
BACKEND_APP_ENV ?= development
BACKEND_API_ADDR ?= :8080
BACKEND_POSTGRES_DSN ?= postgres://shorts_fans:shorts_fans@localhost:5432/shorts_fans?sslmode=disable
BACKEND_REDIS_ADDR ?= localhost:6379
BACKEND_AWS_REGION ?=
BACKEND_SQS_QUEUE_URL ?=
BACKEND_MEDIA_JOBS_QUEUE_URL ?= $(BACKEND_SQS_QUEUE_URL)
BACKEND_MEDIA_RAW_BUCKET_NAME ?=
BACKEND_MEDIA_SHORT_PUBLIC_BUCKET_NAME ?=
BACKEND_MEDIA_SHORT_PUBLIC_BASE_URL ?=
BACKEND_MEDIA_MAIN_PRIVATE_BUCKET_NAME ?=
BACKEND_MEDIACONVERT_SERVICE_ROLE_ARN ?=
BACKEND_COVERAGE_MIN ?=
BACKEND_COVERAGE_PROFILE ?=
SQLC_VERSION := v1.27.0

help:
	@printf '%s\n' \
		'Usage:' \
		'  make codex branch=<branch-name> [ARGS="..."]   # creates/switches codex/<branch-name>' \
		'  make codex-worktree branch=<branch-name> [ARGS="..."]   # creates/switches codex/<branch-name>' \
		'  make backend-dev-up' \
		'  make backend-dev-seed' \
		'  make backend-run' \
		'  make backend-worker' \
		'  make backend-media-smoke' \
		'  make backend-schema' \
		'  make backend-coverage-check [BACKEND_COVERAGE_MIN=<min-percent>]' \
		'' \
		'Examples:' \
		'  make codex branch=frontend-shell          # branch: codex/frontend-shell' \
		'  make codex branch=frontend-shell ARGS="exec"' \
		'  make backend-run' \
		'  make backend-dev-seed' \
		'  make backend-media-smoke' \
		'  make backend-coverage-check BACKEND_COVERAGE_MIN=30'

codex: codex-worktree

codex-worktree:
	@if [[ -z "$(strip $(branch))" ]]; then \
		echo 'error: branch is required. Usage: make codex branch=<branch-name> [ARGS="..."]' >&2; \
		exit 1; \
	fi
	@normalized_branch='$(strip $(branch))'; \
	if [[ "$$normalized_branch" != codex/* ]]; then \
		normalized_branch="codex/$$normalized_branch"; \
	fi; \
	if [[ -n "$(strip $(ARGS))" ]]; then \
		./scripts/codex-worktree.sh "$$normalized_branch" -- $(ARGS); \
	else \
		./scripts/codex-worktree.sh "$$normalized_branch"; \
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
		MEDIA_JOBS_QUEUE_URL='$(BACKEND_MEDIA_JOBS_QUEUE_URL)' \
		MEDIA_RAW_BUCKET_NAME='$(BACKEND_MEDIA_RAW_BUCKET_NAME)' \
		MEDIA_SHORT_PUBLIC_BUCKET_NAME='$(BACKEND_MEDIA_SHORT_PUBLIC_BUCKET_NAME)' \
		MEDIA_SHORT_PUBLIC_BASE_URL='$(BACKEND_MEDIA_SHORT_PUBLIC_BASE_URL)' \
		MEDIA_MAIN_PRIVATE_BUCKET_NAME='$(BACKEND_MEDIA_MAIN_PRIVATE_BUCKET_NAME)' \
		MEDIACONVERT_SERVICE_ROLE_ARN='$(BACKEND_MEDIACONVERT_SERVICE_ROLE_ARN)' \
		go run ./cmd/api

backend-worker:
	cd $(BACKEND_DIR) && \
		APP_ENV='$(BACKEND_APP_ENV)' \
		API_ADDR='$(BACKEND_API_ADDR)' \
		POSTGRES_DSN='$(BACKEND_POSTGRES_DSN)' \
		REDIS_ADDR='$(BACKEND_REDIS_ADDR)' \
		AWS_REGION='$(BACKEND_AWS_REGION)' \
		MEDIA_JOBS_QUEUE_URL='$(BACKEND_MEDIA_JOBS_QUEUE_URL)' \
		MEDIA_RAW_BUCKET_NAME='$(BACKEND_MEDIA_RAW_BUCKET_NAME)' \
		MEDIA_SHORT_PUBLIC_BUCKET_NAME='$(BACKEND_MEDIA_SHORT_PUBLIC_BUCKET_NAME)' \
		MEDIA_SHORT_PUBLIC_BASE_URL='$(BACKEND_MEDIA_SHORT_PUBLIC_BASE_URL)' \
		MEDIA_MAIN_PRIVATE_BUCKET_NAME='$(BACKEND_MEDIA_MAIN_PRIVATE_BUCKET_NAME)' \
		MEDIACONVERT_SERVICE_ROLE_ARN='$(BACKEND_MEDIACONVERT_SERVICE_ROLE_ARN)' \
		go run ./cmd/worker

backend-media-smoke:
	cd $(BACKEND_DIR) && \
		AWS_REGION='$(BACKEND_AWS_REGION)' \
		MEDIA_JOBS_QUEUE_URL='$(BACKEND_MEDIA_JOBS_QUEUE_URL)' \
		MEDIA_RAW_BUCKET_NAME='$(BACKEND_MEDIA_RAW_BUCKET_NAME)' \
		MEDIA_SHORT_PUBLIC_BUCKET_NAME='$(BACKEND_MEDIA_SHORT_PUBLIC_BUCKET_NAME)' \
		MEDIA_SHORT_PUBLIC_BASE_URL='$(BACKEND_MEDIA_SHORT_PUBLIC_BASE_URL)' \
		MEDIA_MAIN_PRIVATE_BUCKET_NAME='$(BACKEND_MEDIA_MAIN_PRIVATE_BUCKET_NAME)' \
		MEDIACONVERT_SERVICE_ROLE_ARN='$(BACKEND_MEDIACONVERT_SERVICE_ROLE_ARN)' \
		go run ./cmd/media-smoke

backend-dev-seed:
	cd $(BACKEND_DIR) && \
		APP_ENV='$(BACKEND_APP_ENV)' \
		POSTGRES_DSN='$(BACKEND_POSTGRES_DSN)' \
		go run ./cmd/devseed

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

backend-schema:
	APP_ENV='$(BACKEND_APP_ENV)' \
	POSTGRES_DSN='$(BACKEND_POSTGRES_DSN)' \
	./scripts/generate-db-schema.sh

backend-test:
	cd $(BACKEND_DIR) && go test ./...

backend-coverage-check:
	BACKEND_DIR='$(BACKEND_DIR)' \
	BACKEND_COVERAGE_MIN='$(BACKEND_COVERAGE_MIN)' \
	BACKEND_COVERAGE_PROFILE='$(BACKEND_COVERAGE_PROFILE)' \
	./scripts/check-backend-coverage.sh

backend-vet:
	cd $(BACKEND_DIR) && go vet ./...

backend-fmt:
	cd $(BACKEND_DIR) && gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

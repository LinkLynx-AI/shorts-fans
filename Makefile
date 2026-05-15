SHELL := /bin/bash

-include .env
-include .env.local

.DEFAULT_GOAL := help

.PHONY: help codex codex-worktree backend-dev-up backend-dev-down backend-run backend-worker backend-media-smoke backend-dev-seed backend-migrate-up backend-migrate-down backend-generate backend-schema backend-test backend-coverage-check backend-vet backend-fmt aws-dev-ecr-bootstrap aws-dev-build-push aws-dev-apply aws-dev-force-deploy aws-dev-migrate aws-dev-seed

BACKEND_DIR := backend
BACKEND_APP_ENV ?= development
BACKEND_API_ADDR ?= :8080
BACKEND_POSTGRES_DSN ?= postgres://shorts_fans:shorts_fans@localhost:5432/shorts_fans?sslmode=disable
BACKEND_REDIS_ADDR ?= localhost:6379
AWS_PROFILE ?=
BACKEND_AWS_REGION ?=
BACKEND_COGNITO_USER_POOL_ID ?=
BACKEND_COGNITO_USER_POOL_CLIENT_ID ?=
BACKEND_SQS_QUEUE_URL ?=
BACKEND_MEDIA_JOBS_QUEUE_URL ?= $(BACKEND_SQS_QUEUE_URL)
BACKEND_MEDIA_RAW_BUCKET_NAME ?=
BACKEND_MEDIA_SHORT_PUBLIC_BUCKET_NAME ?=
BACKEND_MEDIA_SHORT_PUBLIC_BASE_URL ?=
BACKEND_MEDIA_MAIN_PRIVATE_BUCKET_NAME ?=
BACKEND_MEDIACONVERT_SERVICE_ROLE_ARN ?=
BACKEND_CREATOR_AVATAR_UPLOAD_BUCKET_NAME ?=
BACKEND_CREATOR_AVATAR_DELIVERY_BUCKET_NAME ?=
BACKEND_CREATOR_AVATAR_BASE_URL ?=
BACKEND_CREATOR_REVIEW_EVIDENCE_BUCKET_NAME ?=
BACKEND_COVERAGE_MIN ?=
BACKEND_COVERAGE_PROFILE ?=
SQLC_VERSION := v1.27.0
AWS_DEV_TF_DIR := infra/terraform/dev
AWS_DEV_TFVARS ?= terraform.tfvars
AWS_DEV_IMAGE_TAG ?= dev

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
		'  make aws-dev-build-push [AWS_DEV_IMAGE_TAG=dev]' \
		'  make aws-dev-apply' \
		'  make aws-dev-migrate' \
		'  make aws-dev-seed' \
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
		AWS_PROFILE='$(AWS_PROFILE)' \
		AWS_REGION='$(BACKEND_AWS_REGION)' \
		COGNITO_USER_POOL_ID='$(BACKEND_COGNITO_USER_POOL_ID)' \
		COGNITO_USER_POOL_CLIENT_ID='$(BACKEND_COGNITO_USER_POOL_CLIENT_ID)' \
		MEDIA_JOBS_QUEUE_URL='$(BACKEND_MEDIA_JOBS_QUEUE_URL)' \
		MEDIA_RAW_BUCKET_NAME='$(BACKEND_MEDIA_RAW_BUCKET_NAME)' \
		MEDIA_SHORT_PUBLIC_BUCKET_NAME='$(BACKEND_MEDIA_SHORT_PUBLIC_BUCKET_NAME)' \
		MEDIA_SHORT_PUBLIC_BASE_URL='$(BACKEND_MEDIA_SHORT_PUBLIC_BASE_URL)' \
		MEDIA_MAIN_PRIVATE_BUCKET_NAME='$(BACKEND_MEDIA_MAIN_PRIVATE_BUCKET_NAME)' \
		MEDIACONVERT_SERVICE_ROLE_ARN='$(BACKEND_MEDIACONVERT_SERVICE_ROLE_ARN)' \
		CREATOR_AVATAR_UPLOAD_BUCKET_NAME='$(BACKEND_CREATOR_AVATAR_UPLOAD_BUCKET_NAME)' \
		CREATOR_AVATAR_DELIVERY_BUCKET_NAME='$(BACKEND_CREATOR_AVATAR_DELIVERY_BUCKET_NAME)' \
		CREATOR_AVATAR_BASE_URL='$(BACKEND_CREATOR_AVATAR_BASE_URL)' \
		CREATOR_REVIEW_EVIDENCE_BUCKET_NAME='$(BACKEND_CREATOR_REVIEW_EVIDENCE_BUCKET_NAME)' \
		go run ./cmd/api

backend-worker:
	cd $(BACKEND_DIR) && \
		APP_ENV='$(BACKEND_APP_ENV)' \
		API_ADDR='$(BACKEND_API_ADDR)' \
		POSTGRES_DSN='$(BACKEND_POSTGRES_DSN)' \
		REDIS_ADDR='$(BACKEND_REDIS_ADDR)' \
		AWS_PROFILE='$(AWS_PROFILE)' \
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
		AWS_PROFILE='$(AWS_PROFILE)' \
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

aws-dev-ecr-bootstrap:
	terraform -chdir=$(AWS_DEV_TF_DIR) apply \
		-var-file=$(AWS_DEV_TFVARS) \
		-target=aws_ecr_repository.backend \
		-target=aws_ecr_repository.frontend

aws-dev-build-push: aws-dev-ecr-bootstrap
	@region="$$(terraform -chdir=$(AWS_DEV_TF_DIR) output -raw aws_region)"; \
	account_id="$$(aws sts get-caller-identity --query Account --output text)"; \
	registry="$$account_id.dkr.ecr.$$region.amazonaws.com"; \
	backend_repo="$$(terraform -chdir=$(AWS_DEV_TF_DIR) output -raw backend_ecr_repository_url)"; \
	frontend_repo="$$(terraform -chdir=$(AWS_DEV_TF_DIR) output -raw frontend_ecr_repository_url)"; \
	aws ecr get-login-password --region "$$region" | docker login --username AWS --password-stdin "$$registry"; \
	docker build --platform linux/amd64 -t "$$backend_repo:$(AWS_DEV_IMAGE_TAG)" backend; \
	docker push "$$backend_repo:$(AWS_DEV_IMAGE_TAG)"; \
	docker build --platform linux/amd64 -t "$$frontend_repo:$(AWS_DEV_IMAGE_TAG)" frontend; \
	docker push "$$frontend_repo:$(AWS_DEV_IMAGE_TAG)"

aws-dev-apply:
	terraform -chdir=$(AWS_DEV_TF_DIR) apply -var-file=$(AWS_DEV_TFVARS)

aws-dev-force-deploy:
	@cluster="$$(terraform -chdir=$(AWS_DEV_TF_DIR) output -raw ecs_cluster_name)"; \
	backend_service="$$(terraform -chdir=$(AWS_DEV_TF_DIR) output -raw backend_service_name)"; \
	frontend_service="$$(terraform -chdir=$(AWS_DEV_TF_DIR) output -raw frontend_service_name)"; \
	worker_service="$$(terraform -chdir=$(AWS_DEV_TF_DIR) output -raw worker_service_name)"; \
	aws ecs update-service --cluster "$$cluster" --service "$$backend_service" --force-new-deployment >/dev/null; \
	aws ecs update-service --cluster "$$cluster" --service "$$frontend_service" --force-new-deployment >/dev/null; \
	aws ecs update-service --cluster "$$cluster" --service "$$worker_service" --force-new-deployment >/dev/null

aws-dev-migrate:
	@cluster="$$(terraform -chdir=$(AWS_DEV_TF_DIR) output -raw ecs_cluster_name)"; \
	task_definition="$$(terraform -chdir=$(AWS_DEV_TF_DIR) output -raw backend_maintenance_task_definition_arn)"; \
	security_group="$$(terraform -chdir=$(AWS_DEV_TF_DIR) output -raw ecs_task_security_group_id)"; \
	subnets="$$(terraform -chdir=$(AWS_DEV_TF_DIR) output -json ecs_public_subnet_ids)"; \
	network_configuration="$$(printf '{"awsvpcConfiguration":{"subnets":%s,"securityGroups":["%s"],"assignPublicIp":"ENABLED"}}' "$$subnets" "$$security_group")"; \
	overrides='{"containerOverrides":[{"name":"backend-maintenance","command":["/app/migrate","up"]}]}'; \
	task_arn="$$(aws ecs run-task --cluster "$$cluster" --task-definition "$$task_definition" --launch-type FARGATE --network-configuration "$$network_configuration" --overrides "$$overrides" --query 'tasks[0].taskArn' --output text)"; \
	aws ecs wait tasks-stopped --cluster "$$cluster" --tasks "$$task_arn"; \
	exit_code="$$(aws ecs describe-tasks --cluster "$$cluster" --tasks "$$task_arn" --query 'tasks[0].containers[0].exitCode' --output text)"; \
	test "$$exit_code" = "0"

aws-dev-seed:
	@cluster="$$(terraform -chdir=$(AWS_DEV_TF_DIR) output -raw ecs_cluster_name)"; \
	task_definition="$$(terraform -chdir=$(AWS_DEV_TF_DIR) output -raw backend_maintenance_task_definition_arn)"; \
	security_group="$$(terraform -chdir=$(AWS_DEV_TF_DIR) output -raw ecs_task_security_group_id)"; \
	subnets="$$(terraform -chdir=$(AWS_DEV_TF_DIR) output -json ecs_public_subnet_ids)"; \
	network_configuration="$$(printf '{"awsvpcConfiguration":{"subnets":%s,"securityGroups":["%s"],"assignPublicIp":"ENABLED"}}' "$$subnets" "$$security_group")"; \
	overrides='{"containerOverrides":[{"name":"backend-maintenance","command":["/app/devseed"]}]}'; \
	task_arn="$$(aws ecs run-task --cluster "$$cluster" --task-definition "$$task_definition" --launch-type FARGATE --network-configuration "$$network_configuration" --overrides "$$overrides" --query 'tasks[0].taskArn' --output text)"; \
	aws ecs wait tasks-stopped --cluster "$$cluster" --tasks "$$task_arn"; \
	exit_code="$$(aws ecs describe-tasks --cluster "$$cluster" --tasks "$$task_arn" --query 'tasks[0].containers[0].exitCode' --output text)"; \
	test "$$exit_code" = "0"

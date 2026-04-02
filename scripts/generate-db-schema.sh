#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"

cd "${repo_root}/backend"

APP_ENV="${APP_ENV:-development}" \
POSTGRES_DSN="${POSTGRES_DSN:-postgres://shorts_fans:shorts_fans@localhost:5432/shorts_fans?sslmode=disable}" \
go run ./cmd/schema "$@"

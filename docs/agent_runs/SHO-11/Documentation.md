# Documentation.md (Status / audit log)

## Current status
- Now: implementation updated to add fan-side follow/pin/progress tables
- Next: run live migration validation in an environment with Docker or local Postgres access

## Decisions
- `submission package` table is deferred from SHO-11.
- `media_assets` is included, but ingest/transcode workflow detail is deferred to SHO-23.
- `short -> canonical main` is represented by `shorts.canonical_main_id`, not a join table.
- review state uses `text + CHECK`; `main` / `short` also keep a nullable `review_reason_code`.
- creator-side draft/profile/media/content tables reference `creator_capabilities`, so fan-only users cannot own creator objects.
- `creator_profiles` stores both draft and public profile data; `published_at IS NULL` means draft, non-NULL means public.
- fan-side state uses dedicated tables instead of expanding `users` or `main_unlocks`: `creator_follows`, `pinned_shorts`, `main_playback_progress`.
- `main_playback_progress` is limited to purchased mains via a composite FK to `main_unlocks`.
- `consumer_settings` remains deferred because current docs fix the surface but not concrete fields.
- `users` table does not store auth-provider-specific references because the current docs define login identity but do not fix an external auth provider.
- initial `reviewer_simple` found one blocking issue: creator consistency between `media_assets`, `mains`, and `shorts` was not DB-enforced. This was fixed with composite uniqueness/FKs.

## How to run / demo
- `cd backend && GOCACHE=/tmp/go-build-sho11 go test ./...`
- `cd backend && GOCACHE=/tmp/go-build-sho11 go vet ./...`
- `git diff --check`
- `make backend-dev-up`
- `cd backend && POSTGRES_DSN='postgres://shorts_fans:shorts_fans@localhost:5432/shorts_fans?sslmode=disable' go run ./cmd/migrate up`

## Known issues / follow-ups
- `sqlc generate` may depend on whether `github.com/sqlc-dev/sqlc` is available in the local module cache.
- media pipeline state machine and submission package persistence remain follow-up work.
- `go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0 generate` failed because the sandbox cannot resolve `proxy.golang.org`.
- migration の実 apply / rollback は、Docker socket と localhost:5432 の両方が sandbox で遮断されているため未検証。
- frontend 配下の差分は 0 件のため、UI checks は skip と判断した。

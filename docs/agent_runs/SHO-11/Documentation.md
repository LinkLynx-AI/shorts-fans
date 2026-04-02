# Documentation.md (Status / audit log)

## Current status
- Now: implementation, local validation, and specialist reviewer fallback gate are complete
- Next: re-check migration apply/rollback in a DB-capable environment and refresh sqlc generation if needed

## Decisions
- `submission package` table is deferred from SHO-11.
- `media_assets` is included, but ingest/transcode workflow detail is deferred to SHO-23.
- `short -> canonical main` is represented by `shorts.canonical_main_id`, not a join table.
- review state uses `text + CHECK`; `main` / `short` also keep a nullable `review_reason_code`.
- creator-owned draft/profile/media/content tables hang off `creator_capabilities`, while follow targets hang off public `creator_profiles`.
- `creator_profile_drafts` and public `creator_profiles` are separate tables again so approval gate is expressible at the schema boundary.
- fan-side state uses dedicated tables instead of expanding `users` or `main_unlocks`: `creator_follows`, `pinned_shorts`, `main_playback_progress`.
- `main_playback_progress` is limited to purchased mains via a composite FK to `main_unlocks`.
- `consumer_settings` remains deferred because current docs fix the surface but not concrete fields.
- `users` table does not store auth-provider-specific references because the current docs define login identity but do not fix an external auth provider.
- initial `reviewer_simple` found one blocking issue: creator consistency between `media_assets`, `mains`, and `shorts` was not DB-enforced. This was fixed with composite uniqueness/FKs.
- DB is now the source of truth for high-risk invariants only:
  - public creator profile requires `creator_capabilities.state = 'approved'`
  - creator-owned media/main/short rows require approved creator capability at insert time
  - `creator_follows` can only target creators whose capability is still `approved`
  - `main_unlocks` cannot encode creator self-purchase and can only target `approved_for_unlock` mains that are not access-limited by post-report state
- DB read-model views document public/access predicates for later query work:
  - `app.public_creator_profiles`
  - `app.unlockable_mains`
  - `app.public_shorts`
- meta `reviewer` could not complete in this sandbox because its internal review command hit `system-configuration` / OTEL initialization failures.
- specialist reviewer fallback completed and returned `No blocking findings remain.` for the current diff.

## How to run / demo
- `cd backend && GOCACHE=/tmp/go-build-sho11 go test ./...`
- `cd backend && GOCACHE=/tmp/go-build-sho11 go vet ./...`
- `git diff --check`
- `codex review --uncommitted`
- `make backend-dev-up`
- `cd backend && POSTGRES_DSN='postgres://shorts_fans:shorts_fans@localhost:5432/shorts_fans?sslmode=disable' go run ./cmd/migrate up`

## Known issues / follow-ups
- `sqlc generate` may depend on whether `github.com/sqlc-dev/sqlc` is available in the local module cache.
- media pipeline state machine and submission package persistence remain follow-up work.
- `go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0 generate` failed because the sandbox cannot resolve `proxy.golang.org`.
- migration の実 apply / rollback は、Docker socket と localhost:5432 の両方が sandbox で遮断されているため未検証。
- `consumer settings` は SSOT 上 surface としては in-scope だが、列契約が docs で未確定のためこの issue では追加していない。
- `codex review --uncommitted` / `codex exec review --uncommitted` は、この sandbox では `system-configuration` / OTEL 初期化 panic で review 結果を返せなかった。
- frontend 配下の差分は 0 件のため、UI checks は skip と判断した。

## Validation log
- 2026-04-02: `git diff --check` passed after migration/doc updates.
- 2026-04-02: `cd backend && GOCACHE=/tmp/go-build-sho11 go test ./...` passed.
- 2026-04-02: `cd backend && GOCACHE=/tmp/go-build-sho11 go vet ./...` passed.
- 2026-04-02: reviewer gate requested; local follow-up tightened `app.public_shorts` to require `published_at IS NOT NULL` before exposing a short as public.
- 2026-04-02: local follow-up removed a duplicate `creator_follows` trigger/function definition and re-ran `git diff --check`, `go test`, and `go vet`.
- 2026-04-02: specialist reviewer fallback returned `No blocking findings remain.` after the final schema/doc updates.

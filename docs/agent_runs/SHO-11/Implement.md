# Implement.md (Runbook)

- Follow `Plan.md` in order.
- Keep the diff scoped to SHO-11 migration work.
- Prefer `text + CHECK` over PostgreSQL enum so later migrations stay easy to evolve with `sqlc`.
- Treat `main unlock` as purchase rows only. Do not encode creator ownership in the unlock table.
- Keep `submission package` schema out of this issue. Use object-level review state on `creator`, `main`, and `short`.
- Continuously update `Documentation.md` with decisions, validation output, and follow-ups.

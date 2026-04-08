<!-- BEGIN:nextjs-agent-rules -->
# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` before writing any code. Heed deprecation notices.
<!-- END:nextjs-agent-rules -->

## Frontend Test Policy

- Default frontend validation must use `pnpm test` or `pnpm test:unit`; both are unit/component-only.
- Do not run Playwright or `pnpm test:e2e` unless the user explicitly asks for it.
- Keep coverage checks on `pnpm test:coverage:check`; Playwright E2E is outside the default coverage gate.

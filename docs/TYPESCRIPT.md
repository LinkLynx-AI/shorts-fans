# TypeScript Frontend Rules

## 0. Purpose
- Provide consistent rules that scale with code quality, readability, maintainability, team size, and feature growth.
- Shift review focus toward design and requirements by reducing style-driven diffs.
- Automate what can be automated (`lint`, `format`, `typecheck`), and document only human-judgment rules.

## 1. Priority of Rules (When Conflicts Exist)
1. Type checker (TypeScript)
2. Linter (ESLint)
3. Formatter (Prettier)
4. This document (AGENTS.md)

## 2. Non-Negotiables
- Assume `strict: true`.
- `any` is prohibited by default. Parse and guard `unknown` at boundaries.
- `eslint-disable` and `@ts-expect-error` are allowed only with a reason comment, ticket ID, expiry date, and line-level scope.
- `@ts-ignore` is prohibited.
- Imports that break FSD boundaries or Public API contracts are prohibited (no deep imports).
- `useEffect` must not be used except for synchronization with external systems.

## 3. FSD (Feature-Sliced Design) Rules

### 3.1 Layers and Dependency Direction
- Dependencies must flow only from upper layers to lower layers.
- Default layer stack for this project: `app -> widgets -> features -> entities -> shared`.
- `pages` is optional. If used: `app -> pages -> widgets -> features -> entities -> shared`.
- `shared` must not contain domain-specific knowledge.

### 3.2 Slice/Segment Responsibilities
- `ui/`: presentation (React components, styles)
- `model/`: state and domain logic (store, hooks, reducers, use cases)
- `api/`: communication (fetchers, query functions, DTOs, clients)
- `lib/`: slice-local helpers (pure utilities)
- `config/`: constants and configuration

### 3.3 Public API (`index.ts`)
- Every slice must expose only approved exports through `index.ts`.
- Imports must go through Public API only (no deep imports).
- Treat Public API as a contract that isolates internal implementation.

### 3.4 Type Placement
- Slice-local types: inside each slice (e.g., `model/types.ts`).
- Domain types shared across slices: place in `entities/<entity>/model`.
- Generic shared types (`Result`, `ID`, etc.): place in `shared`.

## 4. TypeScript Rules

### 4.1 Prohibitions/Restrictions
- `any` is prohibited by default (`unknown` + parser/guard for temporary handling).
- Non-null assertion (`!`) is prohibited by default.
- Forceful `as` assertions must be contained at boundary layers.

### 4.2 Type Definition Style
- Prefer `type`.
- Use `interface` only when extension-oriented contracts are required.
- `enum` is prohibited by default. Prefer `as const` + unions.

### 4.3 Extended Strict Settings
- `noUncheckedIndexedAccess: true`
- `exactOptionalPropertyTypes: true`
- `noImplicitOverride: true`

## 5. React Rules (Design)
- Separate UI from state/side-effects (`ui/` vs `model/`).
- Limit `useEffect` to external synchronization (network, browser API, subscriptions, timers, non-React libraries).
- Do not synchronize derivable state using `useEffect` + `setState`.

## 5.1 Implementation and Testing Policy
- Split functions by responsibility to keep them testable.
- Keep one primary responsibility per file; split before a file starts mixing multiple concerns (UI/state/API/helpers).
- Implement tests whenever practical, prioritizing regression prevention.
- Separate pure functions from external interactions (I/O, HTTP, browser APIs, state mutation).

## 6. Comment Rules
- Write why, not what.
- TODO/FIXME must include a ticket ID.
- `eslint-disable` and `@ts-expect-error` must include reason, ticket ID, and expiry.

## 7. JSDoc Rules
- Required for:
  - Public APIs (functions/hooks/components exported via `index.ts`)
  - External I/O boundaries in `api/`
  - Logic with pitfalls (caching, concurrency, expensive computation, race conditions)
- Function documentation must use JSDoc syntax.
- The one-line function summary in JSDoc must be written in Japanese.

```ts
/**
 * ユーザー情報を取得する。
 *
 * Contract:
 * - Input preconditions
 * - Return value semantics
 *
 * Errors:
 * - Failure behavior (throw / null / Result)
 *
 * Side effects:
 * - network / storage / analytics / DOM
 */
```

## 8. Import and Dependency Rules
- Use alias imports (`@/`) and avoid deep relative paths.
- `features/entities/widgets/pages` imports must use Public API paths only.
- Circular dependencies are prohibited by default.
- `default export` is prohibited by default.
- Exception: Next.js required files (`app/**/page.tsx`, `app/**/layout.tsx`, `app/**/loading.tsx`, `app/**/error.tsx`, `app/**/not-found.tsx`, `app/**/template.tsx`, `app/**/default.tsx`) may use `default export`.

## 9. Boundary Rules (Zod)
- Parse all external inputs (API responses, URL query params, storage, env) in `api/` or boundary layers using `Zod`.
- Pass only parsed values to inner layers.

## 10. Minimum Testing Rules
- `model/` (pure logic): prioritize unit tests.
- `ui/` (rendering): keep component tests minimal but meaningful.
- E2E: prioritize critical flows (auth, message send, permissions).
- Before review, run `make coverage-check` (or `pnpm run test:coverage:check`) and ensure coverage thresholds are satisfied.
- Coverage excludes static/non-executable artifacts (test files, `d.ts`, `src/shared/styles/**`) and UI shell entries (`src/app/**`).
- Coverage gate is separated by layer:
  - Global: `lines/functions/statements >= 80%`, `branches >= 70%`
  - `src/shared/**`, `src/{entities,features,widgets}/**`: `lines/functions/statements >= 80%`, `branches >= 70%`

## 11. Exception Process
- Any rule exception must include reason, ticket ID, and expiry in comments.
- If an exception becomes permanent, update this rule document or lint configuration.

## 12. Automated Check Mapping
- No `any`: ESLint (`@typescript-eslint/no-explicit-any`)
- No non-null assertion: ESLint (`@typescript-eslint/no-non-null-assertion`)
- Prefer `type`: ESLint (`@typescript-eslint/consistent-type-definitions`)
- No deep imports: ESLint (`no-restricted-imports`, `import/no-internal-modules`)
- FSD layer/Public API boundaries: custom checker (`pnpm run fsd:check` / `make ts-fsd-check`)
- No circular dependencies: ESLint (`import/no-cycle`)
- No default export: ESLint (`import/no-default-export`)
- React Hooks compliance: ESLint (`react-hooks`)
- Type safety: TypeScript (`tsc --noEmit`)
- Formatting: Prettier
- Coverage thresholds (layered): Vitest (`pnpm run test:coverage:check` / `make coverage-check`)

## 13. Review-Reject Anti-Patterns
- Deep imports (bypassing Public API)
- Domain logic placed in `shared`
- Mixed fetch/persistence/analytics logic inside `ui`
- Derived state synced via `useEffect`
- Blanket memoization without measurable justification

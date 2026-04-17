# frontend

`frontend/` は `Next.js App Router` を前提にしたフロントエンド実装基盤です。

## Development

依存を入れたあと、開発サーバーを起動します。

```bash
pnpm dev
```

ブラウザで `http://localhost:3000` を開くと確認できます。

admin review UI を確認するときは、別ターミナルで次を使います。

```bash
pnpm dev:admin
```

ブラウザで `http://localhost:3001/admin` を開くと、local admin の creator review UI を確認できます。
この script は admin route を有効化する専用 env を付けたうえで `127.0.0.1:3001` に loopback bind します。

## Scripts

- `pnpm dev`: 開発サーバーを起動
- `pnpm dev:admin`: admin review UI 用に専用 env を有効化し、`127.0.0.1:3001` へ loopback bind して開発サーバーを起動
- `pnpm lint`: ESLint を実行
- `pnpm lint:eslint`: ESLint だけを実行
- `pnpm fsd:check`: FSD checker を実行
- `pnpm typecheck`: TypeScript の型検査を実行
- `pnpm test`: 既定の frontend test を実行 (`unit / component` のみ)
- `pnpm test:unit`: Vitest の unit / component test を実行
- `pnpm test:coverage:check`: Vitest の coverage 計測を実行
- `pnpm test:all`: unit / component test と Playwright E2E をまとめて実行
- `pnpm test:e2e`: Playwright の明示的な E2E / smoke test を実行
- `pnpm test:e2e:install`: Playwright の Chromium を導入
- `pnpm build`: production build を実行
- `pnpm start`: build 済みアプリを起動

## Structure

- `src/app`: App Router の route / layout / file conventions
- `src/shared`: UI primitive, env, API client, styles
- `src/{entities,features,widgets}`: 具体的な UI/機能が固まった段階で追加する FSD layer
- `tests/e2e`: Playwright smoke test

## Notes

- routing は `App Router` を前提にします。
- import alias は `@/*` を使用します。
- 追加の設計ルールは repo 直下の `AGENTS.md` と `docs/TYPESCRIPT.md` に従います。
- UI は `Tailwind CSS v4 + shadcn/ui 互換 primitive + Radix Primitives` を前提にします。
- `.env.example` に frontend 起動時の最小 env 契約を定義しています。
- Playwright E2E は既定の review/CI 経路には含めず、必要時だけ明示的に実行します。

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.

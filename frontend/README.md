# frontend

`frontend/` は `Next.js App Router` を前提にしたフロントエンド実装基盤です。

## Development

依存を入れたあと、開発サーバーを起動します。

```bash
pnpm dev
```

ブラウザで `http://localhost:3000` を開くと確認できます。

## Scripts

- `pnpm dev`: 開発サーバーを起動
- `pnpm lint`: ESLint を実行
- `pnpm typecheck`: TypeScript の型検査を実行
- `pnpm build`: production build を実行
- `pnpm start`: build 済みアプリを起動

## Structure

- `src/app`: App Router の route / layout
- `src/app/layout.tsx`: root layout
- `src/app/page.tsx`: `/` の初期ページ
- `src/app/globals.css`: アプリ全体の基礎スタイル

## Notes

- routing は `App Router` を前提にします。
- import alias は `@/*` を使用します。
- 追加の設計ルールは repo 直下の `AGENTS.md` と `docs/TYPESCRIPT.md` に従います。

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.

import Link from "next/link";

import { Button } from "@/shared/ui";

export default function NotFound() {
  return (
    <main className="flex min-h-screen items-center justify-center px-6 py-12">
      <section className="w-full max-w-xl rounded-[2rem] border border-white/80 bg-white/80 p-8 shadow-[0_24px_80px_rgba(87,38,8,0.14)] backdrop-blur">
        <p className="text-sm font-semibold uppercase tracking-[0.24em] text-accent">404 / route miss</p>
        <h1 className="mt-4 text-3xl font-semibold tracking-tight text-foreground">
          指定された route はまだ用意されていません。
        </h1>
        <p className="mt-3 text-sm leading-7 text-muted">
          固定のページ構造を置かない前提なので、未定義 URL は root の `not-found.tsx` でまとめて扱います。
        </p>
        <div className="mt-8 flex flex-wrap gap-3">
          <Button asChild>
            <Link href="/">トップへ戻る</Link>
          </Button>
        </div>
      </section>
    </main>
  );
}

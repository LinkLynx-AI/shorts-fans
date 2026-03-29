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
          App Router の `not-found.tsx` を root に置いて、未定義 URL でも shell ごと崩れない基盤にしています。
        </p>
        <div className="mt-8 flex flex-wrap gap-3">
          <Button asChild>
            <Link href="/shorts">shorts を開く</Link>
          </Button>
          <Button asChild variant="secondary">
            <Link href="/home">home に戻る</Link>
          </Button>
        </div>
      </section>
    </main>
  );
}

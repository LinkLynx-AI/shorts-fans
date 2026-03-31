"use client";

import Link from "next/link";
import { useEffect } from "react";

import { Button } from "@/shared/ui";

export default function Error({
  error,
  unstable_retry,
}: {
  error: Error & { digest?: string };
  unstable_retry: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <main className="flex min-h-screen items-center justify-center px-6 py-12">
      <section className="w-full max-w-xl rounded-[2rem] border border-white/80 bg-white/80 p-8 shadow-[0_24px_80px_rgba(87,38,8,0.14)] backdrop-blur">
        <p className="text-sm font-semibold uppercase tracking-[0.24em] text-accent">runtime boundary</p>
        <h1 className="mt-4 text-3xl font-semibold tracking-tight text-foreground">
          不意の例外で route shell の描画が止まりました。
        </h1>
        <p className="mt-3 text-sm leading-7 text-muted">
          想定外のエラーはここで受けて再試行できるようにしておきます。
          digest は server 側ログの突合用です。
        </p>
        {error.digest ? (
          <p className="mt-4 rounded-full bg-stone-900 px-4 py-2 text-xs font-medium tracking-[0.18em] text-stone-100">
            digest: {error.digest}
          </p>
        ) : null}
        <div className="mt-8 flex flex-wrap gap-3">
          <Button onClick={() => unstable_retry()}>再試行する</Button>
          <Button asChild variant="secondary">
            <Link href="/">shorts に戻る</Link>
          </Button>
        </div>
      </section>
    </main>
  );
}

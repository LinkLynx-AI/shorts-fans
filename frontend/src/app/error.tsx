"use client";

import Link from "next/link";
import { useEffect } from "react";

import { Button, SurfacePanel } from "@/shared/ui";

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
      <SurfacePanel className="w-full max-w-xl px-8 py-9">
        <p className="font-display text-[11px] font-semibold uppercase tracking-[0.24em] text-accent">runtime boundary</p>
        <h1 className="mt-4 font-display text-3xl font-semibold tracking-[-0.05em] text-foreground">
          route shell の描画が一時的に止まりました。
        </h1>
        <p className="mt-3 text-sm leading-7 text-muted">
          想定外の例外はここで受けて再試行できます。digest は server 側ログの突合用です。
        </p>
        {error.digest ? (
          <p className="mt-5 rounded-full bg-accent-strong px-4 py-2 text-xs font-semibold tracking-[0.18em] text-white">
            digest: {error.digest}
          </p>
        ) : null}
        <div className="mt-8 flex flex-wrap gap-3">
          <Button onClick={() => unstable_retry()}>再試行する</Button>
          <Button asChild variant="secondary">
            <Link href="/">shorts に戻る</Link>
          </Button>
        </div>
      </SurfacePanel>
    </main>
  );
}

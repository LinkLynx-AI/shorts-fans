import Link from "next/link";

import { Button, SurfacePanel } from "@/shared/ui";

export default function NotFound() {
  return (
    <main className="flex min-h-screen items-center justify-center px-6 py-12">
      <SurfacePanel className="w-full max-w-xl px-8 py-9">
        <p className="font-display text-[11px] font-semibold uppercase tracking-[0.24em] text-accent">404 / route miss</p>
        <h1 className="mt-4 font-display text-3xl font-semibold tracking-[-0.05em] text-foreground">
          指定された surface はまだ用意されていません。
        </h1>
        <p className="mt-3 text-sm leading-7 text-muted">
          fan shell の対象外 URL、または fixture に存在しない dynamic route です。
        </p>
        <div className="mt-8 flex flex-wrap gap-3">
          <Button asChild>
            <Link href="/">feed に戻る</Link>
          </Button>
        </div>
      </SurfacePanel>
    </main>
  );
}

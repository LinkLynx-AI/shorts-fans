import Link from "next/link";

import { Button, SurfacePanel } from "@/shared/ui";

/**
 * protected fan surface から到達する fan login entry を表示する。
 */
export function FanAuthEntryShell() {
  return (
    <main className="mx-auto flex min-h-svh w-full max-w-[408px] items-center px-4 py-10">
      <SurfacePanel className="w-full px-5 py-6 text-foreground">
        <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent">fan access</p>
        <h1 className="mt-3 font-display text-[30px] font-semibold leading-[1.08] tracking-[-0.04em]">
          続けるにはログインが必要です
        </h1>
        <p className="mt-3 text-sm leading-6 text-muted">
          フォロー中、library、main 再生のような protected fan surface は、fan session を開始してから続けられます。
        </p>

        <div className="mt-5 grid gap-2.5">
          <Button className="w-full" disabled type="button">
            サインイン / 新規登録
          </Button>
          <Button asChild className="w-full" variant="secondary">
            <Link href="/">feed に戻る</Link>
          </Button>
        </div>

        <p className="mt-3 text-xs leading-5 text-muted">
          sign in / sign up の具体的な modal 接続は後続 issue で行います。この route は protected flow の共通 entry として使います。
        </p>
      </SurfacePanel>
    </main>
  );
}

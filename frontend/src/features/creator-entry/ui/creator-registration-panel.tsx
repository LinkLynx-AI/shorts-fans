"use client";

import Link from "next/link";
import { Button, SurfacePanel } from "@/shared/ui";

import { useCreatorRegistration } from "../model/use-creator-registration";

/**
 * fan profile から始める creator registration confirm panel を表示する。
 */
export function CreatorRegistrationPanel() {
  const { errorMessage, isSubmitting, submit } = useCreatorRegistration();

  return (
    <main className="mx-auto flex min-h-full w-full max-w-[408px] flex-col px-4 pb-28 pt-5">
      <SurfacePanel className="w-full px-5 py-6 text-foreground">
        <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent">creator entry</p>
        <h1 className="mt-3 font-display text-[30px] font-semibold leading-[1.08] tracking-[-0.04em]">
          Creator登録を始める
        </h1>
        <p className="mt-3 text-sm leading-6 text-muted">
          表示名、handle、avatar は signup と profile settings で fan / creator 共通のものを使用します。
          この申込では追加入力なしで creator capability を有効化します。
        </p>

        <form
          className="mt-5 grid gap-3"
          onSubmit={(event) => {
            event.preventDefault();
            void submit();
          }}
        >
          <div className="rounded-[22px] border border-[#d7e7ef] bg-[#f7fbfd] px-4 py-4 text-sm leading-6 text-muted">
            <p className="font-semibold text-foreground">申込に使うプロフィール</p>
            <p className="mt-2">
              現在の fan profile の表示名、handle、avatar がそのまま creator profile に同期されます。
              変更したい場合は先に signup 時または profile settings で更新してください。
            </p>
          </div>

          {errorMessage ? (
            <p
              aria-live="polite"
              className="rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
              role="alert"
            >
              {errorMessage}
            </p>
          ) : null}

          <Button className="w-full" disabled={isSubmitting} type="submit">
            {isSubmitting ? "登録中..." : "申し込む"}
          </Button>
        </form>

        <div className="mt-3">
          <Button asChild className="w-full" disabled={isSubmitting} variant="secondary">
            <Link href="/fan">あとで戻る</Link>
          </Button>
        </div>
      </SurfacePanel>
    </main>
  );
}

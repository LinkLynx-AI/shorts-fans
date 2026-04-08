"use client";

import Link from "next/link";

import {
  Button,
  SurfacePanel,
} from "@/shared/ui";

import { useCreatorModeEntry } from "../model/use-creator-mode-entry";

/**
 * creator registration 完了後の success surface を表示する。
 */
export function CreatorRegistrationSuccessPanel() {
  const {
    enterCreatorMode,
    errorMessage,
    isSubmitting,
  } = useCreatorModeEntry();

  return (
    <main className="mx-auto flex min-h-full w-full max-w-[408px] flex-col px-4 pb-28 pt-5">
      <SurfacePanel className="w-full px-5 py-6 text-foreground">
        <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent">
          creator ready
        </p>
        <h1 className="mt-3 font-display text-[30px] font-semibold leading-[1.08] tracking-[-0.04em]">
          Creator登録が完了しました
        </h1>
        <p className="mt-3 text-sm leading-6 text-muted">
          審査なしの最小実装なので、申込完了後すぐに creator mode を開けます。準備ができたら次へ進んでください。
        </p>

        {errorMessage ? (
          <p
            aria-live="polite"
            className="mt-5 rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
            role="alert"
          >
            {errorMessage}
          </p>
        ) : null}

        <div className="mt-5 grid gap-3">
          <Button
            className="w-full"
            disabled={isSubmitting}
            onClick={() => {
              void enterCreatorMode();
            }}
            type="button"
          >
            {isSubmitting ? "Creator mode を開いています..." : "Creator mode を開く"}
          </Button>

          <Button asChild className="w-full" disabled={isSubmitting} variant="secondary">
            <Link href="/fan">あとで fan に戻る</Link>
          </Button>
        </div>
      </SurfacePanel>
    </main>
  );
}

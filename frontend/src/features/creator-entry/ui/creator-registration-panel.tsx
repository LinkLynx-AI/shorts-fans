"use client";

import Link from "next/link";

import {
  Button,
  SurfacePanel,
} from "@/shared/ui";

import { useCreatorRegistration } from "../model/use-creator-registration";

/**
 * fan profile から始める creator registration form を表示する。
 */
export function CreatorRegistrationPanel() {
  const {
    avatar,
    avatarInputKey,
    bio,
    clearAvatarSelection,
    displayName,
    errorMessage,
    handle,
    isSubmitting,
    selectAvatarFile,
    setBio,
    setDisplayName,
    setHandle,
    submit,
  } = useCreatorRegistration();

  return (
    <main className="mx-auto flex min-h-full w-full max-w-[408px] flex-col px-4 pb-28 pt-5">
      <SurfacePanel className="w-full px-5 py-6 text-foreground">
        <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent">
          creator entry
        </p>
        <h1 className="mt-3 font-display text-[30px] font-semibold leading-[1.08] tracking-[-0.04em]">
          Creator登録を始める
        </h1>
        <p className="mt-3 text-sm leading-6 text-muted">
          この最小実装では申込完了後すぐに creator mode を使えます。ここでは表示名、unique な handle、
          自己紹介に加えて、任意で avatar を登録時にアップロードできます。
        </p>

        <form
          className="mt-5 grid gap-3"
          onSubmit={(event) => {
            event.preventDefault();
            void submit();
          }}
        >
          <label className="grid gap-1.5">
            <span className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
              Display name
            </span>
            <input
              autoComplete="nickname"
              className="min-h-12 rounded-[18px] border border-[#bae7ff]/90 bg-white/88 px-4 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-accent focus:ring-4 focus:ring-ring/60"
              disabled={isSubmitting}
              onChange={(event) => setDisplayName(event.target.value)}
              placeholder="Mina Rei"
              type="text"
              value={displayName}
            />
          </label>

          <div className="grid gap-1.5">
            <label
              className="grid gap-1.5"
              htmlFor="creator-registration-handle"
            >
              <span className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
                Handle
              </span>
            </label>
            <input
              aria-describedby="creator-registration-handle-help"
              autoCapitalize="none"
              autoCorrect="off"
              className="min-h-12 rounded-[18px] border border-[#bae7ff]/90 bg-white/88 px-4 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-accent focus:ring-4 focus:ring-ring/60"
              disabled={isSubmitting}
              id="creator-registration-handle"
              onChange={(event) => setHandle(event.target.value)}
              placeholder="@minarei"
              spellCheck={false}
              type="text"
              value={handle}
            />
            <p
              className="text-xs leading-5 text-muted"
              id="creator-registration-handle-help"
            >
              creator ごとに unique です。`@` は省略可、使える文字は英数字・`.`・`_` です。
            </p>
          </div>

          <section className="rounded-[20px] border border-dashed border-[#c9d8e1] bg-[#f7fafc] px-4 py-3">
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
                  Avatar
                </p>
                <p className="mt-1 text-sm leading-6 text-muted">
                  任意です。JPEG / PNG / WebP の 5MB 以下の画像を 1 枚だけ登録できます。
                </p>
              </div>
              {avatar.canClear ? (
                <button
                  className="rounded-full border border-[#c9d8e1] px-3 py-1 text-xs font-semibold text-muted transition hover:border-accent hover:text-accent disabled:cursor-not-allowed disabled:opacity-60"
                  disabled={isSubmitting}
                  onClick={() => clearAvatarSelection()}
                  type="button"
                >
                  画像を外す
                </button>
              ) : null}
            </div>

            <label
              className="mt-3 grid gap-1.5"
              htmlFor="creator-registration-avatar"
            >
              <span className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
                Avatar image
              </span>
            </label>
            <input
              accept={avatar.inputAccept}
              className="mt-1 block w-full rounded-[18px] border border-[#bae7ff]/90 bg-white/88 px-4 py-3 text-sm text-foreground file:mr-3 file:rounded-full file:border-0 file:bg-accent file:px-3 file:py-1.5 file:text-sm file:font-semibold file:text-accent-foreground outline-none transition focus:border-accent focus:ring-4 focus:ring-ring/60"
              disabled={isSubmitting}
              id="creator-registration-avatar"
              key={avatarInputKey}
              onChange={(event) => selectAvatarFile(event.target.files?.[0] ?? null)}
              type="file"
            />
            {avatar.fileName ? (
              <p className="mt-2 text-xs leading-5 text-muted">
                選択中: {avatar.fileName}
              </p>
            ) : null}
            <p
              aria-live="polite"
              className={`mt-2 text-sm leading-6 ${avatar.isError ? "text-[#b2394f]" : "text-muted"}`}
            >
              {avatar.message}
            </p>
          </section>

          <label className="grid gap-1.5">
            <span className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
              Bio
            </span>
            <textarea
              className="min-h-28 rounded-[18px] border border-[#bae7ff]/90 bg-white/88 px-4 py-3 text-sm leading-6 text-foreground outline-none transition placeholder:text-muted focus:border-accent focus:ring-4 focus:ring-ring/60"
              disabled={isSubmitting}
              onChange={(event) => setBio(event.target.value)}
              placeholder="quiet rooftop の continuation を中心に投稿します。"
              rows={4}
              value={bio}
            />
          </label>

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

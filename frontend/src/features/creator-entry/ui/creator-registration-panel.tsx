"use client";

import Link from "next/link";
import {
  AlertCircle,
  Camera,
  CheckCheck,
  ImagePlus,
  LoaderCircle,
} from "lucide-react";

import { getCreatorInitials } from "@/entities/creator";
import { cn } from "@/shared/lib";

import {
  Avatar,
  AvatarFallback,
  Button,
  SurfacePanel,
} from "@/shared/ui";

import { useCreatorRegistration } from "../model/use-creator-registration";

function buildAvatarSurfaceClasses(kind: "empty" | "invalid" | "selected" | "uploading" | "completed" | "failed") {
  switch (kind) {
    case "invalid":
    case "failed":
      return "border-[#f2ccd5] bg-[#fff7f8]";
    case "uploading":
      return "border-[#cfe4ef] bg-[#f5fbfe]";
    case "completed":
      return "border-[#cfe5d8] bg-[#f6fbf7]";
    case "selected":
      return "border-[#d7e7ef] bg-[#f7fbfd]";
    case "empty":
    default:
      return "border-[#dde9ef] bg-[#f8fbfc]";
  }
}

function buildAvatarPreviewClasses(kind: "empty" | "invalid" | "selected" | "uploading" | "completed" | "failed") {
  switch (kind) {
    case "invalid":
    case "failed":
      return "border-[#f0c6d0] bg-[#fdecef] text-[#9b3950]";
    case "uploading":
      return "border-[#c7deea] bg-[#edf7fb] text-[#2f6c82]";
    case "completed":
      return "border-[#c7dfd1] bg-[#eef8f1] text-[#2b6c4d]";
    case "selected":
      return "border-[#d2e3eb] bg-[#eef6fa] text-[#31687e]";
    case "empty":
    default:
      return "border-[#d7e3e9] bg-[#f1f7fa] text-[#3a6678]";
  }
}

function renderAvatarStateIcon(kind: "empty" | "invalid" | "selected" | "uploading" | "completed" | "failed") {
  switch (kind) {
    case "invalid":
    case "failed":
      return <AlertCircle aria-hidden="true" className="size-3.5 shrink-0" strokeWidth={2.2} />;
    case "uploading":
      return <LoaderCircle aria-hidden="true" className="size-3.5 shrink-0 animate-spin" strokeWidth={2.2} />;
    case "completed":
      return <CheckCheck aria-hidden="true" className="size-3.5 shrink-0" strokeWidth={2.2} />;
    case "selected":
      return <ImagePlus aria-hidden="true" className="size-3.5 shrink-0" strokeWidth={2.2} />;
    case "empty":
    default:
      return <Camera aria-hidden="true" className="size-3.5 shrink-0" strokeWidth={2.2} />;
  }
}

function buildAvatarTitle(fileName: string | null) {
  return fileName ?? "未設定";
}

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
  const avatarMonogram = displayName.trim() === "" ? "ME" : getCreatorInitials(displayName);

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

          <section
            className={cn(
              "rounded-[22px] border px-4 py-4 text-foreground transition",
              buildAvatarSurfaceClasses(avatar.kind),
            )}
          >
            <div className="flex items-start gap-4">
              <div className="shrink-0">
                <Avatar
                  className={cn(
                    "size-[72px] rounded-full border text-[19px] font-display font-semibold uppercase tracking-[0.12em] shadow-none",
                    buildAvatarPreviewClasses(avatar.kind),
                  )}
                >
                  {avatar.previewUrl ? (
                    <span
                      aria-label="選択した avatar プレビュー"
                      className="block size-full bg-cover bg-center"
                      role="img"
                      style={{ backgroundImage: `url("${avatar.previewUrl}")` }}
                    />
                  ) : (
                    <AvatarFallback className="bg-transparent text-inherit">
                      {avatarMonogram}
                    </AvatarFallback>
                  )}
                </Avatar>
              </div>

              <div className="min-w-0 flex-1">
                <div className="flex items-center justify-between gap-3">
                  <p className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
                    Avatar
                  </p>
                  <span className="text-xs text-muted">任意</span>
                </div>
                <p className="mt-2 truncate text-sm font-semibold tracking-[-0.02em] text-foreground">
                  {buildAvatarTitle(avatar.fileName)}
                </p>
                <p
                  aria-live="polite"
                  className={cn(
                    "mt-1 inline-flex items-start gap-1.5 text-xs leading-5",
                    avatar.isError ? "text-[#a83853]" : "text-muted",
                  )}
                >
                  {renderAvatarStateIcon(avatar.kind)}
                  {avatar.message}
                </p>
                <div className="mt-3 flex flex-wrap items-center gap-x-3 gap-y-2">
                  <label
                    className={cn(
                      "inline-flex min-h-9 cursor-pointer items-center justify-center rounded-full border border-white/84 bg-white/84 px-3.5 text-[13px] font-semibold text-accent-strong transition hover:bg-white",
                      isSubmitting && "cursor-not-allowed opacity-60",
                    )}
                    htmlFor="creator-registration-avatar"
                  >
                    {avatar.fileName ? "画像を変更" : "画像を選択"}
                  </label>
                  {avatar.canClear ? (
                    <button
                      className="text-[13px] font-semibold text-muted transition hover:text-accent-strong disabled:cursor-not-allowed disabled:opacity-60"
                      disabled={isSubmitting}
                      onClick={() => clearAvatarSelection()}
                      type="button"
                    >
                      外す
                    </button>
                  ) : null}
                  <span className="text-[11px] text-muted">JPEG / PNG / WebP, 5MB</span>
                </div>
              </div>
            </div>

            <label className="sr-only" htmlFor="creator-registration-avatar">
              Avatar image
            </label>
            <input
              accept={avatar.inputAccept}
              className="sr-only"
              disabled={isSubmitting}
              id="creator-registration-avatar"
              key={avatarInputKey}
              onChange={(event) => selectAvatarFile(event.target.files?.[0] ?? null)}
              type="file"
            />
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

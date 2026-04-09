"use client";

import {
  useId,
  type ChangeEvent,
  type FormEvent,
} from "react";

import { cn } from "@/shared/lib";

import {
  type CreatorUploadTransferState,
  getCreatorUploadPendingMessage,
  getCreatorUploadSelectedShortCount,
  getCreatorUploadSubmitLabel,
  isCreatorUploadReady,
  isCreatorUploadSubmitting,
} from "../model/creator-upload-draft";
import { useCreatorUpload } from "../model/use-creator-upload";

function getFieldStateClassName(selected: boolean, tone: "default" | "error" | "success" = "default"): string {
  if (tone === "error") {
    return "inline-flex min-h-7 items-center justify-center rounded-full bg-[#fff0f1] px-3 text-[11px] font-bold tracking-[0.02em] text-[#b2394f]";
  }

  if (tone === "success") {
    return "inline-flex min-h-7 items-center justify-center rounded-full bg-[rgba(26,152,80,0.12)] px-3 text-[11px] font-bold tracking-[0.02em] text-[#197040]";
  }

  return cn(
    "inline-flex min-h-7 items-center justify-center rounded-full px-3 text-[11px] font-bold tracking-[0.02em]",
    selected ? "bg-[rgba(16,130,200,0.12)] text-[#0a5b8c]" : "bg-[rgba(7,19,29,0.06)] text-muted",
  );
}

function getPickerButtonClassName(disabled: boolean): string {
  return cn(
    "inline-flex min-h-[46px] items-center justify-center rounded-full bg-[rgba(16,130,200,0.1)] px-5 text-[13px] font-bold text-[#0a5b8c] transition",
    disabled ? "cursor-not-allowed opacity-45" : "cursor-pointer hover:bg-[rgba(16,130,200,0.16)]",
  );
}

function getTransferStateLabel(
  file: File | null,
  transferState: CreatorUploadTransferState,
): { label: string; tone: "default" | "error" | "success" } {
  if (transferState.kind === "uploaded") {
    return { label: "完了", tone: "success" };
  }

  if (transferState.kind === "failed") {
    return { label: "失敗", tone: "error" };
  }

  if (transferState.kind === "uploading") {
    return { label: "アップロード中", tone: "default" };
  }

  if (file !== null) {
    return { label: "選択済み", tone: "default" };
  }

  return { label: "未選択", tone: "default" };
}

/**
 * creator upload page の upload 接続済み form を表示する。
 */
export function CreatorUploadForm() {
  const {
    addShortSlot,
    draft,
    removeShortSlot,
    selectMainFile,
    selectShortFile,
    submit,
  } = useCreatorUpload();
  const inputIdBase = useId();
  const selectedShortCount = getCreatorUploadSelectedShortCount(draft);
  const ready = isCreatorUploadReady(draft);
  const isSubmitting = isCreatorUploadSubmitting(draft);
  const pendingMessage = getCreatorUploadPendingMessage(draft);
  const submitLabel = getCreatorUploadSubmitLabel(draft);
  const errorMessage = draft.submissionState.kind === "error" ? draft.submissionState.message : null;
  const successState = draft.submissionState.kind === "success" ? draft.submissionState : null;
  const mainStatus = getTransferStateLabel(draft.mainFile, draft.mainTransferState);

  function handleMainFileChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0] ?? null;

    selectMainFile(file);
    event.target.value = "";
  }

  function handleShortFileChange(index: number, event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0] ?? null;

    selectShortFile(index, file);
    event.target.value = "";
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await submit();
  }

  return (
    <form className="grid gap-[18px]" onSubmit={(event) => void handleSubmit(event)}>
      <h1 className="mt-2 text-[28px] font-semibold leading-[1.04] tracking-[-0.04em] text-foreground">
        本編とショートを追加
      </h1>

      <input
        accept="video/*"
        aria-label="本編動画ファイル"
        className="sr-only"
        disabled={isSubmitting}
        id={`${inputIdBase}-main`}
        onChange={handleMainFileChange}
        type="file"
      />

      <div className="grid gap-[14px]">
        <section className="grid gap-[14px] rounded-[20px] border border-[#bae7ff]/90 bg-white/84 p-[14px]">
          <div className="flex items-start justify-between gap-3">
            <div>
              <p className="m-0 text-[11px] font-bold uppercase tracking-[0.12em] text-[#0a5b8c]">main</p>
              <h2 className="mt-1 text-sm font-bold text-foreground">本編動画</h2>
            </div>
            <span className={getFieldStateClassName(draft.mainFile !== null, mainStatus.tone)}>{mainStatus.label}</span>
          </div>

          {draft.mainFile ? (
            <p className="m-0 text-[13px] leading-[1.5] text-foreground">{draft.mainFile.name}</p>
          ) : (
            <p className="m-0 text-[13px] leading-[1.5] text-muted">本編動画を追加してください</p>
          )}

          {draft.mainTransferState.kind === "failed" ? (
            <p className="m-0 text-[12px] leading-[1.5] text-[#b2394f]">{draft.mainTransferState.message}</p>
          ) : null}

          <label className={getPickerButtonClassName(isSubmitting)} htmlFor={`${inputIdBase}-main`}>
            {draft.mainFile ? "本編を選び直す" : "本編を追加"}
          </label>
        </section>

        <section className="grid gap-[14px]">
          <div className="flex items-start justify-between gap-3">
            <div>
              <p className="m-0 text-[11px] font-bold uppercase tracking-[0.12em] text-[#0a5b8c]">shorts</p>
              <h2 className="mt-1 text-sm font-bold text-foreground">ショート動画</h2>
            </div>
            <div className="flex items-center gap-2.5">
              <span className={getFieldStateClassName(selectedShortCount > 0)}>
                {selectedShortCount > 0 ? `${selectedShortCount}本` : "未選択"}
              </span>
              <button
                aria-label="ショート欄を追加"
                className="inline-flex size-7 items-center justify-center bg-transparent text-[30px] font-extralight leading-none text-[#0a5b8c] disabled:cursor-not-allowed disabled:opacity-45"
                disabled={isSubmitting}
                onClick={addShortSlot}
                type="button"
              >
                +
              </button>
            </div>
          </div>

          <div className="grid gap-3">
            {draft.shortSlots.length > 0 ? (
              draft.shortSlots.map((slot, index) => {
                const shortInputId = `${inputIdBase}-short-${slot.id}`;
                const shortStatus = getTransferStateLabel(slot.file, slot.transferState);

                return (
                  <section
                    className="grid gap-[14px] rounded-[20px] border border-[#bae7ff]/90 bg-white/84 p-[14px]"
                    key={shortInputId}
                  >
                    <input
                      accept="video/*"
                      aria-label={`ショート動画 ${index + 1} ファイル`}
                      className="sr-only"
                      disabled={isSubmitting}
                      id={shortInputId}
                      onChange={(event) => {
                        handleShortFileChange(index, event);
                      }}
                      type="file"
                    />

                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <p className="m-0 text-[11px] font-bold uppercase tracking-[0.12em] text-[#0a5b8c]">
                          {`short ${index + 1}`}
                        </p>
                        <h3 className="mt-1 text-sm font-bold text-foreground">{`ショート動画 ${index + 1}`}</h3>
                      </div>
                      <div className="flex items-center gap-2.5">
                        <span className={getFieldStateClassName(slot.file !== null, shortStatus.tone)}>
                          {shortStatus.label}
                        </span>
                        <button
                          aria-label="ショート欄を削除"
                          className="inline-flex size-7 items-center justify-center bg-transparent text-[28px] font-extralight leading-none text-muted disabled:cursor-not-allowed disabled:opacity-45"
                          disabled={isSubmitting}
                          onClick={() => {
                            removeShortSlot(index);
                          }}
                          type="button"
                        >
                          -
                        </button>
                      </div>
                    </div>

                    {slot.file ? (
                      <p className="m-0 text-[13px] leading-[1.5] text-foreground">{slot.file.name}</p>
                    ) : (
                      <p className="m-0 text-[13px] leading-[1.5] text-muted">ショート動画を追加してください</p>
                    )}

                    {slot.transferState.kind === "failed" ? (
                      <p className="m-0 text-[12px] leading-[1.5] text-[#b2394f]">{slot.transferState.message}</p>
                    ) : null}

                    <label className={getPickerButtonClassName(isSubmitting)} htmlFor={shortInputId}>
                      {slot.file ? "ショートを選び直す" : "ショートを追加"}
                    </label>
                  </section>
                );
              })
            ) : (
              <p className="m-0 text-[13px] leading-[1.5] text-muted">ショート動画を追加してください</p>
            )}
          </div>
        </section>
      </div>

      {pendingMessage ? (
        <p
          aria-live="polite"
          className="rounded-[18px] border border-[#d7e7ee] bg-[#f5fbfd] px-4 py-3 text-sm leading-6 text-[#0f6172]"
        >
          {pendingMessage}
        </p>
      ) : null}

      {successState ? (
        <section className="grid gap-2 rounded-[18px] border border-[rgba(26,152,80,0.22)] bg-[rgba(245,255,249,0.95)] px-4 py-3 text-sm leading-6 text-[#197040]">
          <p aria-live="polite" className="m-0 font-semibold">
            ドラフトの作成まで完了しました。
          </p>
          <p className="m-0">
            {`main 1本 / short ${successState.shortIds.length}本 を保存しました。公開や審査提出はまだ行われていません。`}
          </p>
        </section>
      ) : null}

      {errorMessage ? (
        <p
          aria-live="polite"
          className="rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
          role="alert"
        >
          {`${errorMessage} 再試行するか、ファイルを選び直してください。`}
        </p>
      ) : null}

      <div className="mt-2">
        <button
          className="inline-flex min-h-[46px] w-full items-center justify-center rounded-full bg-[#0a5b8c] px-5 text-[13px] font-bold text-white shadow-[0_18px_44px_rgba(16,130,200,0.2)] transition hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-40"
          disabled={!ready || isSubmitting || successState !== null}
          type="submit"
        >
          {submitLabel}
        </button>
      </div>
    </form>
  );
}

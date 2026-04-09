"use client";

import {
  useId,
  useState,
  type ChangeEvent,
  type FormEvent,
} from "react";

import { cn } from "@/shared/lib";

import {
  addCreatorUploadShortSlot,
  createInitialCreatorUploadDraft,
  getCreatorUploadSelectedShortCount,
  isCreatorUploadReady,
  removeCreatorUploadShortSlot,
  setCreatorUploadMainFile,
  setCreatorUploadShortFile,
  setCreatorUploadSubmissionError,
  startCreatorUploadSubmission,
} from "../model/creator-upload-draft";

const pendingSubmissionMessage = "upload package を準備しています...";
const submissionErrorMessage = "アップロード API はまだ接続されていません。UI shell のみ先に実装されています。";

function getFieldStateClassName(selected: boolean): string {
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

/**
 * creator upload page の local-only form を表示する。
 */
export function CreatorUploadForm() {
  const [draft, setDraft] = useState(createInitialCreatorUploadDraft);
  const inputIdBase = useId();
  const selectedShortCount = getCreatorUploadSelectedShortCount(draft);
  const ready = isCreatorUploadReady(draft);
  const isSubmitting = draft.submissionState.kind === "submitting";
  const errorMessage = draft.submissionState.kind === "error" ? draft.submissionState.message : null;

  function handleMainFileChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0] ?? null;

    setDraft((currentDraft) => setCreatorUploadMainFile(currentDraft, file));
    event.target.value = "";
  }

  function handleShortFileChange(index: number, event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0] ?? null;

    setDraft((currentDraft) => setCreatorUploadShortFile(currentDraft, index, file));
    event.target.value = "";
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!ready || isSubmitting) {
      return;
    }

    setDraft((currentDraft) => startCreatorUploadSubmission(currentDraft));

    await new Promise<void>((resolve) => {
      window.setTimeout(() => resolve(), 300);
    });

    setDraft((currentDraft) => setCreatorUploadSubmissionError(currentDraft, submissionErrorMessage));
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
            <span className={getFieldStateClassName(draft.mainFile !== null)}>
              {draft.mainFile ? "選択済み" : "未選択"}
            </span>
          </div>

          {draft.mainFile ? (
            <p className="m-0 text-[13px] leading-[1.5] text-foreground">{draft.mainFile.name}</p>
          ) : (
            <p className="m-0 text-[13px] leading-[1.5] text-muted">本編動画を追加してください</p>
          )}

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
                onClick={() => {
                  setDraft((currentDraft) => addCreatorUploadShortSlot(currentDraft));
                }}
                type="button"
              >
                +
              </button>
            </div>
          </div>

          <div className="grid gap-3">
            {draft.shortFiles.length > 0 ? (
              draft.shortFiles.map((file, index) => {
                const shortInputId = `${inputIdBase}-short-${index}`;

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
                        <span className={getFieldStateClassName(file !== null)}>
                          {file ? "選択済み" : "未選択"}
                        </span>
                        <button
                          aria-label="ショート欄を削除"
                          className="inline-flex size-7 items-center justify-center bg-transparent text-[28px] font-extralight leading-none text-muted disabled:cursor-not-allowed disabled:opacity-45"
                          disabled={isSubmitting}
                          onClick={() => {
                            setDraft((currentDraft) => removeCreatorUploadShortSlot(currentDraft, index));
                          }}
                          type="button"
                        >
                          -
                        </button>
                      </div>
                    </div>

                    {file ? (
                      <p className="m-0 text-[13px] leading-[1.5] text-foreground">{file.name}</p>
                    ) : (
                      <p className="m-0 text-[13px] leading-[1.5] text-muted">ショート動画を追加してください</p>
                    )}

                    <label className={getPickerButtonClassName(isSubmitting)} htmlFor={shortInputId}>
                      {file ? "ショートを選び直す" : "ショートを追加"}
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

      {isSubmitting ? (
        <p
          aria-live="polite"
          className="rounded-[18px] border border-[#d7e7ee] bg-[#f5fbfd] px-4 py-3 text-sm leading-6 text-[#0f6172]"
        >
          {pendingSubmissionMessage}
        </p>
      ) : null}

      {errorMessage ? (
        <p
          aria-live="polite"
          className="rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
          role="alert"
        >
          {errorMessage}
        </p>
      ) : null}

      <div className="mt-2">
        <button
          className="inline-flex min-h-[46px] w-full items-center justify-center rounded-full bg-[#0a5b8c] px-5 text-[13px] font-bold text-white shadow-[0_18px_44px_rgba(16,130,200,0.2)] transition hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-40"
          disabled={!ready || isSubmitting}
          type="submit"
        >
          {isSubmitting ? "接続準備中..." : "アップロード"}
        </button>
      </div>
    </form>
  );
}

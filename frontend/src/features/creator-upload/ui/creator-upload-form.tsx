"use client";

import {
  Plus,
  Trash2,
  UploadCloud,
  Video,
} from "lucide-react";
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

type UploadPickerProps = {
  compact?: boolean;
  disabled: boolean;
  fileName: string | null;
  htmlFor: string;
  label: string;
  subtitle?: string;
};

type ConfirmationCheckboxProps = {
  ariaLabel: string;
  checked: boolean;
  disabled: boolean;
  label: string;
  onChange: (checked: boolean) => void;
};

function getFieldStateClassName(selected: boolean, tone: "default" | "error" | "success" = "default"): string {
  if (tone === "error") {
    return "inline-flex min-h-8 items-center justify-center rounded-[10px] bg-[#fff0f1] px-3 text-[12px] font-semibold text-[#b2394f]";
  }

  if (tone === "success") {
    return "inline-flex min-h-8 items-center justify-center rounded-[10px] bg-[rgba(26,152,80,0.12)] px-3 text-[12px] font-semibold text-[#197040]";
  }

  return cn(
    "inline-flex min-h-8 items-center justify-center rounded-[10px] px-3 text-[12px] font-semibold",
    selected ? "bg-[#edf5ff] text-[#5b84eb]" : "bg-[#f2f4f7] text-[#7f8898]",
  );
}

function getPickerAreaClassName(disabled: boolean, compact: boolean): string {
  return cn(
    "flex w-full flex-col items-center justify-center rounded-[22px] border-2 border-dashed border-[#d9e0ea] bg-white text-center transition",
    compact ? "min-h-[124px] px-5 py-6" : "min-h-[154px] px-5 py-7",
    disabled ? "cursor-not-allowed opacity-55" : "cursor-pointer hover:bg-[#fafcff]",
  );
}

function getInputClassName(disabled: boolean): string {
  return cn(
    "min-h-[64px] w-full rounded-[20px] border border-transparent bg-[#f5f7fb] px-5 text-[16px] font-semibold text-foreground outline-none transition placeholder:font-medium placeholder:text-[#b2b9c6]",
    disabled ? "cursor-not-allowed opacity-60" : "focus:border-[#d7eaff] focus:bg-white focus:ring-4 focus:ring-[#71b4ea]/15",
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

function getPrimarySubmitLabel(submitLabel: string, submissionStateKind: string): string {
  return submissionStateKind === "idle" ? "保存してアップロード" : submitLabel;
}

function UploadPicker({
  compact = false,
  disabled,
  fileName,
  htmlFor,
  label,
  subtitle,
}: UploadPickerProps) {
  return (
    <div className="grid gap-2.5">
      <label className={getPickerAreaClassName(disabled, compact)} htmlFor={htmlFor}>
        <UploadCloud
          className={cn("mb-2 text-[#6790ff]", compact ? "size-7" : "size-8")}
          strokeWidth={1.9}
        />
        <span className={cn("font-semibold text-[#4e7ff7]", compact ? "text-[14px]" : "text-[15px]")}>{label}</span>
        {subtitle ? <span className="mt-1 text-[12px] font-medium text-[#a7b0bf]">{subtitle}</span> : null}
      </label>

      {fileName ? <p className="px-1 text-[12px] font-medium text-muted">{fileName}</p> : null}
    </div>
  );
}

function ConfirmationCheckbox({
  ariaLabel,
  checked,
  disabled,
  label,
  onChange,
}: ConfirmationCheckboxProps) {
  return (
    <label className="flex items-start gap-3 rounded-[20px] border border-[#edf1f6] bg-[#fafbfe] px-4 py-4">
      <input
        aria-label={ariaLabel}
        checked={checked}
        className="mt-1 size-5 rounded-[6px] border border-[#cfd7e3] accent-accent-strong"
        disabled={disabled}
        onChange={(event) => {
          onChange(event.target.checked);
        }}
        type="checkbox"
      />
      <span className="text-[14px] font-medium leading-[1.55] text-muted-strong">{label}</span>
    </label>
  );
}

/**
 * creator upload page の upload 接続済み form を表示する。
 */
export function CreatorUploadForm() {
  const {
    addShortSlot,
    draft,
    removeShortSlot,
    setMainConsentConfirmed,
    setMainOwnershipConfirmed,
    setMainPriceJpyInput,
    setShortCaption,
    selectMainFile,
    selectShortFile,
    submit,
  } = useCreatorUpload();
  const inputIdBase = useId();
  const ready = isCreatorUploadReady(draft);
  const isSubmitting = isCreatorUploadSubmitting(draft);
  const pendingMessage = getCreatorUploadPendingMessage(draft);
  const submitLabel = getPrimarySubmitLabel(getCreatorUploadSubmitLabel(draft), draft.submissionState.kind);
  const errorMessage = draft.submissionState.kind === "error" ? draft.submissionState.message : null;
  const successState = draft.submissionState.kind === "success" ? draft.submissionState : null;
  const mainStatus = getTransferStateLabel(draft.mainFile, draft.mainTransferState);
  const shortSlotCountLabel = `${getCreatorUploadSelectedShortCount(draft)} videos`;

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

  function handleMainPriceJpyInputChange(event: ChangeEvent<HTMLInputElement>) {
    setMainPriceJpyInput(event.target.value);
  }

  function handleShortCaptionChange(index: number, event: ChangeEvent<HTMLInputElement>) {
    setShortCaption(index, event.target.value);
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await submit();
  }

  return (
    <form className="flex min-h-0 flex-1 flex-col" onSubmit={(event) => void handleSubmit(event)}>
      <div className="min-h-0 flex-1 overflow-y-auto px-4 pb-36 pt-7">
        <div className="px-1">
          <h1 className="text-[25px] font-extrabold leading-[1.18] tracking-[-0.05em] text-foreground">
            本編とショートを追加
          </h1>
          <p className="mt-2 text-[14px] font-medium text-muted">動画ファイルと詳細情報を入力してください</p>
        </div>

        <input
          accept="video/*"
          aria-label="本編動画ファイル"
          className="sr-only"
          disabled={isSubmitting}
          id={`${inputIdBase}-main`}
          onChange={handleMainFileChange}
          type="file"
        />

        <div className="mt-7 grid gap-6">
          <section className="rounded-[32px] border border-[#eef2f6] bg-white px-4 py-5 shadow-[0_18px_40px_rgba(23,32,51,0.05)]">
            <div className="flex items-start justify-between gap-3">
              <div className="flex items-center gap-2">
                <Video className="size-[18px] shrink-0 text-[#6890ff]" strokeWidth={2.1} />
                <div className="flex items-baseline gap-2">
                  <p className="m-0 text-[15px] font-bold uppercase tracking-[0.02em] text-foreground">MAIN</p>
                  <p className="m-0 text-[14px] font-semibold text-muted">本編動画</p>
                </div>
              </div>
              <span className={getFieldStateClassName(draft.mainFile !== null, mainStatus.tone)}>{mainStatus.label}</span>
            </div>

            <p className="mt-5 text-[14px] font-medium text-muted">
              {draft.mainFile ? "本編動画を選択しました" : "本編動画を追加してください"}
            </p>

            <div className="mt-4">
              <UploadPicker
                disabled={isSubmitting}
                fileName={draft.mainFile?.name ?? null}
                htmlFor={`${inputIdBase}-main`}
                label={draft.mainFile ? "本編を選び直す" : "本編を追加"}
                subtitle="MP4, MOV / 最大10GB"
              />
            </div>

            {draft.mainTransferState.kind === "failed" ? (
              <p className="mt-3 text-[13px] font-medium leading-[1.5] text-[#b2394f]">
                {draft.mainTransferState.message}
              </p>
            ) : null}

            <div className="mt-6 grid gap-4">
              <div className="grid gap-2">
                <label className="text-[15px] font-bold tracking-[-0.02em] text-foreground" htmlFor={`${inputIdBase}-price-jpy`}>
                  価格（円）
                </label>
                <div className="relative">
                  <input
                    aria-label="価格（円）"
                    className={cn(getInputClassName(isSubmitting), "pr-14")}
                    disabled={isSubmitting}
                    id={`${inputIdBase}-price-jpy`}
                    inputMode="numeric"
                    min="1"
                    onChange={handleMainPriceJpyInputChange}
                    pattern="[0-9]*"
                    placeholder="1800"
                    step="1"
                    type="number"
                    value={draft.mainPriceJpyInput}
                  />
                  <span className="pointer-events-none absolute right-5 top-1/2 -translate-y-1/2 text-[26px] font-bold text-[#aeb6c3]">
                    ¥
                  </span>
                </div>
              </div>

              <ConfirmationCheckbox
                ariaLabel="本編の権利確認"
                checked={draft.mainOwnershipConfirmed}
                disabled={isSubmitting}
                label="本編の販売に必要な権利を自分が保有していることを確認しました。"
                onChange={setMainOwnershipConfirmed}
              />

              <ConfirmationCheckbox
                ariaLabel="本編の同意確認"
                checked={draft.mainConsentConfirmed}
                disabled={isSubmitting}
                label="出演者全員の公開・販売同意が取得済みであることを確認しました。"
                onChange={setMainConsentConfirmed}
              />
            </div>
          </section>

          <section className="grid gap-4">
            <div className="flex items-center justify-between gap-3 px-1">
              <div className="flex items-baseline gap-2">
                <p className="m-0 text-[15px] font-bold uppercase tracking-[0.02em] text-foreground">SHORTS</p>
                <p className="m-0 text-[14px] font-semibold text-muted">ショート動画</p>
              </div>

              <div className="flex items-center gap-3">
                <span className="inline-flex min-h-8 items-center justify-center rounded-[10px] bg-[#f2f4f7] px-3 text-[12px] font-semibold text-[#6f7786]">
                  {shortSlotCountLabel}
                </span>
                <button
                  aria-label="ショート欄を追加"
                  className="inline-flex size-9 items-center justify-center rounded-full bg-[#eef6ff] text-[#5c84eb] transition hover:bg-[#e2efff] disabled:cursor-not-allowed disabled:opacity-45"
                  disabled={isSubmitting}
                  onClick={addShortSlot}
                  type="button"
                >
                  <Plus className="size-[18px]" strokeWidth={2.4} />
                </button>
              </div>
            </div>

            <div className="grid gap-4">
              {draft.shortSlots.length > 0 ? (
                draft.shortSlots.map((slot, index) => {
                  const shortInputId = `${inputIdBase}-short-${slot.id}`;
                  const shortStatus = getTransferStateLabel(slot.file, slot.transferState);

                  return (
                    <section
                      className="rounded-[32px] border border-[#eef2f6] bg-white px-4 py-5 shadow-[0_18px_40px_rgba(23,32,51,0.05)]"
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
                        <div className="flex items-baseline gap-2">
                          <p className="m-0 text-[15px] font-bold uppercase tracking-[0.02em] text-[#5c84eb]">
                            {`SHORT ${index + 1}`}
                          </p>
                          <p className="m-0 text-[14px] font-semibold text-foreground">{`ショート動画 ${index + 1}`}</p>
                        </div>

                        <div className="flex items-center gap-2">
                          <span className={getFieldStateClassName(slot.file !== null, shortStatus.tone)}>
                            {shortStatus.label}
                          </span>
                          {draft.shortSlots.length > 1 ? (
                            <button
                              aria-label="ショート欄を削除"
                              className="inline-flex size-8 items-center justify-center rounded-full text-[#a4acb8] transition hover:bg-[#fff3f4] hover:text-[#cc5466] disabled:cursor-not-allowed disabled:opacity-45"
                              disabled={isSubmitting}
                              onClick={() => {
                                removeShortSlot(index);
                              }}
                              type="button"
                            >
                              <Trash2 className="size-4" strokeWidth={2} />
                            </button>
                          ) : null}
                        </div>
                      </div>

                      <p className="mt-5 text-[14px] font-medium text-muted">
                        {slot.file ? `ショート動画 ${index + 1} を選択しました` : "ショート動画を追加してください"}
                      </p>

                      <div className="mt-4">
                        <UploadPicker
                          compact
                          disabled={isSubmitting}
                          fileName={slot.file?.name ?? null}
                          htmlFor={shortInputId}
                          label={slot.file ? "ショートを選び直す" : "ショートを追加"}
                        />
                      </div>

                      {slot.transferState.kind === "failed" ? (
                        <p className="mt-3 text-[13px] font-medium leading-[1.5] text-[#b2394f]">
                          {slot.transferState.message}
                        </p>
                      ) : null}

                      <div className="mt-5 grid gap-2">
                        <label className="text-[15px] font-bold tracking-[-0.02em] text-foreground" htmlFor={`${shortInputId}-caption`}>
                          {`ショート動画 ${index + 1} の caption`}
                          <span className="ml-1 font-medium text-muted">(任意)</span>
                        </label>
                        <input
                          aria-label={`ショート動画 ${index + 1} の caption`}
                          className={getInputClassName(isSubmitting)}
                          disabled={isSubmitting}
                          id={`${shortInputId}-caption`}
                          onChange={(event) => {
                            handleShortCaptionChange(index, event);
                          }}
                          placeholder="quiet rooftop preview."
                          type="text"
                          value={slot.caption}
                        />
                      </div>
                    </section>
                  );
                })
              ) : (
                <p className="px-1 text-[14px] font-medium text-muted">ショート動画を追加してください</p>
              )}
            </div>
          </section>

          {pendingMessage ? (
            <p
              aria-live="polite"
              className="rounded-[22px] border border-[#d7eaff] bg-[#f5fbff] px-4 py-4 text-[14px] font-medium leading-[1.6] text-[#436aa7]"
            >
              {pendingMessage}
            </p>
          ) : null}

          {successState ? (
            <section className="grid gap-2 rounded-[22px] border border-[rgba(26,152,80,0.22)] bg-[rgba(245,255,249,0.95)] px-4 py-4 text-[14px] leading-[1.6] text-[#197040]">
              <p aria-live="polite" className="m-0 font-semibold">
                処理開始を受け付けました。
              </p>
              <p className="m-0 font-medium">
                {`main 1本 / short ${successState.shortIds.length}本 の処理を開始しました。公開や審査提出はまだ行われていません。`}
              </p>
            </section>
          ) : null}

          {errorMessage ? (
            <p
              aria-live="polite"
              className="rounded-[22px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-4 text-[14px] font-medium leading-[1.6] text-[#b2394f]"
              role="alert"
            >
              {`${errorMessage} 再試行するか、ファイルを選び直してください。`}
            </p>
          ) : null}
        </div>
      </div>

      <div className="sticky bottom-0 z-10 border-t border-[#eef2f6] bg-white/95 px-4 pb-6 pt-4 backdrop-blur-sm">
        <button
          className="inline-flex min-h-[56px] w-full items-center justify-center rounded-full bg-accent text-[16px] font-bold text-white shadow-[0_18px_40px_rgba(113,180,234,0.34)] transition hover:bg-accent-strong disabled:cursor-not-allowed disabled:opacity-45"
          disabled={!ready || isSubmitting || successState !== null}
          type="submit"
        >
          {submitLabel}
        </button>
      </div>
    </form>
  );
}

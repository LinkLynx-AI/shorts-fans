"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { useState } from "react";

import { BottomSheetDialogContent, Button } from "@/shared/ui";

import { useCreatorWorkspaceMetadataEdit } from "../model/use-creator-workspace-metadata-edit";
import type { CreatorWorkspacePreviewDetailData } from "./creator-mode-shell.types";

type CreatorWorkspaceMetadataEditSheetProps = {
  detail: CreatorWorkspacePreviewDetailData;
  onDetailSaved: (detail: CreatorWorkspacePreviewDetailData) => void;
  onMainPriceSaved: (mainId: string, priceJpy: number) => void;
};

function parseMainPriceInput(value: string): number | null {
  const normalized = value.trim();

  if (!/^\d+$/.test(normalized)) {
    return null;
  }

  const parsed = Number(normalized);

  if (!Number.isSafeInteger(parsed) || parsed <= 0) {
    return null;
  }

  return parsed;
}

/**
 * creator workspace detail から price / caption を編集する bottom sheet を表示する。
 */
export function CreatorWorkspaceMetadataEditSheet({
  detail,
  onDetailSaved,
  onMainPriceSaved,
}: CreatorWorkspaceMetadataEditSheetProps) {
  const [captionValue, setCaptionValue] = useState(
    detail.kind === "preview-short" ? detail.detail.short.caption : "",
  );
  const [mainPriceValue, setMainPriceValue] = useState(
    detail.kind === "preview-main" ? String(detail.detail.main.priceJpy) : "",
  );
  const [open, setOpen] = useState(false);
  const [validationMessage, setValidationMessage] = useState<string | null>(null);
  const {
    clearError,
    errorMessage,
    isSubmitting,
    saveMainPrice,
    saveShortCaption,
  } = useCreatorWorkspaceMetadataEdit();

  const resetDraft = () => {
    clearError();
    setValidationMessage(null);

    if (detail.kind === "preview-main") {
      setMainPriceValue(String(detail.detail.main.priceJpy));
      return;
    }

    setCaptionValue(detail.detail.short.caption);
  };

  const handleOpenChange = (nextOpen: boolean) => {
    if (nextOpen) {
      resetDraft();
    }

    setOpen(nextOpen);
  };

  const handleSave = async () => {
    setValidationMessage(null);

    if (detail.kind === "preview-main") {
      const priceJpy = parseMainPriceInput(mainPriceValue);

      if (priceJpy === null) {
        setValidationMessage("価格は1円以上の整数で入力してください。");
        return;
      }

      const saved = await saveMainPrice(detail.detail.main.id, priceJpy);

      if (!saved) {
        return;
      }

      onDetailSaved({
        detail: {
          ...detail.detail,
          main: {
            ...detail.detail.main,
            priceJpy,
          },
        },
        kind: "preview-main",
      });
      onMainPriceSaved(detail.detail.main.id, priceJpy);
      setOpen(false);
      return;
    }

    const normalizedCaption = captionValue.trim();
    const saved = await saveShortCaption(detail.detail.short.id, normalizedCaption === "" ? null : normalizedCaption);

    if (!saved) {
      return;
    }

    onDetailSaved({
      detail: {
        ...detail.detail,
        short: {
          ...detail.detail.short,
          caption: normalizedCaption,
        },
      },
      kind: "preview-short",
    });
    setOpen(false);
  };

  const message = validationMessage ?? errorMessage;
  const actionLabel = detail.kind === "preview-main" ? "価格を編集" : "caption を編集";
  const submitLabel = detail.kind === "preview-main"
    ? (isSubmitting ? "価格を保存しています..." : "価格を保存")
    : (isSubmitting ? "caption を保存しています..." : "caption を保存");

  return (
    <Dialog.Root onOpenChange={handleOpenChange} open={open}>
      <Dialog.Trigger asChild>
        <button
          aria-label="投稿操作"
          className="inline-flex min-h-8 min-w-7 items-center justify-center gap-1 bg-transparent text-[#1082c8] transition hover:bg-[#1082c8]/10"
          type="button"
        >
          <span className="size-1 rounded-full bg-current" />
          <span className="size-1 rounded-full bg-current" />
          <span className="size-1 rounded-full bg-current" />
        </button>
      </Dialog.Trigger>

      <BottomSheetDialogContent description={`${actionLabel}を行うメニュー`} title={actionLabel}>
        <div className="rounded-[24px] bg-[#f3f6f8] p-4">
          <div className="grid gap-3">
            <div className="grid gap-1">
              <p className="m-0 text-sm font-bold text-foreground">{actionLabel}</p>
              <p className="m-0 text-[12px] leading-5 text-muted">
                {detail.kind === "preview-main"
                  ? "本編 price を owner preview から更新します。"
                  : "ショート caption を owner preview から更新します。"}
              </p>
            </div>

            {detail.kind === "preview-main" ? (
              <label className="grid gap-2 text-sm font-semibold text-foreground">
                <span>価格（円）</span>
                <input
                  className="min-h-11 rounded-[16px] border border-[rgba(167,220,249,0.48)] bg-white px-4 text-[15px] text-foreground outline-none transition focus:border-[#1082c8]"
                  inputMode="numeric"
                  onChange={(event) => {
                    clearError();
                    setValidationMessage(null);
                    setMainPriceValue(event.target.value);
                  }}
                  value={mainPriceValue}
                />
              </label>
            ) : (
              <label className="grid gap-2 text-sm font-semibold text-foreground">
                <span>caption</span>
                <textarea
                  className="min-h-[116px] rounded-[16px] border border-[rgba(167,220,249,0.48)] bg-white px-4 py-3 text-[15px] leading-6 text-foreground outline-none transition focus:border-[#1082c8]"
                  onChange={(event) => {
                    clearError();
                    setValidationMessage(null);
                    setCaptionValue(event.target.value);
                  }}
                  value={captionValue}
                />
              </label>
            )}
          </div>
        </div>

        {message ? (
          <p
            aria-live="polite"
            className="mt-3 rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
            role="alert"
          >
            {message}
          </p>
        ) : null}

        <div className="mt-3 flex gap-2">
          <Dialog.Close asChild>
            <Button className="flex-1" disabled={isSubmitting} type="button" variant="secondary">
              キャンセル
            </Button>
          </Dialog.Close>
          <Button className="flex-1" disabled={isSubmitting} onClick={() => void handleSave()} type="button">
            {submitLabel}
          </Button>
        </div>
      </BottomSheetDialogContent>
    </Dialog.Root>
  );
}

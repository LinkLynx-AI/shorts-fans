"use client";

import * as Dialog from "@radix-ui/react-dialog";
import {
  useId,
  useState,
  type FormEvent,
} from "react";

import { ApiError } from "@/shared/api";
import { cn } from "@/shared/lib";
import { Button } from "@/shared/ui";

import {
  CreatorWorkspaceMainPriceApiError,
  type CreatorWorkspaceMainPriceMutationResult,
  updateCreatorWorkspaceMainPrice,
} from "../api/update-creator-workspace-main-price";
import {
  formatCreatorWorkspaceMainPriceInput,
  parseCreatorWorkspaceMainPriceInput,
} from "../model/creator-workspace-main-price";

export type CreatorWorkspaceMainPriceDialogProps = {
  currentPriceJpy: number;
  mainId: string;
  onClose: () => void;
  onUpdated: (result: CreatorWorkspaceMainPriceMutationResult) => void;
  open: boolean;
};

function buildCreatorWorkspaceMainPriceErrorMessage(error: unknown): string {
  if (error instanceof CreatorWorkspaceMainPriceApiError) {
    return error.message;
  }

  if (error instanceof ApiError && error.code === "network") {
    return "価格を更新できませんでした。通信環境を確認してからやり直してください。";
  }

  return "価格を更新できませんでした。少し時間を置いてからやり直してください。";
}

function getInputClassName(disabled: boolean, invalid: boolean): string {
  return cn(
    "min-h-[52px] w-full rounded-[18px] border bg-white px-4 text-[15px] text-foreground outline-none transition placeholder:text-muted",
    invalid ? "border-[#b2394f] text-[#7f1f33]" : "border-[#d7e7ee]",
    disabled ? "cursor-not-allowed opacity-60" : "focus:border-[#0a5b8c]",
  );
}

/**
 * creator workspace の本編価格変更 modal を表示する。
 */
export function CreatorWorkspaceMainPriceDialog({
  currentPriceJpy,
  mainId,
  onClose,
  onUpdated,
  open,
}: CreatorWorkspaceMainPriceDialogProps) {
  const priceInputID = useId();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [priceInput, setPriceInput] = useState(formatCreatorWorkspaceMainPriceInput(currentPriceJpy));
  const nextPriceJpy = parseCreatorWorkspaceMainPriceInput(priceInput);
  const hasChanged = nextPriceJpy !== null && nextPriceJpy !== currentPriceJpy;
  const isInvalid = priceInput.trim() !== "" && nextPriceJpy === null;
  const isSaveDisabled = isSubmitting || nextPriceJpy === null || !hasChanged;

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (isSaveDisabled || nextPriceJpy === null) {
      return;
    }

    setErrorMessage(null);
    setIsSubmitting(true);

    try {
      const result = await updateCreatorWorkspaceMainPrice({
        mainId,
        priceJpy: nextPriceJpy,
      });

      onUpdated(result);
      onClose();
    } catch (error: unknown) {
      setErrorMessage(buildCreatorWorkspaceMainPriceErrorMessage(error));
      setIsSubmitting(false);
    }
  }

  return (
    <Dialog.Root
      onOpenChange={(nextOpen) => {
        if (!nextOpen) {
          onClose();
        }
      }}
      open={open}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-[#061521]/36 backdrop-blur-[2px]" />
        <Dialog.Content className="fixed inset-x-4 bottom-[104px] z-50 mx-auto max-w-[376px] rounded-[30px] border border-white/72 bg-[rgba(255,255,255,0.9)] p-4 text-foreground shadow-[0_24px_60px_rgba(28,78,114,0.16)] backdrop-blur-xl">
          <form className="grid gap-4" onSubmit={handleSubmit}>
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="m-0 text-[11px] font-bold uppercase tracking-[0.24em] text-accent">main price</p>
                <Dialog.Title className="mt-2 font-display text-[26px] font-semibold leading-[1.08] tracking-[-0.04em]">
                  価格を変更
                </Dialog.Title>
                <Dialog.Description className="mt-2 text-[13px] leading-[1.6] text-muted">
                  本編詳細に戻ったまま、公開価格だけを更新します。
                </Dialog.Description>
              </div>
              <span className="inline-flex min-h-10 items-center rounded-full bg-accent/12 px-3 text-[11px] font-bold uppercase tracking-[0.14em] text-accent">
                現在 ¥{currentPriceJpy.toLocaleString("ja-JP")}
              </span>
            </div>

            <div className="rounded-[22px] border border-[#d7e7ee] bg-white/82 p-4 shadow-[inset_0_0_0_1px_rgba(167,220,249,0.18)]">
              <label className="block text-[12px] font-semibold text-foreground" htmlFor={priceInputID}>
                価格（円）
              </label>
              <div className="mt-2 flex items-center gap-3">
                <span className="text-[18px] font-bold text-accent-strong">¥</span>
                <input
                  aria-invalid={isInvalid}
                  className={getInputClassName(isSubmitting, isInvalid)}
                  disabled={isSubmitting}
                  id={priceInputID}
                  inputMode="numeric"
                  min="1"
                  onChange={(event) => {
                    setPriceInput(event.target.value);
                  }}
                  placeholder="1800"
                  step="1"
                  type="number"
                  value={priceInput}
                />
              </div>
              <p className={cn("mb-0 mt-2 text-[12px] leading-[1.5]", isInvalid ? "text-[#b2394f]" : "text-muted")}>
                1円以上の整数で入力してください。
              </p>
            </div>

            {errorMessage ? (
              <p className="m-0 rounded-[18px] bg-[#fff0f1] px-4 py-3 text-[13px] leading-[1.6] text-[#b2394f]" role="alert">
                {errorMessage}
              </p>
            ) : null}

            <div className="flex gap-2.5">
              <Button disabled={isSubmitting} onClick={onClose} type="button" variant="secondary">
                キャンセル
              </Button>
              <Button disabled={isSaveDisabled} type="submit">
                {isSubmitting ? "保存中..." : "保存する"}
              </Button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

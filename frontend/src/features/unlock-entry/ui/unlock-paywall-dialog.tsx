import * as Dialog from "@radix-ui/react-dialog";
import { cn } from "@/shared/lib";

import { getUnlockCtaMeta } from "../model/unlock-cta";
import type { UnlockSurfaceModel } from "../model/unlock-entry";

export type UnlockPaywallDialogProps = {
  acceptAge: boolean;
  acceptTerms: boolean;
  isSubmitting?: boolean;
  onAcceptAgeChange: (checked: boolean) => void;
  onAcceptTermsChange: (checked: boolean) => void;
  onClose: () => void;
  onConfirm: () => void;
  open: boolean;
  unlock: UnlockSurfaceModel;
};

function buildPaywallTitle(caption: string): string {
  const normalizedCaption = caption.trim().replace(/[。.!?]+$/u, "");

  return normalizedCaption ? `${normalizedCaption} の続きを見る` : "この short の続きを見る";
}

function getUnlockButtonLabel(unlock: UnlockSurfaceModel): string {
  const meta = getUnlockCtaMeta(unlock.unlockCta);
  return meta ? `Unlock ${meta}` : "Unlock";
}

/**
 * 初回視聴用の unlock setup dialog を表示する。
 */
export function UnlockPaywallDialog({
  acceptAge,
  acceptTerms,
  isSubmitting = false,
  onAcceptAgeChange,
  onAcceptTermsChange,
  onClose,
  onConfirm,
  open,
  unlock,
}: UnlockPaywallDialogProps) {
  const confirmEnabled =
    (!unlock.setup.requiresAgeConfirmation || acceptAge) &&
    (!unlock.setup.requiresTermsAcceptance || acceptTerms);
  const title = buildPaywallTitle(unlock.short.caption);
  const buttonLabel = getUnlockButtonLabel(unlock);

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
        <Dialog.Content className="fixed inset-x-4 bottom-[176px] z-50 mx-auto max-w-[376px] rounded-[30px] border border-white/72 bg-[rgba(255,255,255,0.82)] p-4 text-foreground shadow-[0_24px_60px_rgba(28,78,114,0.16)] backdrop-blur-xl">
          <div className="flex items-start justify-between gap-3">
            <div>
              <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent">unlock</p>
              <Dialog.Title className="mt-2 font-display text-[26px] font-semibold leading-[1.08] tracking-[-0.04em]">
                {title}
              </Dialog.Title>
              <Dialog.Description className="sr-only">
                この short の続きを見るための setup dialog
              </Dialog.Description>
            </div>
            <span className="inline-flex min-h-10 items-center rounded-full bg-accent/12 px-3 text-[11px] font-bold uppercase tracking-[0.14em] text-accent">
              ¥{unlock.main.priceJpy.toLocaleString("ja-JP")}
            </span>
          </div>

          <div className="mt-4 flex items-center gap-3 rounded-[20px] border border-[#bae7ff]/90 bg-white/86 px-3.5 py-3.5">
            <span className="inline-flex size-[38px] shrink-0 items-center justify-center rounded-full bg-accent-strong text-[11px] font-bold uppercase tracking-[0.14em] text-white">
              Setup
            </span>
            <span className="min-w-0">
              <p className="truncate text-sm font-bold">この段階では購入はまだ実行しません。</p>
              <p className="mt-0.5 text-xs text-muted">年齢確認と利用規約の同意だけで main 再生へ進みます。</p>
            </span>
          </div>

          <div className="mt-3 grid gap-2">
            {unlock.setup.requiresAgeConfirmation ? (
              <label className="flex items-start gap-2.5 rounded-[18px] bg-white/76 px-3 py-3 text-xs leading-6 text-muted">
                <input
                  checked={acceptAge}
                  className="mt-1"
                  onChange={(event) => onAcceptAgeChange(event.target.checked)}
                  type="checkbox"
                />
                <span>18歳以上であり、年齢確認に同意する</span>
              </label>
            ) : null}
            {unlock.setup.requiresTermsAcceptance ? (
              <label className="flex items-start gap-2.5 rounded-[18px] bg-white/76 px-3 py-3 text-xs leading-6 text-muted">
                <input
                  checked={acceptTerms}
                  className="mt-1"
                  onChange={(event) => onAcceptTermsChange(event.target.checked)}
                  type="checkbox"
                />
                <span>利用規約とポリシーに同意し、main 再生へ進む</span>
              </label>
            ) : null}
          </div>

          <div className="mt-4 flex gap-2.5">
            <Dialog.Close asChild>
              <button
                className="flex min-h-[46px] flex-1 items-center justify-center rounded-full bg-accent/10 px-4 text-[13px] font-bold text-accent-strong"
                disabled={isSubmitting}
                type="button"
              >
                閉じる
              </button>
            </Dialog.Close>
            {confirmEnabled && !isSubmitting ? (
              <button
                className="flex min-h-[46px] flex-1 items-center justify-center rounded-full bg-accent-strong px-4 text-[13px] font-bold text-white"
                onClick={onConfirm}
                type="button"
              >
                {buttonLabel}
              </button>
            ) : (
              <button
                className={cn(
                  "flex min-h-[46px] flex-1 items-center justify-center rounded-full bg-accent-strong px-4 text-[13px] font-bold text-white",
                  "cursor-not-allowed opacity-40",
                )}
                disabled
                type="button"
              >
                {buttonLabel}
              </button>
            )}
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

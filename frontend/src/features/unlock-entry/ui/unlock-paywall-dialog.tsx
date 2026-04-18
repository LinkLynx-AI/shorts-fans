import * as Dialog from "@radix-ui/react-dialog";
import { buildShortPaywallTitle } from "@/entities/short";
import { cn } from "@/shared/lib";

import { getUnlockCtaMeta } from "../model/unlock-cta";
import type { UnlockSurfaceModel } from "../model/unlock-entry";
import { CCBillPaymentWidget } from "./ccbill-payment-widget";

export type PaywallPaymentSelection =
  | {
      mode: "new_card";
    }
  | {
      mode: "saved_card";
      paymentMethodId: string;
    };

type PaymentWidgetSession = {
  apiBaseUrl: string;
  apiKey: string;
  clientAccount: string;
  currency: "JPY";
  initialPeriod: string;
  initialPrice: string;
  sessionToken: string;
  subAccount: string;
};

export type UnlockPaywallDialogProps = {
  acceptAge: boolean;
  acceptTerms: boolean;
  cardSetupErrorMessage?: string | null | undefined;
  cardSetupSession?: PaymentWidgetSession | null | undefined;
  isLoadingCardSetupSession?: boolean;
  isSubmitting?: boolean;
  onAcceptAgeChange: (checked: boolean) => void;
  onAcceptTermsChange: (checked: boolean) => void;
  onCardPaymentTokenCreated: (paymentTokenId: string) => void;
  onClose: () => void;
  onConfirm: () => void;
  onPaymentSelectionChange: (selection: PaywallPaymentSelection) => void;
  open: boolean;
  purchaseErrorMessage?: string | null | undefined;
  selection: PaywallPaymentSelection;
  unlock: UnlockSurfaceModel;
  usePurchaseFlow?: boolean;
};

function getUnlockButtonLabel(unlock: UnlockSurfaceModel): string {
  const meta = getUnlockCtaMeta(unlock.unlockCta);
  return meta ? `Unlock ${meta}` : "Unlock";
}

function getCardBrandLabel(brand: UnlockSurfaceModel["purchase"]["supportedCardBrands"][number]) {
  switch (brand) {
    case "american_express":
      return "American Express";
    case "jcb":
      return "JCB";
    case "mastercard":
      return "Mastercard";
    case "visa":
      return "Visa";
  }
}

function getSavedCardLabel(method: UnlockSurfaceModel["purchase"]["savedPaymentMethods"][number]) {
  return `${getCardBrandLabel(method.brand)} •••• ${method.last4}`;
}

function getPrimaryActionLabel(unlock: UnlockSurfaceModel, selection: PaywallPaymentSelection): string | null {
  switch (unlock.purchase.state) {
    case "already_purchased":
      return "Continue main";
    case "owner_preview":
      return "Owner preview";
    case "purchase_ready":
    case "setup_required":
      if (selection.mode === "saved_card") {
        return `Purchase ¥${unlock.main.priceJpy.toLocaleString("ja-JP")}`;
      }

      return null;
    default:
      return null;
  }
}

function getStateSummary(unlock: UnlockSurfaceModel): { body: string; title: string } {
  switch (unlock.purchase.state) {
    case "already_purchased":
      return {
        body: "この main はすでに購入済みです。再購入ではなく再開導線で進みます。",
        title: "購入済みのため、そのまま再開できます。",
      };
    case "owner_preview":
      return {
        body: "creator owner として preview access が使えます。purchase は実行されません。",
        title: "owner preview で続きへ進めます。",
      };
    case "purchase_pending":
      return {
        body: "provider 側の処理完了待ちです。pending 中は unlock されません。",
        title: "決済処理を確認しています。",
      };
    case "purchase_ready":
      return {
        body: "saved card を使うか、新しい card を入力して purchase できます。",
        title: "4 ブランドの card だけで purchase できます。",
      };
    case "setup_required":
      return {
        body: "main purchase の前に new card setup と required consent が必要です。",
        title: "最初の purchase 前に card setup が必要です。",
      };
    case "unavailable":
      return {
        body: "現在はこの main を purchase できません。",
        title: "現在は利用できません。",
      };
  }
}

/**
 * purchase-aware な mini paywall dialog を表示する。
 */
export function UnlockPaywallDialog({
  acceptAge,
  acceptTerms,
  cardSetupErrorMessage = null,
  cardSetupSession = null,
  isLoadingCardSetupSession = false,
  isSubmitting = false,
  onAcceptAgeChange,
  onAcceptTermsChange,
  onCardPaymentTokenCreated,
  onClose,
  onConfirm,
  onPaymentSelectionChange,
  open,
  purchaseErrorMessage = null,
  selection,
  unlock,
  usePurchaseFlow = true,
}: UnlockPaywallDialogProps) {
  const consentSatisfied =
    (!unlock.setup.requiresAgeConfirmation || acceptAge) &&
    (!unlock.setup.requiresTermsAcceptance || acceptTerms);
  const confirmEnabled = consentSatisfied && !isSubmitting;
  const title = buildShortPaywallTitle(unlock.short.caption);
  const stateSummary = getStateSummary(unlock);
  const buttonLabel = usePurchaseFlow ? getPrimaryActionLabel(unlock, selection) : getUnlockButtonLabel(unlock);
  const supportsSelection =
    usePurchaseFlow && (unlock.purchase.state === "purchase_ready" || unlock.purchase.state === "setup_required");
  const showSavedCards = supportsSelection && unlock.purchase.savedPaymentMethods.length > 0;
  const showNewCardWidget = supportsSelection && selection.mode === "new_card";
  const canRenderNewCardWidget = showNewCardWidget && consentSatisfied;

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
        <Dialog.Content className="fixed inset-x-4 bottom-[176px] z-50 mx-auto max-h-[calc(100vh-220px)] max-w-[392px] overflow-y-auto rounded-[30px] border border-white/72 bg-[rgba(255,255,255,0.86)] p-4 text-foreground shadow-[0_24px_60px_rgba(28,78,114,0.16)] backdrop-blur-xl">
          <div className="flex items-start justify-between gap-3">
            <div>
              <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent">unlock</p>
              <Dialog.Title className="mt-2 font-display text-[26px] font-semibold leading-[1.08] tracking-[-0.04em]">
                {title}
              </Dialog.Title>
              <Dialog.Description className="sr-only">
                この short の続きを見るための purchase dialog
              </Dialog.Description>
            </div>
            <span className="inline-flex min-h-10 items-center rounded-full bg-accent/12 px-3 text-[11px] font-bold uppercase tracking-[0.14em] text-accent">
              ¥{unlock.main.priceJpy.toLocaleString("ja-JP")}
            </span>
          </div>

          <div className="mt-4 rounded-[20px] border border-[#bae7ff]/90 bg-white/86 px-4 py-3.5">
            <p className="text-sm font-bold">{stateSummary.title}</p>
            <p className="mt-1 text-xs leading-6 text-muted">{stateSummary.body}</p>
          </div>

          {unlock.purchase.supportedCardBrands.length > 0 ? (
            <div className="mt-3">
              <p className="text-[11px] font-bold uppercase tracking-[0.16em] text-muted">supported cards</p>
              <div className="mt-2 flex flex-wrap gap-2">
                {unlock.purchase.supportedCardBrands.map((brand) => (
                  <span
                    key={brand}
                    className="inline-flex min-h-9 items-center rounded-full border border-[#bce6f5] bg-white/84 px-3 text-xs font-semibold text-foreground"
                  >
                    {getCardBrandLabel(brand)}
                  </span>
                ))}
              </div>
            </div>
          ) : null}

          {showSavedCards ? (
            <div className="mt-4 grid gap-2">
              {unlock.purchase.savedPaymentMethods.map((method) => {
                const checked =
                  selection.mode === "saved_card" && selection.paymentMethodId === method.paymentMethodId;

                return (
                  <label
                    key={method.paymentMethodId}
                    className={cn(
                      "flex items-center gap-3 rounded-[18px] border bg-white/80 px-3 py-3 text-sm transition",
                      checked ? "border-accent bg-[#f1fbff]" : "border-[#d7eef7]",
                    )}
                  >
                    <input
                      checked={checked}
                      name="paywall-payment-method"
                      onChange={() => {
                        onPaymentSelectionChange({
                          mode: "saved_card",
                          paymentMethodId: method.paymentMethodId,
                        });
                      }}
                      type="radio"
                    />
                    <span className="font-semibold">{getSavedCardLabel(method)}</span>
                  </label>
                );
              })}
            </div>
          ) : null}

          {supportsSelection ? (
            <label
              className={cn(
                "mt-2 flex items-center gap-3 rounded-[18px] border bg-white/80 px-3 py-3 text-sm transition",
                selection.mode === "new_card" ? "border-accent bg-[#f1fbff]" : "border-[#d7eef7]",
              )}
            >
              <input
                checked={selection.mode === "new_card"}
                name="paywall-payment-method"
                onChange={() => {
                  onPaymentSelectionChange({
                    mode: "new_card",
                  });
                }}
                type="radio"
              />
              <span className="font-semibold">新しい card を使う</span>
            </label>
          ) : null}

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
                <span>利用規約とポリシーに同意し、main purchase へ進む</span>
              </label>
            ) : null}
          </div>

          {showNewCardWidget ? (
            <div className="mt-4">
              <div className="rounded-[20px] border border-[#d7eef7] bg-white/78 px-4 py-3 text-xs leading-6 text-muted">
                widget 内の <span className="font-semibold text-foreground">Place your order</span> を押すと
                payment token を受け取り、そのまま purchase を続行します。
              </div>
              {!consentSatisfied ? (
                <div className="mt-3 rounded-[20px] border border-[#d7eef7] bg-white/78 px-4 py-4 text-sm text-muted">
                  年齢確認と利用規約への同意を完了すると card widget を表示します。
                </div>
              ) : null}
              {isLoadingCardSetupSession ? (
                <div className="mt-3 rounded-[20px] border border-[#d7eef7] bg-white/78 px-4 py-4 text-sm text-muted">
                  card setup session を準備しています。
                </div>
              ) : null}
              {!isLoadingCardSetupSession && canRenderNewCardWidget && cardSetupSession ? (
                <CCBillPaymentWidget
                  className="mt-3"
                  onPaymentTokenCreated={onCardPaymentTokenCreated}
                  session={cardSetupSession}
                />
              ) : null}
            </div>
          ) : null}

          {cardSetupErrorMessage ? (
            <p
              className="mt-3 rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
              role="alert"
            >
              {cardSetupErrorMessage}
            </p>
          ) : null}

          {purchaseErrorMessage ? (
            <p
              className="mt-3 rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
              role="alert"
            >
              {purchaseErrorMessage}
            </p>
          ) : null}

          {unlock.purchase.pendingReason ? (
            <p className="mt-3 rounded-[18px] border border-[#c6e9f7] bg-[#f4fcff] px-4 py-3 text-sm leading-6 text-[#23536e]">
              pending reason: {unlock.purchase.pendingReason}
            </p>
          ) : null}

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
            {buttonLabel ? (
              <button
                className={cn(
                  "flex min-h-[46px] flex-1 items-center justify-center rounded-full px-4 text-[13px] font-bold text-white",
                  confirmEnabled ? "bg-accent-strong" : "cursor-not-allowed bg-accent-strong opacity-40",
                )}
                disabled={!confirmEnabled}
                onClick={onConfirm}
                type="button"
              >
                {buttonLabel}
              </button>
            ) : null}
          </div>

          {!buttonLabel ? (
            <p className="mt-3 text-center text-[11px] leading-5 text-muted">
              {selection.mode === "new_card" ? getUnlockButtonLabel(unlock) : "現在は purchase action を表示していません。"}
            </p>
          ) : null}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

"use client";

import { useState } from "react";

import { ApiError } from "@/shared/api";

import { requestCardSetupSession } from "../api/request-card-setup-session";
import { requestCardSetupToken } from "../api/request-card-setup-token";
import { requestMainAccessEntry } from "../api/request-main-access-entry";
import { requestMainPurchase } from "../api/request-main-purchase";
import { requestUnlockSurfaceByShortId } from "../api/request-unlock-surface";
import { getUnlockEntryAction, type UnlockSurfaceModel } from "./unlock-entry";

type PaymentSelection =
  | {
      mode: "new_card";
      paymentMethodId?: undefined;
    }
  | {
      mode: "saved_card";
      paymentMethodId: string;
    };

type CardSetupSessionModel = Awaited<ReturnType<typeof requestCardSetupSession>>;
type AccessEntryContextModel = Pick<UnlockSurfaceModel["entryContext"], "accessEntryPath" | "token">;
type RestoredViewer = {
  id: string;
} | null | undefined;

type OpenUnlockAuthDialog = (options: {
  onAfterAuthenticated?: ((restoredViewer: RestoredViewer) => Promise<void> | void) | undefined;
}) => void;

type UseUnlockPaywallControllerDeps = {
  requestCardSetupSession: typeof requestCardSetupSession;
  requestCardSetupToken: typeof requestCardSetupToken;
  requestMainAccessEntry: typeof requestMainAccessEntry;
  requestMainPurchase: typeof requestMainPurchase;
  requestUnlockSurfaceByShortId: typeof requestUnlockSurfaceByShortId;
};

type UseUnlockPaywallControllerOptions = {
  baseUnlock: UnlockSurfaceModel;
  deps?: Partial<UseUnlockPaywallControllerDeps>;
  hasViewerSession: boolean;
  onFallbackNavigation: () => void;
  onNavigateToMain: (href: string) => void;
  onOpenAuthDialog: OpenUnlockAuthDialog;
  onOpenReAuthDialog: OpenUnlockAuthDialog;
  shortId: string;
  usePurchaseFlow: boolean;
  viewerIdentityKey: string | null;
};

function buildUnlockStateKey(unlock: UnlockSurfaceModel): string {
  return [
    unlock.short.id,
    unlock.main.id,
    unlock.access.status,
    unlock.access.reason,
    unlock.unlockCta.state,
    unlock.entryContext.accessEntryPath,
    unlock.entryContext.purchasePath,
    unlock.entryContext.token,
    unlock.purchase.state,
    unlock.purchase.pendingReason ?? "no-pending",
    unlock.purchase.savedPaymentMethods.map((method) => method.paymentMethodId).join(","),
    unlock.setup.required ? "setup-required" : "setup-optional",
    unlock.setup.requiresAgeConfirmation ? "age-required" : "age-optional",
    unlock.setup.requiresTermsAcceptance ? "terms-required" : "terms-optional",
  ].join("::");
}

function getDefaultPaywallSelection(unlock: UnlockSurfaceModel): PaymentSelection {
  const firstSavedMethod = unlock.purchase.savedPaymentMethods[0];

  if (firstSavedMethod && !unlock.purchase.setup.requiresCardSetup) {
    return {
      mode: "saved_card",
      paymentMethodId: firstSavedMethod.paymentMethodId,
    };
  }

  return {
    mode: "new_card",
  };
}

function buildCardSetupSessionKey(unlock: UnlockSurfaceModel): string {
  return [unlock.main.id, unlock.short.id, unlock.entryContext.token].join("::");
}

function mapPurchaseFailureMessage(
  failureReason: "authentication_failed" | "card_brand_unsupported" | "purchase_declined" | null,
) {
  switch (failureReason) {
    case "authentication_failed":
      return "カード認証を完了できませんでした。入力内容を確認して再度お試しください。";
    case "card_brand_unsupported":
      return "このカードブランドは利用できません。Visa / Mastercard / JCB / American Express をご利用ください。";
    case "purchase_declined":
      return "購入を完了できませんでした。カード情報をご確認のうえ再度お試しください。";
    default:
      return "購入を完了できませんでした。少し時間を置いてから再度お試しください。";
  }
}

function isMainLockedApiError(error: unknown): boolean {
  return isApiErrorWithCode(error, 403, "main_locked");
}

function isApiErrorWithCode(error: unknown, status: number, code: string): boolean {
  if (!(error instanceof ApiError) || error.status !== status || !error.details) {
    return false;
  }

  try {
    const parsed = JSON.parse(error.details) as {
      error?: {
        code?: string;
      } | null;
    };

    return parsed.error?.code === code;
  } catch {
    return false;
  }
}

function isAuthRequiredApiError(error: unknown): boolean {
  return isApiErrorWithCode(error, 401, "auth_required");
}

function isFreshAuthRequiredApiError(error: unknown): boolean {
  return isApiErrorWithCode(error, 403, "fresh_auth_required");
}

function supportsPurchaseSelection(unlock: UnlockSurfaceModel) {
  return unlock.purchase.state === "purchase_ready" || unlock.purchase.state === "setup_required";
}

function isPaywallSelectionValid(
  unlock: UnlockSurfaceModel,
  selection: PaymentSelection,
) {
  if (!supportsPurchaseSelection(unlock)) {
    return selection.mode === "new_card";
  }

  if (selection.mode === "new_card") {
    return true;
  }

  return unlock.purchase.savedPaymentMethods.some(
    (method) => method.paymentMethodId === selection.paymentMethodId,
  );
}

function buildPendingUnlockSurface(
  unlock: UnlockSurfaceModel,
  access: UnlockSurfaceModel["access"],
): UnlockSurfaceModel {
  return {
    ...unlock,
    access,
    purchase: {
      ...unlock.purchase,
      pendingReason: "provider_processing",
      state: "purchase_pending",
    },
  };
}

/**
 * API-backed paywall の state と purchase/access orchestration を管理する。
 */
export function useUnlockPaywallController({
  baseUnlock,
  deps,
  hasViewerSession,
  onFallbackNavigation,
  onNavigateToMain,
  onOpenAuthDialog,
  onOpenReAuthDialog,
  shortId,
  usePurchaseFlow,
  viewerIdentityKey,
}: UseUnlockPaywallControllerOptions) {
  const [acceptAge, setAcceptAge] = useState(false);
  const [acceptTerms, setAcceptTerms] = useState(false);
  const [isPaywallOpen, setIsPaywallOpen] = useState(false);
  const [isResolvingUnlock, setIsResolvingUnlock] = useState(false);
  const [isSubmittingMainAccess, setIsSubmittingMainAccess] = useState(false);
  const [isSubmittingPurchase, setIsSubmittingPurchase] = useState(false);
  const [isExchangingCardSetupToken, setIsExchangingCardSetupToken] = useState(false);
  const [isLoadingCardSetupSession, setIsLoadingCardSetupSession] = useState(false);
  const [purchaseErrorMessage, setPurchaseErrorMessage] = useState<string | null>(null);
  const [cardSetupErrorMessage, setCardSetupErrorMessage] = useState<string | null>(null);
  const [selectedPaymentSelection, setSelectedPaymentSelection] = useState<PaymentSelection>({
    mode: "new_card",
  });
  const [cardSetupSessionState, setCardSetupSessionState] = useState<{
    key: string;
    session: CardSetupSessionModel;
  } | null>(null);
  const [resolvedUnlockState, setResolvedUnlockState] = useState<{
    baseUnlockKey: string;
    shortId: string;
    viewerIdentityKey: string | null;
    unlock: UnlockSurfaceModel;
  } | null>(null);
  const propUnlockKey = buildUnlockStateKey(baseUnlock);
  const resolvedUnlock =
    resolvedUnlockState?.shortId === shortId &&
    resolvedUnlockState.baseUnlockKey === propUnlockKey &&
    resolvedUnlockState.viewerIdentityKey === viewerIdentityKey
      ? resolvedUnlockState.unlock
      : null;
  const activeUnlock = resolvedUnlock ?? baseUnlock;
  const activeCardSetupSessionKey = buildCardSetupSessionKey(activeUnlock);
  const activeCardSetupSession =
    cardSetupSessionState?.key === activeCardSetupSessionKey
      ? cardSetupSessionState.session
      : null;
  const paymentSelection = isPaywallSelectionValid(activeUnlock, selectedPaymentSelection)
    ? selectedPaymentSelection
    : getDefaultPaywallSelection(activeUnlock);
  const unlockAction = getUnlockEntryAction(activeUnlock);
  const requestCardSetupSessionImpl = deps?.requestCardSetupSession ?? requestCardSetupSession;
  const requestCardSetupTokenImpl = deps?.requestCardSetupToken ?? requestCardSetupToken;
  const requestMainAccessEntryImpl = deps?.requestMainAccessEntry ?? requestMainAccessEntry;
  const requestMainPurchaseImpl = deps?.requestMainPurchase ?? requestMainPurchase;
  const requestUnlockSurfaceByShortIdImpl =
    deps?.requestUnlockSurfaceByShortId ?? requestUnlockSurfaceByShortId;
  const isBusy =
    isResolvingUnlock ||
    isSubmittingMainAccess ||
    isSubmittingPurchase ||
    isExchangingCardSetupToken;

  const closePaywall = () => {
    setIsPaywallOpen(false);
  };

  const clearPaywallMessages = () => {
    setPurchaseErrorMessage(null);
    setCardSetupErrorMessage(null);
  };

  const resetPaywallSelections = () => {
    setAcceptAge(false);
    setAcceptTerms(false);
    setSelectedPaymentSelection({
      mode: "new_card",
    });
    clearPaywallMessages();
  };

  const resetPaywallState = () => {
    resetPaywallSelections();
    closePaywall();
  };

  const storeResolvedUnlock = (
    nextUnlock: UnlockSurfaceModel,
    resolvedViewerIdentityKey: string | null,
  ) => {
    setResolvedUnlockState({
      baseUnlockKey: propUnlockKey,
      shortId,
      viewerIdentityKey: resolvedViewerIdentityKey,
      unlock: nextUnlock,
    });
  };

  const resolveUnlockSurfaceAfterAuth = async ({
    resolvedViewerIdentityKey,
  }: {
    resolvedViewerIdentityKey: string | null;
  }) => {
    if (!usePurchaseFlow) {
      return null;
    }

    const nextUnlock = await requestUnlockSurfaceByShortIdImpl({
      shortId,
    });

    storeResolvedUnlock(nextUnlock, resolvedViewerIdentityKey);

    return nextUnlock;
  };

  const shouldOpenPaywallForUnlock = (targetUnlock: UnlockSurfaceModel) => {
    if (usePurchaseFlow) {
      return (
        targetUnlock.purchase.state === "setup_required" ||
        targetUnlock.purchase.state === "purchase_ready" ||
        targetUnlock.purchase.state === "purchase_pending"
      );
    }

    const targetAction = getUnlockEntryAction(targetUnlock);

    return targetAction === "open_paywall" || targetUnlock.unlockCta.state === "unlock_available";
  };

  const restoreUnlockSurfaceAfterAuth = async ({
    preservePaywallSelections,
    restoredViewer,
    targetUnlock,
  }: {
    preservePaywallSelections: boolean;
    restoredViewer?: RestoredViewer;
    targetUnlock: UnlockSurfaceModel;
  }) => {
    let unlockAfterAuth = targetUnlock;
    const restoredViewerIdentityKey = restoredViewer?.id ?? null;

    if (usePurchaseFlow) {
      const nextUnlock = await resolveUnlockSurfaceAfterAuth({
        resolvedViewerIdentityKey: restoredViewerIdentityKey,
      });
      if (nextUnlock) {
        unlockAfterAuth = nextUnlock;
      }
    }

    const shouldRestorePaywall = shouldOpenPaywallForUnlock(unlockAfterAuth);

    if (!preservePaywallSelections) {
      resetPaywallSelections();
    }

    closePaywall();

    if (shouldRestorePaywall) {
      await openPaywallForUnlock(unlockAfterAuth, {
        preserveSelections: preservePaywallSelections,
      });
    }
  };

  const openReAuthDialog = ({
    preservePaywallSelections = isPaywallOpen,
    targetUnlock = activeUnlock,
  }: {
    preservePaywallSelections?: boolean;
    targetUnlock?: UnlockSurfaceModel;
  } = {}) => {
    onOpenReAuthDialog({
      onAfterAuthenticated:
        usePurchaseFlow || preservePaywallSelections
          ? async (restoredViewer) => {
              await restoreUnlockSurfaceAfterAuth({
                preservePaywallSelections,
                restoredViewer,
                targetUnlock,
              });
            }
          : undefined,
    });
  };

  const openAuthDialogForUnlock = (
    targetUnlock: UnlockSurfaceModel = activeUnlock,
    {
      preservePaywallSelections = isPaywallOpen,
    }: {
      preservePaywallSelections?: boolean;
    } = {},
  ) => {
    const shouldRefreshUnlockAfterAuth = usePurchaseFlow;
    const shouldRestoreUnlockAfterAuth =
      shouldRefreshUnlockAfterAuth || preservePaywallSelections || shouldOpenPaywallForUnlock(targetUnlock);

    onOpenAuthDialog({
      onAfterAuthenticated: shouldRestoreUnlockAfterAuth
        ? async (restoredViewer) => {
            await restoreUnlockSurfaceAfterAuth({
              preservePaywallSelections,
              restoredViewer,
              targetUnlock,
            });
          }
        : undefined,
    });
  };

  const ensureCardSetupSession = async (
    targetUnlock: UnlockSurfaceModel,
  ) => {
    if (!usePurchaseFlow || !supportsPurchaseSelection(targetUnlock)) {
      return;
    }

    const sessionKey = buildCardSetupSessionKey(targetUnlock);

    if (cardSetupSessionState?.key === sessionKey) {
      setCardSetupErrorMessage(null);
      return;
    }

    setIsLoadingCardSetupSession(true);
    setCardSetupErrorMessage(null);

    try {
      const nextSession = await requestCardSetupSessionImpl({
        entryToken: targetUnlock.entryContext.token,
        fromShortId: shortId,
        mainId: targetUnlock.main.id,
      });

      setCardSetupSessionState({
        key: sessionKey,
        session: nextSession,
      });
    } catch (error) {
      if (isAuthRequiredApiError(error)) {
        openAuthDialogForUnlock(targetUnlock, {
          preservePaywallSelections: true,
        });
        return;
      }

      if (isFreshAuthRequiredApiError(error)) {
        openReAuthDialog({
          preservePaywallSelections: true,
          targetUnlock,
        });
        return;
      }

      if (isMainLockedApiError(error) && (await recoverFromLockedPurchaseState(targetUnlock))) {
        return;
      }

      setCardSetupErrorMessage("カード登録フォームの準備に失敗しました。少し時間を置いてから再度お試しください。");
    } finally {
      setIsLoadingCardSetupSession(false);
    }
  };

  const openPaywallForUnlock = async (
    targetUnlock: UnlockSurfaceModel,
    {
      preserveSelections = false,
    }: {
      preserveSelections?: boolean;
    } = {},
  ) => {
    clearPaywallMessages();

    const nextSelection =
      preserveSelections && isPaywallSelectionValid(targetUnlock, selectedPaymentSelection)
        ? selectedPaymentSelection
        : getDefaultPaywallSelection(targetUnlock);

    if (!preserveSelections) {
      setAcceptAge(false);
      setAcceptTerms(false);
    }

    setSelectedPaymentSelection(nextSelection);
    setIsPaywallOpen(true);

    if (usePurchaseFlow && nextSelection.mode === "new_card") {
      await ensureCardSetupSession(targetUnlock);
    }
  };

  const handleClosePaywall = () => {
    if (isSubmittingMainAccess || isSubmittingPurchase || isExchangingCardSetupToken) {
      return;
    }

    resetPaywallState();
  };

  const handleOpenMain = async (
    targetUnlock: UnlockSurfaceModel,
    {
      acceptedAge: acceptedAgeOverride,
      acceptedTerms: acceptedTermsOverride,
      entryContext,
    }: {
      acceptedAge?: boolean | undefined;
      acceptedTerms?: boolean | undefined;
      entryContext?: AccessEntryContextModel | undefined;
    } = {},
  ) => {
    if (!hasViewerSession) {
      openAuthDialogForUnlock(targetUnlock);
      return;
    }

    if (isSubmittingMainAccess) {
      return;
    }

    setIsSubmittingMainAccess(true);

    try {
      if (usePurchaseFlow) {
        const resolvedEntryContext = entryContext ?? targetUnlock.entryContext;
        const response = await requestMainAccessEntryImpl({
          ...(typeof acceptedAgeOverride === "boolean" ? { acceptedAge: acceptedAgeOverride } : {}),
          ...(typeof acceptedTermsOverride === "boolean" ? { acceptedTerms: acceptedTermsOverride } : {}),
          entryToken: resolvedEntryContext.token,
          fromShortId: shortId,
          mainId: targetUnlock.main.id,
          routePath: resolvedEntryContext.accessEntryPath as `/${string}`,
        });

        resetPaywallState();
        onNavigateToMain(response.href);
        return;
      }

      const response = await requestMainAccessEntryImpl({
        ...(typeof acceptedAgeOverride === "boolean" ? { acceptedAge: acceptedAgeOverride } : {}),
        ...(typeof acceptedTermsOverride === "boolean" ? { acceptedTerms: acceptedTermsOverride } : {}),
        entryToken: targetUnlock.mainAccessEntry.token,
        fromShortId: shortId,
        mainId: targetUnlock.main.id,
        routePath: targetUnlock.mainAccessEntry.routePath as `/${string}`,
      });

      resetPaywallState();
      onNavigateToMain(response.href);
      return;
    } catch (error) {
      if (isAuthRequiredApiError(error)) {
        openAuthDialogForUnlock(targetUnlock, {
          preservePaywallSelections: isPaywallOpen,
        });
        return;
      }

      if (isFreshAuthRequiredApiError(error)) {
        openReAuthDialog({
          preservePaywallSelections: isPaywallOpen,
          targetUnlock,
        });
        return;
      }

      onFallbackNavigation();
    } finally {
      setIsSubmittingMainAccess(false);
    }
  };

  const refreshUnlockSurfaceAfterPurchase = async (
    fallbackUnlock: UnlockSurfaceModel,
  ) => {
    try {
      const nextUnlock = await requestUnlockSurfaceByShortIdImpl({
        shortId,
      });

      storeResolvedUnlock(nextUnlock, viewerIdentityKey);

      return nextUnlock;
    } catch {
      storeResolvedUnlock(fallbackUnlock, viewerIdentityKey);
      return fallbackUnlock;
    }
  };

  const recoverFromLockedPurchaseState = async (
    targetUnlock: UnlockSurfaceModel,
    {
      preserveSelections = true,
    }: {
      preserveSelections?: boolean;
    } = {},
  ) => {
    if (!usePurchaseFlow) {
      return false;
    }

    let nextUnlock: UnlockSurfaceModel;

    try {
      nextUnlock = await requestUnlockSurfaceByShortIdImpl({
        shortId,
      });
      storeResolvedUnlock(nextUnlock, viewerIdentityKey);
    } catch (error) {
      if (isAuthRequiredApiError(error)) {
        openAuthDialogForUnlock(targetUnlock, {
          preservePaywallSelections: preserveSelections,
        });
        return true;
      }

      if (isFreshAuthRequiredApiError(error)) {
        openReAuthDialog({
          preservePaywallSelections: preserveSelections,
          targetUnlock,
        });
        return true;
      }

      return false;
    }

    if (shouldOpenPaywallForUnlock(nextUnlock)) {
      await openPaywallForUnlock(nextUnlock, {
        preserveSelections,
      });
      return true;
    }

    if (getUnlockEntryAction(nextUnlock) === "open_main") {
      await handleOpenMain(nextUnlock);
      return true;
    }

    return false;
  };

  const handlePurchase = async (
    targetUnlock: UnlockSurfaceModel,
    paymentMethod: Parameters<typeof requestMainPurchase>[0]["paymentMethod"],
  ) => {
    if (!hasViewerSession) {
      openAuthDialogForUnlock(targetUnlock, {
        preservePaywallSelections: true,
      });
      return;
    }

    if (isSubmittingPurchase || isSubmittingMainAccess) {
      return;
    }

    clearPaywallMessages();
    setIsSubmittingPurchase(true);

    try {
      const result = await requestMainPurchaseImpl({
        acceptedAge: acceptAge,
        acceptedTerms: acceptTerms,
        entryToken: targetUnlock.entryContext.token,
        fromShortId: shortId,
        mainId: targetUnlock.main.id,
        paymentMethod,
        purchasePath: targetUnlock.entryContext.purchasePath as `/${string}`,
      });

      switch (result.purchase.status) {
        case "succeeded":
        case "already_purchased":
        case "owner_preview":
          if (result.entryContext) {
            await handleOpenMain(targetUnlock, {
              entryContext: {
                accessEntryPath: result.entryContext.accessEntryPath,
                token: result.entryContext.token,
              },
            });
            return;
          }

          setPurchaseErrorMessage("購入後の再開導線を復元できませんでした。もう一度お試しください。");
          return;
        case "pending": {
          const pendingUnlock = buildPendingUnlockSurface(targetUnlock, result.access);
          await refreshUnlockSurfaceAfterPurchase(pendingUnlock);
          return;
        }
        case "failed":
          setPurchaseErrorMessage(mapPurchaseFailureMessage(result.purchase.failureReason));
          return;
      }
    } catch (error) {
      if (isAuthRequiredApiError(error)) {
        openAuthDialogForUnlock(targetUnlock, {
          preservePaywallSelections: true,
        });
        return;
      }

      if (isFreshAuthRequiredApiError(error)) {
        openReAuthDialog({
          preservePaywallSelections: true,
          targetUnlock,
        });
        return;
      }

      if (isMainLockedApiError(error) && (await recoverFromLockedPurchaseState(targetUnlock))) {
        return;
      }

      setPurchaseErrorMessage(mapPurchaseFailureMessage(null));
    } finally {
      setIsSubmittingPurchase(false);
    }
  };

  const handleCardPaymentTokenCreated = async (paymentTokenId: string) => {
    if (!usePurchaseFlow || !supportsPurchaseSelection(activeUnlock)) {
      return;
    }

    if (isSubmittingPurchase || isSubmittingMainAccess || isExchangingCardSetupToken) {
      return;
    }

    if (activeCardSetupSession === null) {
      setPurchaseErrorMessage("カード登録セッションを再開できませんでした。もう一度お試しください。");
      return;
    }

    if (
      (activeUnlock.setup.requiresAgeConfirmation && !acceptAge) ||
      (activeUnlock.setup.requiresTermsAcceptance && !acceptTerms)
    ) {
      setPurchaseErrorMessage("年齢確認と利用規約への同意を完了してから purchase を続けてください。");
      return;
    }

    clearPaywallMessages();
    setIsExchangingCardSetupToken(true);

    try {
      const tokenResponse = await requestCardSetupTokenImpl({
        cardSetupSessionToken: activeCardSetupSession.sessionToken,
        entryToken: activeUnlock.entryContext.token,
        fromShortId: shortId,
        mainId: activeUnlock.main.id,
        paymentTokenId,
      });

      await handlePurchase(activeUnlock, {
        cardSetupToken: tokenResponse.cardSetupToken,
        mode: "new_card",
      });
    } catch (error) {
      if (isAuthRequiredApiError(error)) {
        openAuthDialogForUnlock(activeUnlock, {
          preservePaywallSelections: true,
        });
        return;
      }

      if (isFreshAuthRequiredApiError(error)) {
        openReAuthDialog({
          preservePaywallSelections: true,
          targetUnlock: activeUnlock,
        });
        return;
      }

      if (isMainLockedApiError(error) && (await recoverFromLockedPurchaseState(activeUnlock))) {
        return;
      }

      setCardSetupErrorMessage("カード登録情報を確認できませんでした。少し時間を置いてから再度お試しください。");
    } finally {
      setIsExchangingCardSetupToken(false);
    }
  };

  const handleActivateUnlock = async () => {
    if (!hasViewerSession) {
      openAuthDialogForUnlock(activeUnlock);
      return;
    }

    if (isBusy) {
      return;
    }

    let targetUnlock = activeUnlock;

    if (usePurchaseFlow && resolvedUnlock === null) {
      setIsResolvingUnlock(true);

      try {
        targetUnlock = await requestUnlockSurfaceByShortIdImpl({
          shortId,
        });
        storeResolvedUnlock(targetUnlock, viewerIdentityKey);
      } catch (error) {
        if (isAuthRequiredApiError(error)) {
          openAuthDialogForUnlock(activeUnlock);
          return;
        }

        if (isFreshAuthRequiredApiError(error)) {
          openReAuthDialog({
            targetUnlock: activeUnlock,
          });
          return;
        }

        onFallbackNavigation();
        return;
      } finally {
        setIsResolvingUnlock(false);
      }
    }

    if (shouldOpenPaywallForUnlock(targetUnlock)) {
      await openPaywallForUnlock(targetUnlock);
      return;
    }

    if (getUnlockEntryAction(targetUnlock) === "open_main") {
      await handleOpenMain(targetUnlock);
    }
  };

  const handlePaywallConfirm = async () => {
    if (!usePurchaseFlow) {
      await handleOpenMain(activeUnlock, {
        acceptedAge: acceptAge,
        acceptedTerms: acceptTerms,
      });
      return;
    }

    switch (activeUnlock.purchase.state) {
      case "already_purchased":
      case "owner_preview":
        await handleOpenMain(activeUnlock);
        return;
      case "purchase_ready":
      case "setup_required":
        if (paymentSelection.mode === "saved_card") {
          await handlePurchase(activeUnlock, {
            mode: "saved_card",
            paymentMethodId: paymentSelection.paymentMethodId,
          });
        }
        return;
      default:
        return;
    }
  };

  const handlePaymentSelectionChange = (nextSelection: PaymentSelection) => {
    setSelectedPaymentSelection(nextSelection);
    clearPaywallMessages();

    if (nextSelection.mode === "new_card") {
      void ensureCardSetupSession(activeUnlock);
    }
  };

  return {
    acceptAge,
    acceptTerms,
    activeCardSetupSession,
    activeUnlock,
    cardSetupErrorMessage,
    handleActivateUnlock,
    handleCardPaymentTokenCreated,
    handleClosePaywall,
    handlePaywallConfirm,
    handlePaymentSelectionChange,
    isBusy,
    isLoadingCardSetupSession,
    isPaywallOpen,
    isSubmitting: isSubmittingMainAccess || isSubmittingPurchase || isExchangingCardSetupToken,
    paymentSelection,
    purchaseErrorMessage,
    setAcceptAge,
    setAcceptTerms,
    unlockAction,
  };
}

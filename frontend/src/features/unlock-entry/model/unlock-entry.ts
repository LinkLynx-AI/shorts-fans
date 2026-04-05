import type { CreatorSummary } from "@/entities/creator";
import type { MainSummary } from "@/entities/main";
import type { ShortSummary } from "@/entities/short";

import type { UnlockCtaState } from "./unlock-cta";

export type PurchaseStatus = "not_purchased" | "purchased";
export type MainAccessStatus = "locked" | "owner" | "purchased";
export type MainAccessReason = "owner_preview" | "purchase_required" | "purchased_access";

export type PurchaseState = {
  mainId: string;
  status: PurchaseStatus;
};

export type MainAccessState = {
  mainId: string;
  reason: MainAccessReason;
  status: MainAccessStatus;
};

export type UnlockSetupState = {
  required: boolean;
  requiresAgeConfirmation: boolean;
  requiresTermsAcceptance: boolean;
};

export type UnlockSurfaceModel = {
  access: MainAccessState;
  creator: CreatorSummary;
  main: MainSummary;
  purchase: PurchaseState;
  setup: UnlockSetupState;
  short: ShortSummary;
  unlockCta: UnlockCtaState;
};

export type UnlockEntryAction = "open_main" | "open_paywall" | "unavailable";

/**
 * unlock state から次の UI action を決定する。
 */
export function getUnlockEntryAction(unlock: Pick<UnlockSurfaceModel, "unlockCta">): UnlockEntryAction {
  switch (unlock.unlockCta.state) {
    case "continue_main":
    case "owner_preview":
    case "unlock_available":
      return "open_main";
    case "setup_required":
      return "open_paywall";
    case "unavailable":
      return "unavailable";
  }
}

/**
 * main playback route の href を組み立てる。
 */
export function getMainPlaybackHref(mainId: string, fromShortId?: string): string {
  if (!fromShortId) {
    return `/mains/${mainId}`;
  }

  const searchParams = new URLSearchParams({
    fromShortId,
  });

  return `/mains/${mainId}?${searchParams.toString()}`;
}

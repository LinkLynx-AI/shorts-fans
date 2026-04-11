import type { CreatorSummary } from "@/entities/creator";
import type { ShortMediaAsset } from "@/entities/short";

import type { UnlockCtaState } from "./unlock-cta";

export type MainAccessStatus = "locked" | "owner" | "unlocked";
export type MainAccessReason = "owner_preview" | "session_unlocked" | "unlock_required";

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

export type MainAccessEntry = {
  routePath: string;
  token: string;
};

export type UnlockMainSummary = {
  durationSeconds: number;
  id: string;
  priceJpy: number;
};

export type UnlockShortSummary = {
  caption: string;
  canonicalMainId: string;
  creatorId: string;
  id: string;
  media: ShortMediaAsset;
  previewDurationSeconds: number;
};

export type MainPlaybackGrantKind = "owner" | "unlocked";

export type UnlockSurfaceModel = {
  access: MainAccessState;
  creator: CreatorSummary;
  mainAccessEntry: MainAccessEntry;
  main: UnlockMainSummary;
  setup: UnlockSetupState;
  short: UnlockShortSummary;
  unlockCta: UnlockCtaState;
};

export type UnlockEntryAction = "open_main" | "open_paywall" | "unavailable";

/**
 * main access entry token の検証用 context を組み立てる。
 */
export function buildMockMainAccessEntryContext(mainId: string, fromShortId: string): string {
  return `main-access-entry::${mainId}::${fromShortId}`;
}

/**
 * main playback grant の検証用 context を組み立てる。
 */
export function buildMockMainPlaybackGrantContext(
  mainId: string,
  fromShortId: string,
  grantKind: MainPlaybackGrantKind,
): string {
  return `main-playback-grant::${mainId}::${fromShortId}::${grantKind}`;
}

/**
 * main playback grant context を解析する。
 */
export function parseMockMainPlaybackGrantContext(
  context: string,
): {
  fromShortId: string;
  grantKind: MainPlaybackGrantKind;
  mainId: string;
} | null {
  const [prefix, mainId, fromShortId, grantKind] = context.split("::");

  if (
    prefix !== "main-playback-grant" ||
    !mainId ||
    !fromShortId ||
    (grantKind !== "owner" && grantKind !== "unlocked")
  ) {
    return null;
  }

  return {
    fromShortId,
    grantKind,
    mainId,
  };
}

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
export function getMainPlaybackHref(mainId: string, fromShortId: string, grantToken: string): string {
  const searchParams = new URLSearchParams({
    fromShortId,
    grant: grantToken,
  });

  return `/mains/${mainId}?${searchParams.toString()}`;
}

/**
 * server access-entry endpoint の href を組み立てる。
 */
export function getMockMainAccessRoutePath(mainId: string): string {
  return `/api/fan/mains/${mainId}/access-entry`;
}

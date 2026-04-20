import { getCreatorById } from "@/entities/creator";
import { getMainById } from "@/entities/main";
import { getShortById } from "@/entities/short";
import { issueMockSignedToken } from "@/shared/lib/mock-signed-token";

import {
  buildMockMainAccessEntryContext,
  getMockMainAccessRoutePath,
  normalizeUnlockSurface,
  type UnlockSurfaceModel,
} from "../model/unlock-entry";
import { unlockSurfaceSchema } from "./contracts";

type RawUnlockState = {
  access: UnlockSurfaceModel["access"];
  purchase: UnlockSurfaceModel["purchase"];
  unlockCta: UnlockSurfaceModel["unlockCta"];
};

const rawUnlockStateByShortId: Readonly<Record<string, RawUnlockState>> = {
  afterrain: {
    access: {
      mainId: "main_sora_after_rain",
      reason: "unlock_required",
      status: "locked",
    },
    purchase: {
      pendingReason: null,
      savedPaymentMethods: [
        {
          brand: "visa",
          last4: "4242",
          paymentMethodId: "paymeth_afterrain_visa",
        },
      ],
      setup: {
        required: false,
        requiresAgeConfirmation: false,
        requiresCardSetup: false,
        requiresTermsAcceptance: false,
      },
      state: "purchase_ready",
      supportedCardBrands: ["visa", "mastercard", "jcb", "american_express"],
    },
    unlockCta: {
      mainDurationSeconds: 540,
      priceJpy: 2100,
      resumePositionSeconds: null,
      state: "unlock_available",
    },
  },
  balcony: {
    access: {
      mainId: "main_aoi_blue_balcony",
      reason: "owner_preview",
      status: "owner",
    },
    purchase: {
      pendingReason: null,
      savedPaymentMethods: [],
      setup: {
        required: false,
        requiresAgeConfirmation: false,
        requiresCardSetup: false,
        requiresTermsAcceptance: false,
      },
      state: "owner_preview",
      supportedCardBrands: ["visa", "mastercard", "jcb", "american_express"],
    },
    unlockCta: {
      mainDurationSeconds: null,
      priceJpy: null,
      resumePositionSeconds: null,
      state: "owner_preview",
    },
  },
  mirror: {
    access: {
      mainId: "main_mina_hotel_mirror",
      reason: "unlock_required",
      status: "locked",
    },
    purchase: {
      pendingReason: null,
      savedPaymentMethods: [],
      setup: {
        required: true,
        requiresAgeConfirmation: true,
        requiresCardSetup: true,
        requiresTermsAcceptance: true,
      },
      state: "setup_required",
      supportedCardBrands: ["visa", "mastercard", "jcb", "american_express"],
    },
    unlockCta: {
      mainDurationSeconds: 660,
      priceJpy: 2400,
      resumePositionSeconds: null,
      state: "setup_required",
    },
  },
  poolcut: {
    access: {
      mainId: "main_sora_poolside_cut",
      reason: "unlock_required",
      status: "locked",
    },
    purchase: {
      pendingReason: null,
      savedPaymentMethods: [
        {
          brand: "mastercard",
          last4: "5555",
          paymentMethodId: "paymeth_poolcut_mastercard",
        },
      ],
      setup: {
        required: false,
        requiresAgeConfirmation: false,
        requiresCardSetup: false,
        requiresTermsAcceptance: false,
      },
      state: "purchase_ready",
      supportedCardBrands: ["visa", "mastercard", "jcb", "american_express"],
    },
    unlockCta: {
      mainDurationSeconds: 480,
      priceJpy: 1900,
      resumePositionSeconds: null,
      state: "unlock_available",
    },
  },
  rooftop: {
    access: {
      mainId: "main_mina_quiet_rooftop",
      reason: "unlock_required",
      status: "locked",
    },
    purchase: {
      pendingReason: null,
      savedPaymentMethods: [],
      setup: {
        required: true,
        requiresAgeConfirmation: true,
        requiresCardSetup: true,
        requiresTermsAcceptance: true,
      },
      state: "setup_required",
      supportedCardBrands: ["visa", "mastercard", "jcb", "american_express"],
    },
    unlockCta: {
      mainDurationSeconds: 480,
      priceJpy: 1800,
      resumePositionSeconds: null,
      state: "setup_required",
    },
  },
  softlight: {
    access: {
      mainId: "main_aoi_blue_balcony",
      reason: "purchased",
      status: "unlocked",
    },
    purchase: {
      pendingReason: null,
      savedPaymentMethods: [],
      setup: {
        required: false,
        requiresAgeConfirmation: false,
        requiresCardSetup: false,
        requiresTermsAcceptance: false,
      },
      state: "already_purchased",
      supportedCardBrands: ["visa", "mastercard", "jcb", "american_express"],
    },
    unlockCta: {
      mainDurationSeconds: null,
      priceJpy: null,
      resumePositionSeconds: 198,
      state: "continue_main",
    },
  },
};

/**
 * short ごとの unlock surface 用 mock を取得する。
 */
export function getUnlockSurfaceByShortId(shortId: string): UnlockSurfaceModel | undefined {
  const short = getShortById(shortId);

  if (!short) {
    return undefined;
  }

  const creator = getCreatorById(short.creatorId);

  if (!creator) {
    throw new Error(`Unknown creator for unlock surface: ${short.creatorId}`);
  }

  const main = getMainById(short.canonicalMainId);

  if (!main) {
    throw new Error(`Unknown main for unlock surface: ${short.canonicalMainId}`);
  }

  const rawState = rawUnlockStateByShortId[short.id];

  if (!rawState) {
    throw new Error(`Unknown unlock state for short: ${short.id}`);
  }

  return normalizeUnlockSurface(unlockSurfaceSchema.parse({
    access: rawState.access,
    creator,
    entryContext: {
      accessEntryPath: getMockMainAccessRoutePath(main.id),
      purchasePath: `/api/fan/mains/${main.id}/purchase`,
      token: issueMockSignedToken(buildMockMainAccessEntryContext(main.id, shortId)),
    },
    main: {
      durationSeconds: main.durationSeconds,
      id: main.id,
      priceJpy: main.priceJpy,
    },
    purchase: rawState.purchase,
    short: {
      ...short,
      id: shortId,
    },
    unlockCta: rawState.unlockCta,
  }));
}

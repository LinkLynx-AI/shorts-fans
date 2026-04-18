import { getFanHubState, type FanLibraryItem } from "@/entities/fan-profile";
import {
  normalizeUnlockSurface,
  requestMainAccessEntry,
  requestUnlockSurfaceByShortId,
} from "@/features/unlock-entry";

import { requestMainPlaybackSurface } from "./request-main-playback-surface";
import { resolveLibraryMainPlaybackSurface } from "./resolve-library-main-playback-surface";

vi.mock("@/features/unlock-entry", async () => {
  const actual = await vi.importActual<typeof import("@/features/unlock-entry")>("@/features/unlock-entry");

  return {
    ...actual,
    requestMainAccessEntry: vi.fn(),
    requestUnlockSurfaceByShortId: vi.fn(),
  };
});

vi.mock("./request-main-playback-surface", () => ({
  requestMainPlaybackSurface: vi.fn(),
}));

const mockedRequestMainAccessEntry = vi.mocked(requestMainAccessEntry);
const mockedRequestMainPlaybackSurface = vi.mocked(requestMainPlaybackSurface);
const mockedRequestUnlockSurfaceByShortId = vi.mocked(requestUnlockSurfaceByShortId);

function createApiLibraryItem(overrides: Partial<FanLibraryItem> = {}): FanLibraryItem {
  return {
    access: {
      mainId: "main_11111111111111111111111111111111",
      reason: "purchased",
      status: "unlocked",
    },
    creator: {
      avatar: null,
      bio: "quiet rooftop と hotel light の preview を軸に投稿。",
      displayName: "Mina Rei",
      handle: "@minarei",
      id: "creator_22222222222222222222222222222222",
    },
    entryShort: {
      caption: "quiet rooftop preview",
      canonicalMainId: "main_11111111111111111111111111111111",
      creatorId: "creator_22222222222222222222222222222222",
      id: "short_33333333333333333333333333333333",
      media: {
        durationSeconds: 16,
        id: "asset_44444444444444444444444444444444",
        kind: "video",
        posterUrl: "https://cdn.example.com/shorts/poster.jpg",
        url: "https://cdn.example.com/shorts/playback.mp4",
      },
      previewDurationSeconds: 16,
    },
    main: {
      durationSeconds: 480,
      id: "main_11111111111111111111111111111111",
    },
    ...overrides,
  };
}

describe("resolveLibraryMainPlaybackSurface", () => {
  beforeEach(() => {
    mockedRequestMainAccessEntry.mockReset();
    mockedRequestMainPlaybackSurface.mockReset();
    mockedRequestUnlockSurfaceByShortId.mockReset();
  });

  it("builds a local playback surface for legacy library items", async () => {
    const item = getFanHubState("library").libraryItems[1];

    if (!item) {
      throw new Error("library item fixture missing");
    }

    await expect(resolveLibraryMainPlaybackSurface(item)).resolves.toMatchObject({
      kind: "ready",
      surface: {
        access: {
          mainId: item.main.id,
          status: "unlocked",
        },
        entryShort: {
          id: item.entryShort.id,
        },
        main: {
          id: item.main.id,
        },
      },
    });
  });

  it("uses unlock -> access-entry -> playback sequencing for api-backed items", async () => {
    const item = createApiLibraryItem();
    const expectedSurface = {
      access: item.access,
      creator: item.creator,
      entryShort: item.entryShort,
      main: {
        durationSeconds: item.main.durationSeconds,
        id: item.main.id,
        media: {
          durationSeconds: item.main.durationSeconds,
          id: "asset_main_55555555555555555555555555555555",
          kind: "video",
          posterUrl: "https://cdn.example.com/mains/poster.jpg",
          url: "https://cdn.example.com/mains/playback.mp4",
        },
      },
      resumePositionSeconds: 120,
      themeShort: item.entryShort,
      viewer: {
        isPinned: null,
      },
    } satisfies Awaited<ReturnType<typeof requestMainPlaybackSurface>>;

    mockedRequestUnlockSurfaceByShortId.mockResolvedValue(normalizeUnlockSurface({
      access: {
        mainId: item.main.id,
        reason: "purchased",
        status: "unlocked",
      },
      creator: item.creator,
      entryContext: {
        accessEntryPath: `/api/fan/mains/${item.main.id}/access-entry`,
        purchasePath: `/api/fan/mains/${item.main.id}/purchase`,
        token: "entry-token",
      },
      main: {
        id: item.main.id,
        durationSeconds: item.main.durationSeconds,
        priceJpy: 1800,
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
      short: {
        ...item.entryShort,
        id: item.entryShort.id,
      },
      unlockCta: {
        state: "continue_main",
        mainDurationSeconds: null,
        priceJpy: null,
        resumePositionSeconds: 120,
      },
    }));
    mockedRequestMainAccessEntry.mockResolvedValue({
      href: `/mains/${item.main.id}?fromShortId=${item.entryShort.id}&grant=grant-token`,
    });
    mockedRequestMainPlaybackSurface.mockResolvedValue(expectedSurface);

    await expect(resolveLibraryMainPlaybackSurface(item)).resolves.toEqual({
      kind: "ready",
      surface: expectedSurface,
    });
    expect(mockedRequestUnlockSurfaceByShortId).toHaveBeenCalledWith({
      shortId: item.entryShort.id,
    });
    expect(mockedRequestMainAccessEntry).toHaveBeenCalledWith({
      entryToken: "entry-token",
      fromShortId: item.entryShort.id,
      mainId: item.main.id,
      routePath: `/api/fan/mains/${item.main.id}/access-entry`,
    });
    expect(mockedRequestMainPlaybackSurface).toHaveBeenCalledWith({
      fromShortId: item.entryShort.id,
      grant: "grant-token",
      mainId: item.main.id,
    });
  });

  it("returns a locked result when the api-backed item still requires setup", async () => {
    const item = createApiLibraryItem();

    mockedRequestUnlockSurfaceByShortId.mockResolvedValue(normalizeUnlockSurface({
      access: {
        mainId: item.main.id,
        reason: "unlock_required",
        status: "locked",
      },
      creator: item.creator,
      entryContext: {
        accessEntryPath: `/api/fan/mains/${item.main.id}/access-entry`,
        purchasePath: `/api/fan/mains/${item.main.id}/purchase`,
        token: "entry-token",
      },
      main: {
        id: item.main.id,
        durationSeconds: item.main.durationSeconds,
        priceJpy: 1800,
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
      short: {
        ...item.entryShort,
        id: item.entryShort.id,
      },
      unlockCta: {
        state: "setup_required",
        mainDurationSeconds: item.main.durationSeconds,
        priceJpy: 1800,
        resumePositionSeconds: null,
      },
    }));

    await expect(resolveLibraryMainPlaybackSurface(item)).resolves.toEqual({
      kind: "locked",
    });
    expect(mockedRequestMainAccessEntry).not.toHaveBeenCalled();
    expect(mockedRequestMainPlaybackSurface).not.toHaveBeenCalled();
  });
});

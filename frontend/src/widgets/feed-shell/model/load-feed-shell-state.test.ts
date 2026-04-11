import { ApiError } from "@/shared/api";
import { getFanFeedPage } from "@/entities/short";
import { isAuthRequiredApiError } from "@/features/fan-auth";
import { buildFeedSurfaceFromApiItem } from "@/widgets/immersive-short-surface";
import { loadFeedShellState } from "@/widgets/feed-shell";

vi.mock("@/entities/short", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/entities/short")>();

  return {
    ...actual,
    getFanFeedPage: vi.fn(),
  };
});

vi.mock("@/features/fan-auth", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/features/fan-auth")>();

  return {
    ...actual,
    isAuthRequiredApiError: vi.fn(),
  };
});

vi.mock("@/widgets/immersive-short-surface", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/widgets/immersive-short-surface")>();

  return {
    ...actual,
    buildFeedSurfaceFromApiItem: vi.fn(),
  };
});

const mockedGetFanFeedPage = vi.mocked(getFanFeedPage);
const mockedIsAuthRequiredApiError = vi.mocked(isAuthRequiredApiError);
const mockedBuildFeedSurfaceFromApiItem = vi.mocked(buildFeedSurfaceFromApiItem);

describe("loadFeedShellState", () => {
  beforeEach(() => {
    mockedGetFanFeedPage.mockReset();
    mockedIsAuthRequiredApiError.mockReset();
    mockedBuildFeedSurfaceFromApiItem.mockReset();
  });

  it("builds the ready state from API items", async () => {
    const apiItem = {
      creator: {
        avatar: null,
        bio: "night preview specialist",
        displayName: "Mina Rei",
        handle: "@minarei" as const,
        id: "creator_mina_rei",
      },
      short: {
        caption: "quiet rooftop preview",
        canonicalMainId: "main_33333333333333333333333333333333",
        creatorId: "creator_mina_rei",
        id: "short_22222222222222222222222222222222",
        media: {
          durationSeconds: 16,
          id: "asset_short_mina_rooftop",
          kind: "video" as const,
          posterUrl: "https://cdn.example.com/shorts/poster.jpg",
          url: "https://cdn.example.com/shorts/playback.mp4",
        },
        previewDurationSeconds: 16,
      },
      unlockCta: {
        mainDurationSeconds: 480,
        priceJpy: 1800,
        resumePositionSeconds: null,
        state: "unlock_available" as const,
      },
      viewer: {
        isFollowingCreator: true,
        isPinned: true,
      },
    };
    const surface = {
      creator: apiItem.creator,
      mainEntryEnabled: false,
      short: {
        ...apiItem.short,
        title: apiItem.short.caption,
      },
      unlock: {
        access: {
          mainId: apiItem.short.canonicalMainId,
          reason: "unlock_required" as const,
          status: "locked" as const,
        },
        creator: apiItem.creator,
        main: {
          durationSeconds: 480,
          id: apiItem.short.canonicalMainId,
          priceJpy: 1800,
          title: apiItem.short.caption,
        },
        mainAccessEntry: {
          routePath: `/api/fan/mains/${apiItem.short.canonicalMainId}/access-entry`,
          token: `disabled-${apiItem.short.id}`,
        },
        setup: {
          required: false,
          requiresAgeConfirmation: false,
          requiresTermsAcceptance: false,
        },
        short: {
          ...apiItem.short,
          title: apiItem.short.caption,
        },
        unlockCta: apiItem.unlockCta,
      },
      viewer: {
        isFollowingCreator: true,
        isPinned: true,
      },
    };

    mockedGetFanFeedPage.mockResolvedValue({
      items: [apiItem],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_feed_001",
      tab: "recommended",
    });
    mockedBuildFeedSurfaceFromApiItem.mockReturnValue(surface);

    await expect(
      loadFeedShellState("recommended", {
        sessionToken: "raw-session-token",
      }),
    ).resolves.toEqual({
      kind: "ready",
      surfaces: [surface],
      tab: "recommended",
    });

    expect(mockedGetFanFeedPage).toHaveBeenCalledWith({
      sessionToken: "raw-session-token",
      tab: "recommended",
    });
    expect(mockedBuildFeedSurfaceFromApiItem).toHaveBeenCalledTimes(1);
    expect(mockedBuildFeedSurfaceFromApiItem.mock.calls[0]?.[0]).toEqual(apiItem);
  });

  it("returns the empty state when the feed has no items", async () => {
    mockedGetFanFeedPage.mockResolvedValue({
      items: [],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_feed_empty_001",
      tab: "following",
    });

    await expect(loadFeedShellState("following")).resolves.toEqual({
      kind: "empty",
      tab: "following",
    });
  });

  it("returns auth_required for following when the API rejects with auth_required", async () => {
    const error = new ApiError("auth required", {
      code: "http",
      status: 401,
    });

    mockedGetFanFeedPage.mockRejectedValue(error);
    mockedIsAuthRequiredApiError.mockReturnValue(true);

    await expect(loadFeedShellState("following")).resolves.toEqual({
      kind: "auth_required",
      tab: "following",
    });
  });

  it("rethrows non auth_required errors", async () => {
    const error = new Error("boom");

    mockedGetFanFeedPage.mockRejectedValue(error);
    mockedIsAuthRequiredApiError.mockReturnValue(false);

    await expect(loadFeedShellState("recommended")).rejects.toThrow("boom");
  });
});

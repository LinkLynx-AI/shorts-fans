import {
  buildDetailSurfaceFromApi,
  buildFeedSurfaceFromApiItem,
} from "@/widgets/immersive-short-surface";

describe("api short surface builders", () => {
  it("maps a feed item to an immersive feed surface", () => {
    const surface = buildFeedSurfaceFromApiItem({
      creator: {
        avatar: null,
        bio: "night preview specialist",
        displayName: "Mina Rei",
        handle: "@minarei",
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
          kind: "video",
          posterUrl: "https://cdn.example.com/shorts/poster.jpg",
          url: "https://cdn.example.com/shorts/playback.mp4",
        },
        previewDurationSeconds: 16,
      },
      unlockCta: {
        mainDurationSeconds: 480,
        priceJpy: 1800,
        resumePositionSeconds: null,
        state: "unlock_available",
      },
      viewer: {
        isFollowingCreator: false,
        isPinned: true,
      },
    });

    expect(surface.mainEntryEnabled).toBe(true);
    expect(surface.short.caption).toBe("quiet rooftop preview");
    expect(surface.unlock.access).toEqual({
      mainId: "main_33333333333333333333333333333333",
      reason: "unlock_required",
      status: "locked",
    });
    expect(surface.unlock.mainAccessEntry).toEqual({
      routePath: "/api/fan/mains/main_33333333333333333333333333333333/access-entry",
      token: "disabled-short_22222222222222222222222222222222",
    });
    expect(surface.viewer).toEqual({
      isFollowingCreator: false,
      isPinned: true,
    });
  });

  it("maps a detail payload to an immersive detail surface", () => {
    const surface = buildDetailSurfaceFromApi({
      creator: {
        avatar: null,
        bio: "night preview specialist",
        displayName: "Mina Rei",
        handle: "@minarei",
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
          kind: "video",
          posterUrl: "https://cdn.example.com/shorts/poster.jpg",
          url: "https://cdn.example.com/shorts/playback.mp4",
        },
        previewDurationSeconds: 16,
      },
      unlockCta: {
        mainDurationSeconds: null,
        priceJpy: null,
        resumePositionSeconds: 42,
        state: "continue_main",
      },
      viewer: {
        isFollowingCreator: true,
        isPinned: false,
      },
    });

    expect(surface.unlock.access).toEqual({
      mainId: "main_33333333333333333333333333333333",
      reason: "session_unlocked",
      status: "unlocked",
    });
    expect(surface.unlock.main.priceJpy).toBe(0);
    expect(surface.unlock.main.durationSeconds).toBe(16);
    expect(surface.mainEntryEnabled).toBe(false);
    expect(surface.viewer).toEqual({
      isFollowingCreator: true,
      isPinned: false,
    });
  });

  it("preserves empty captions for optional-caption shorts", () => {
    const surface = buildFeedSurfaceFromApiItem({
      creator: {
        avatar: null,
        bio: "night preview specialist",
        displayName: "Mina Rei",
        handle: "@minarei",
        id: "creator_mina_rei",
      },
      short: {
        caption: "",
        canonicalMainId: "main_33333333333333333333333333333333",
        creatorId: "creator_mina_rei",
        id: "short_22222222222222222222222222222222",
        media: {
          durationSeconds: 16,
          id: "asset_short_mina_rooftop",
          kind: "video",
          posterUrl: "https://cdn.example.com/shorts/poster.jpg",
          url: "https://cdn.example.com/shorts/playback.mp4",
        },
        previewDurationSeconds: 16,
      },
      unlockCta: {
        mainDurationSeconds: 480,
        priceJpy: 1800,
        resumePositionSeconds: null,
        state: "unlock_available",
      },
      viewer: {
        isFollowingCreator: false,
        isPinned: false,
      },
    });

    expect(surface.short.caption).toBe("");
    expect(surface.unlock.short.caption).toBe("");
  });
});

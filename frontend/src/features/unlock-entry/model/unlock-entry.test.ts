import {
  getMainPlaybackHref,
  getUnlockEntryAction,
  type UnlockSurfaceModel,
} from "@/features/unlock-entry";

function createUnlockSurfaceModel(
  state: UnlockSurfaceModel["unlockCta"]["state"],
): UnlockSurfaceModel {
  return {
    access: {
      mainId: "main_rooftop",
      reason: "purchase_required",
      status: "locked",
    },
    creator: {
      avatar: {
        durationSeconds: null,
        id: "creator_avatar",
        kind: "image",
        posterUrl: null,
        url: "https://cdn.example.com/avatar.jpg",
      },
      bio: "bio",
      displayName: "Mina Rei",
      handle: "@minarei",
      id: "mina",
    },
    main: {
      durationSeconds: 480,
      id: "main_rooftop",
      priceJpy: 1800,
      title: "quiet rooftop main",
    },
    purchase: {
      mainId: "main_rooftop",
      status: "not_purchased",
    },
    setup: {
      required: state === "setup_required",
      requiresAgeConfirmation: state === "setup_required",
      requiresTermsAcceptance: state === "setup_required",
    },
    short: {
      canonicalMainId: "main_rooftop",
      caption: "quiet rooftop preview.",
      creatorId: "mina",
      id: "rooftop",
      media: {
        durationSeconds: 16,
        id: "short_rooftop",
        kind: "video",
        posterUrl: "https://cdn.example.com/short-poster.jpg",
        url: "https://cdn.example.com/short.mp4",
      },
      previewDurationSeconds: 16,
      title: "quiet rooftop preview",
    },
    unlockCta: {
      mainDurationSeconds: state === "continue_main" ? null : 480,
      priceJpy: state === "continue_main" ? null : 1800,
      resumePositionSeconds: state === "continue_main" ? 198 : null,
      state,
    },
  };
}

describe("unlock-entry model", () => {
  it("maps setup-required to paywall", () => {
    expect(getUnlockEntryAction(createUnlockSurfaceModel("setup_required"))).toBe("open_paywall");
  });

  it("maps direct-access states to main playback", () => {
    expect(getUnlockEntryAction(createUnlockSurfaceModel("unlock_available"))).toBe("open_main");
    expect(getUnlockEntryAction(createUnlockSurfaceModel("continue_main"))).toBe("open_main");
    expect(getUnlockEntryAction(createUnlockSurfaceModel("owner_preview"))).toBe("open_main");
  });

  it("builds playback href with an entry short id", () => {
    expect(getMainPlaybackHref("main_mina_quiet_rooftop", "rooftop")).toBe(
      "/mains/main_mina_quiet_rooftop?fromShortId=rooftop",
    );
  });
});

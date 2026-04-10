import {
  buildMockMainAccessEntryContext,
  buildMockMainPlaybackGrantContext,
  getMainPlaybackHref,
  getMockMainAccessRoutePath,
  getUnlockEntryAction,
  parseMockMainPlaybackGrantContext,
  type UnlockSurfaceModel,
} from "@/features/unlock-entry";

function createUnlockSurfaceModel(
  state: UnlockSurfaceModel["unlockCta"]["state"],
): UnlockSurfaceModel {
  return {
    access: {
      mainId: "main_rooftop",
      reason: "unlock_required",
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
    mainAccessEntry: {
      routePath: "/api/fan/mains/main_rooftop/access-entry",
      token: "entry_token",
    },
    main: {
      durationSeconds: 480,
      id: "main_rooftop",
      priceJpy: 1800,
      title: "quiet rooftop main",
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
    expect(getMainPlaybackHref("main_mina_quiet_rooftop", "rooftop", "grant_123")).toBe(
      "/mains/main_mina_quiet_rooftop?fromShortId=rooftop&grant=grant_123",
    );
  });

  it("builds entry and playback grant contexts", () => {
    expect(buildMockMainAccessEntryContext("main_mina_quiet_rooftop", "rooftop")).toBe(
      "main-access-entry::main_mina_quiet_rooftop::rooftop",
    );
    expect(
      buildMockMainPlaybackGrantContext("main_mina_quiet_rooftop", "rooftop", "unlocked"),
    ).toBe("main-playback-grant::main_mina_quiet_rooftop::rooftop::unlocked");
    expect(
      parseMockMainPlaybackGrantContext("main-playback-grant::main_mina_quiet_rooftop::rooftop::owner"),
    ).toEqual({
      fromShortId: "rooftop",
      grantKind: "owner",
      mainId: "main_mina_quiet_rooftop",
    });
    expect(getMockMainAccessRoutePath("main_mina_quiet_rooftop")).toBe(
      "/api/fan/mains/main_mina_quiet_rooftop/access-entry",
    );
    expect(parseMockMainPlaybackGrantContext("main_mina_quiet_rooftop::rooftop")).toBeNull();
  });

  it("parses invalid playback grant context as null", () => {
    expect(
      parseMockMainPlaybackGrantContext(
        "main-playback-grant::main_mina_quiet_rooftop::rooftop::invalid",
      ),
    ).toBeNull();
  });
});

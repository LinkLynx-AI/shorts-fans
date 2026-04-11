import { ApiError } from "@/shared/api";

import { fetchMainPlayback } from "../api/request-main-playback";
import { loadMainPlaybackSurface } from "./load-main-playback-surface";

vi.mock("../api/request-main-playback", () => ({
  fetchMainPlayback: vi.fn(),
}));

describe("loadMainPlaybackSurface", () => {
  it("returns a ready surface when playback payload resolves", async () => {
    vi.mocked(fetchMainPlayback).mockResolvedValue({
      access: {
        mainId: "main_mina_quiet_rooftop",
        reason: "session_unlocked",
        status: "unlocked",
      },
      creator: {
        avatar: {
          durationSeconds: null,
          id: "avatar",
          kind: "image",
          posterUrl: null,
          url: "https://cdn.example.com/avatar.jpg",
        },
        bio: "bio",
        displayName: "Mina Rei",
        handle: "@minarei",
        id: "creator_mina_rei",
      },
      entryShort: {
        canonicalMainId: "main_mina_quiet_rooftop",
        caption: "quiet rooftop preview.",
        creatorId: "mina",
        id: "rooftop",
        media: {
          durationSeconds: 16,
          id: "asset_short_mina_rooftop",
          kind: "video",
          posterUrl: "https://cdn.example.com/shorts/mina-rooftop-poster.jpg",
          url: "https://cdn.example.com/shorts/mina-rooftop.mp4",
        },
        previewDurationSeconds: 16,
        title: "quiet rooftop preview",
      },
      main: {
        durationSeconds: 480,
        id: "main_mina_quiet_rooftop",
        media: {
          durationSeconds: 480,
          id: "asset_main_mina_quiet_rooftop",
          kind: "video",
          posterUrl: "https://cdn.example.com/mains/mina-quiet-rooftop-poster.jpg",
          url: "https://cdn.example.com/mains/mina-quiet-rooftop.mp4",
        },
        priceJpy: 1800,
        title: "quiet rooftop main",
      },
      resumePositionSeconds: null,
    });

    const result = await loadMainPlaybackSurface("main_mina_quiet_rooftop", {
      fromShortId: "rooftop",
      grant: "grant_123",
      sessionToken: "viewer-session",
    });

    expect(result.kind).toBe("ready");

    if (result.kind !== "ready") {
      throw new Error("expected ready result");
    }

    expect(result.surface.themeShort.id).toBe("rooftop");
  });

  it("maps http errors to page-facing states", async () => {
    vi.mocked(fetchMainPlayback).mockRejectedValueOnce(
      new ApiError("auth required", {
        code: "http",
        status: 401,
      }),
    );
    await expect(
      loadMainPlaybackSurface("main_mina_quiet_rooftop", {
        fromShortId: "rooftop",
        grant: "grant_123",
      }),
    ).resolves.toEqual({ kind: "auth_required" });

    vi.mocked(fetchMainPlayback).mockRejectedValueOnce(
      new ApiError("locked", {
        code: "http",
        status: 403,
      }),
    );
    await expect(
      loadMainPlaybackSurface("main_mina_quiet_rooftop", {
        fromShortId: "rooftop",
        grant: "grant_123",
      }),
    ).resolves.toEqual({ kind: "locked" });

    vi.mocked(fetchMainPlayback).mockRejectedValueOnce(
      new ApiError("not found", {
        code: "http",
        status: 404,
      }),
    );
    await expect(
      loadMainPlaybackSurface("main_mina_quiet_rooftop", {
        fromShortId: "rooftop",
        grant: "grant_123",
      }),
    ).resolves.toEqual({ kind: "not_found" });
  });
});

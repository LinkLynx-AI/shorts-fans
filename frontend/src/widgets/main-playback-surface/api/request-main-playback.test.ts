import { viewerSessionCookieName } from "@/entities/viewer";
import { ApiError } from "@/shared/api";

import { fetchMainPlayback } from "./request-main-playback";

describe("fetchMainPlayback", () => {
  it("requests the playback endpoint with the grant query and session cookie", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
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
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_main_playback_001",
          },
        }),
        {
          headers: {
            "Content-Type": "application/json",
          },
          status: 200,
        },
      ),
    );

    const playback = await fetchMainPlayback({
      baseUrl: "https://app.example.com",
      fetcher,
      fromShortId: "rooftop",
      grant: "grant_123",
      mainId: "main_mina_quiet_rooftop",
      sessionToken: "viewer-session",
    });

    expect(playback.main.id).toBe("main_mina_quiet_rooftop");
    expect(fetcher).toHaveBeenCalledTimes(1);
    const [requestedUrl, init] = fetcher.mock.calls[0] ?? [];

    expect(requestedUrl).toBeInstanceOf(URL);
    expect(String(requestedUrl)).toBe(
      "https://app.example.com/api/fan/mains/main_mina_quiet_rooftop/playback?fromShortId=rooftop&grant=grant_123",
    );
    expect(new Headers(init?.headers).get("Cookie")).toBe(
      `${viewerSessionCookieName}=viewer-session`,
    );
  });

  it("throws an ApiError when the playback response is not success", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response("main locked", {
        status: 403,
      }),
    );

    await expect(
      fetchMainPlayback({
        baseUrl: "https://app.example.com",
        fetcher,
        fromShortId: "rooftop",
        grant: "invalid",
        mainId: "main_mina_quiet_rooftop",
      }),
    ).rejects.toBeInstanceOf(ApiError);
  });
});

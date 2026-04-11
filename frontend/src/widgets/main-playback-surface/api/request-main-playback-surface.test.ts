import { viewerSessionCookieName } from "@/entities/viewer";

import { requestMainPlaybackSurface } from "./request-main-playback-surface";

describe("requestMainPlaybackSurface", () => {
  it("requests the API playback surface and maps it to the widget model", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            access: {
              mainId: "main_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
              reason: "session_unlocked",
              status: "unlocked",
            },
            creator: {
              avatar: null,
              bio: "night preview specialist",
              displayName: "Mina Rei",
              handle: "@minarei",
              id: "creator_mina_rei",
            },
            entryShort: {
              caption: "quiet rooftop preview",
              canonicalMainId: "main_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
              creatorId: "creator_mina_rei",
              id: "short_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
              media: {
                durationSeconds: 16,
                id: "asset_short_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
                kind: "video",
                posterUrl: "https://cdn.example.com/shorts/poster.jpg",
                url: "https://cdn.example.com/shorts/playback.mp4",
              },
              previewDurationSeconds: 16,
            },
            main: {
              durationSeconds: 480,
              id: "main_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
              media: {
                durationSeconds: 480,
                id: "asset_main_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                kind: "video",
                posterUrl: "https://cdn.example.com/mains/poster.jpg",
                url: "https://cdn.example.com/mains/playback.mp4",
              },
            },
            resumePositionSeconds: null,
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_main_playback_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      requestMainPlaybackSurface({
        baseUrl: "https://api.example.com",
        fetcher,
        fromShortId: "short_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
        grant: "signed-grant",
        mainId: "main_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
        sessionToken: "raw-session-token",
      }),
    ).resolves.toMatchObject({
      access: {
        status: "unlocked",
      },
      entryShort: {
        id: "short_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
      },
      main: {
        id: "main_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      },
      themeShort: {
        id: "short_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
      },
      viewer: {
        isPinned: null,
      },
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe(
      "https://api.example.com/api/fan/mains/main_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/playback?fromShortId=short_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb&grant=signed-grant",
    );
    expect(fetcher.mock.calls[0]?.[1]?.credentials).toBe("include");
    expect(new Headers(fetcher.mock.calls[0]?.[1]?.headers).get("Cookie")).toBe(
      `${viewerSessionCookieName}=raw-session-token`,
    );
  });
});

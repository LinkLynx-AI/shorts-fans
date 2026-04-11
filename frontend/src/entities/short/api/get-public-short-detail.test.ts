import { viewerSessionCookieName } from "@/entities/viewer";

import { getPublicShortDetail } from "./get-public-short-detail";

describe("getPublicShortDetail", () => {
  it("requests the public short detail and forwards the session cookie", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            detail: {
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
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_short_detail_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getPublicShortDetail({
        baseUrl: "https://api.example.com",
        fetcher,
        sessionToken: "raw-session-token",
        shortId: "short_22222222222222222222222222222222",
      }),
    ).resolves.toEqual({
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

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe(
      "https://api.example.com/api/fan/shorts/short_22222222222222222222222222222222",
    );
    expect(new Headers(fetcher.mock.calls[0]?.[1]?.headers).get("Cookie")).toBe(
      `${viewerSessionCookieName}=raw-session-token`,
    );
  });
});

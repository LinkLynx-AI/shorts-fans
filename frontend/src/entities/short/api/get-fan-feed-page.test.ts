import { viewerSessionCookieName } from "@/entities/viewer";

import { getFanFeedPage } from "./get-fan-feed-page";

describe("getFanFeedPage", () => {
  it("requests the feed page and forwards the session cookie", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [
              {
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
                  isPinned: true,
                },
              },
            ],
            tab: "following",
          },
          error: null,
          meta: {
            page: {
              hasNext: true,
              nextCursor: "cursor_next_001",
            },
            requestId: "req_feed_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getFanFeedPage({
        baseUrl: "https://api.example.com",
        cursor: "cursor_prev_001",
        fetcher,
        sessionToken: "raw-session-token",
        tab: "following",
      }),
    ).resolves.toEqual({
      items: [
        {
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
            isPinned: true,
          },
        },
      ],
      page: {
        hasNext: true,
        nextCursor: "cursor_next_001",
      },
      requestId: "req_feed_001",
      tab: "following",
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe("https://api.example.com/api/fan/feed?tab=following&cursor=cursor_prev_001");
    expect(new Headers(fetcher.mock.calls[0]?.[1]?.headers).get("Cookie")).toBe(
      `${viewerSessionCookieName}=raw-session-token`,
    );
  });
});

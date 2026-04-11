import {
  fetchFanProfilePinnedShortsPage,
  type FanProfilePinnedShortsPage,
} from "@/entities/fan-profile";

describe("fetchFanProfilePinnedShortsPage", () => {
  it("returns a parsed pinned shorts page", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [
              {
                creator: {
                  avatar: null,
                  bio: "after rain と balcony mood の short をまとめています。",
                  displayName: "Sora Vale",
                  handle: "@soravale",
                  id: "creator_11111111111111111111111111111111",
                },
                short: {
                  caption: "after rain preview",
                  canonicalMainId: "main_22222222222222222222222222222222",
                  creatorId: "creator_11111111111111111111111111111111",
                  id: "short_33333333333333333333333333333333",
                  media: {
                    durationSeconds: 17,
                    id: "asset_44444444444444444444444444444444",
                    kind: "video",
                    posterUrl: "https://cdn.example.com/shorts/poster.jpg",
                    url: "https://cdn.example.com/shorts/playback.mp4",
                  },
                  previewDurationSeconds: 17,
                },
              },
            ],
          },
          error: null,
          meta: {
            page: {
              hasNext: false,
              nextCursor: null,
            },
            requestId: "req_fan_profile_pinned_shorts_001",
          },
        }),
        { status: 200 },
      ),
    );
    const expectedPage = {
      items: [
        {
          creator: {
            avatar: null,
            bio: "after rain と balcony mood の short をまとめています。",
            displayName: "Sora Vale",
            handle: "@soravale",
            id: "creator_11111111111111111111111111111111",
          },
          short: {
            caption: "after rain preview",
            canonicalMainId: "main_22222222222222222222222222222222",
            creatorId: "creator_11111111111111111111111111111111",
            id: "short_33333333333333333333333333333333",
            media: {
              durationSeconds: 17,
              id: "asset_44444444444444444444444444444444",
              kind: "video",
              posterUrl: "https://cdn.example.com/shorts/poster.jpg",
              url: "https://cdn.example.com/shorts/playback.mp4",
            },
            previewDurationSeconds: 17,
          },
        },
      ],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_fan_profile_pinned_shorts_001",
    } satisfies FanProfilePinnedShortsPage;

    await expect(
      fetchFanProfilePinnedShortsPage({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toEqual(expectedPage);
  });

  it("forwards the cursor and session cookie", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [],
          },
          error: null,
          meta: {
            page: {
              hasNext: true,
              nextCursor: "cursor_2",
            },
            requestId: "req_fan_profile_pinned_shorts_002",
          },
        }),
        { status: 200 },
      ),
    );

    await fetchFanProfilePinnedShortsPage({
      baseUrl: "https://api.example.com",
      cursor: "cursor_1",
      fetcher,
      sessionToken: "raw-session-token",
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe(
      "https://api.example.com/api/fan/profile/pinned-shorts?cursor=cursor_1",
    );

    const requestInit = fetcher.mock.calls[0]?.[1];
    const headers = new Headers(requestInit?.headers);

    expect(headers.get("Cookie")).toBe("shorts_fans_session=raw-session-token");
  });
});

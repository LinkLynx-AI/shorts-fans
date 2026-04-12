import {
  fetchFanProfileLibraryPage,
  type FanProfileLibraryPage,
} from "@/entities/fan-profile";

describe("fetchFanProfileLibraryPage", () => {
  it("returns a parsed library page", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [
              {
                access: {
                  mainId: "main_11111111111111111111111111111111",
                  reason: "session_unlocked",
                  status: "unlocked",
                },
                creator: {
                  avatar: null,
                  bio: "quiet rooftop と hotel light の preview を軸に投稿。",
                  displayName: "Mina Rei",
                  handle: "@minarei",
                  id: "creator_22222222222222222222222222222222",
                },
                entryShort: {
                  caption: "quiet rooftop preview",
                  canonicalMainId: "main_11111111111111111111111111111111",
                  creatorId: "creator_22222222222222222222222222222222",
                  id: "short_33333333333333333333333333333333",
                  media: {
                    durationSeconds: 16,
                    id: "asset_44444444444444444444444444444444",
                    kind: "video",
                    posterUrl: "https://cdn.example.com/shorts/poster.jpg",
                    url: "https://cdn.example.com/shorts/playback.mp4",
                  },
                  previewDurationSeconds: 16,
                },
                main: {
                  durationSeconds: 480,
                  id: "main_11111111111111111111111111111111",
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
            requestId: "req_fan_profile_library_001",
          },
        }),
        { status: 200 },
      ),
    );
    const expectedPage = {
      items: [
        {
          access: {
            mainId: "main_11111111111111111111111111111111",
            reason: "session_unlocked",
            status: "unlocked",
          },
          creator: {
            avatar: null,
            bio: "quiet rooftop と hotel light の preview を軸に投稿。",
            displayName: "Mina Rei",
            handle: "@minarei",
            id: "creator_22222222222222222222222222222222",
          },
          entryShort: {
            caption: "quiet rooftop preview",
            canonicalMainId: "main_11111111111111111111111111111111",
            creatorId: "creator_22222222222222222222222222222222",
            id: "short_33333333333333333333333333333333",
            media: {
              durationSeconds: 16,
              id: "asset_44444444444444444444444444444444",
              kind: "video",
              posterUrl: "https://cdn.example.com/shorts/poster.jpg",
              url: "https://cdn.example.com/shorts/playback.mp4",
            },
            previewDurationSeconds: 16,
          },
          main: {
            durationSeconds: 480,
            id: "main_11111111111111111111111111111111",
          },
        },
      ],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_fan_profile_library_001",
    } satisfies FanProfileLibraryPage;

    await expect(
      fetchFanProfileLibraryPage({
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
            requestId: "req_fan_profile_library_002",
          },
        }),
        { status: 200 },
      ),
    );

    await fetchFanProfileLibraryPage({
      baseUrl: "https://api.example.com",
      cursor: "cursor_1",
      fetcher,
      sessionToken: "raw-session-token",
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe(
      "https://api.example.com/api/fan/profile/library?cursor=cursor_1",
    );

    const requestInit = fetcher.mock.calls[0]?.[1];
    const headers = new Headers(requestInit?.headers);

    expect(headers.get("Cookie")).toBe("shorts_fans_session=raw-session-token");
  });
});

import { getCreatorProfileShortGrid } from "@/entities/creator";

describe("getCreatorProfileShortGrid", () => {
  it("requests the first page of creator profile shorts", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [
              {
                canonicalMainId: "main_mina_quiet_rooftop",
                creatorId: "creator_mina_rei",
                id: "short_mina_rooftop",
                media: {
                  durationSeconds: 16,
                  id: "asset_short_mina_rooftop",
                  kind: "video",
                  posterUrl: null,
                  url: "https://cdn.example.com/shorts/mina-rooftop.mp4",
                },
                previewDurationSeconds: 16,
              },
            ],
          },
          error: null,
          meta: {
            page: {
              hasNext: false,
              nextCursor: null,
            },
            requestId: "req_creator_profile_shorts_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getCreatorProfileShortGrid({
        baseUrl: "https://api.example.com",
        creatorId: "creator_mina_rei",
        fetcher,
      }),
    ).resolves.toEqual({
      items: [
        {
          canonicalMainId: "main_mina_quiet_rooftop",
          creatorId: "creator_mina_rei",
          id: "short_mina_rooftop",
          media: {
            durationSeconds: 16,
            id: "asset_short_mina_rooftop",
            kind: "video",
            posterUrl: null,
            url: "https://cdn.example.com/shorts/mina-rooftop.mp4",
          },
          previewDurationSeconds: 16,
        },
      ],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_creator_profile_shorts_001",
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe("https://api.example.com/api/fan/creators/creator_mina_rei/shorts");
  });

  it("forwards cursor when loading the next page", async () => {
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
            requestId: "req_creator_profile_shorts_next_001",
          },
        }),
        { status: 200 },
      ),
    );

    await getCreatorProfileShortGrid({
      baseUrl: "https://api.example.com",
      creatorId: "creator_mina_rei",
      cursor: "cursor_1",
      fetcher,
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe(
      "https://api.example.com/api/fan/creators/creator_mina_rei/shorts?cursor=cursor_1",
    );
  });
});

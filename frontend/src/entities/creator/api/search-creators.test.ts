import { getCreatorSearchResults } from "@/entities/creator";

describe("getCreatorSearchResults", () => {
  it("requests recent creators when query is empty", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [
              {
                creator: {
                  avatar: null,
                  bio: "Public shorts から paid main へつながる creator mock profile.",
                  displayName: "Mika Aoi",
                  handle: "@mikaaoi",
                  id: "creator_11111111111111111111111111111111",
                },
              },
            ],
            query: "",
          },
          error: null,
          meta: {
            page: {
              hasNext: false,
              nextCursor: null,
            },
            requestId: "req_search_recent_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getCreatorSearchResults({
        baseUrl: "https://api.example.com",
        fetcher,
        query: "   ",
      }),
    ).resolves.toEqual({
      items: [
        {
          avatar: null,
          bio: "Public shorts から paid main へつながる creator mock profile.",
          displayName: "Mika Aoi",
          handle: "@mikaaoi",
          id: "creator_11111111111111111111111111111111",
        },
      ],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      query: "",
      requestId: "req_search_recent_001",
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe("https://api.example.com/api/fan/creators/search");
  });

  it("forwards query and cursor to the API", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [],
            query: "mika",
          },
          error: null,
          meta: {
            page: {
              hasNext: true,
              nextCursor: "cursor_2",
            },
            requestId: "req_search_filtered_001",
          },
        }),
        { status: 200 },
      ),
    );

    await getCreatorSearchResults({
      baseUrl: "https://api.example.com",
      cursor: "cursor_1",
      fetcher,
      query: "mika",
    });

    expect(fetcher.mock.calls[0]?.[0].toString()).toBe(
      "https://api.example.com/api/fan/creators/search?q=mika&cursor=cursor_1",
    );
  });
});

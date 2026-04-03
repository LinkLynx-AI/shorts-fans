import { fetchRecommendedFeedShellState } from "@/widgets/feed-shell/api/recommended-feed";

describe("fetchRecommendedFeedShellState", () => {
  it("maps a recommended feed response to a ready shell state", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [
              {
                creator: {
                  avatar: {
                    durationSeconds: null,
                    id: "asset_creator_1",
                    kind: "image",
                    posterUrl: null,
                    url: "https://cdn.example.com/avatar.jpg",
                  },
                  bio: "quiet rooftop と hotel light の preview を軸に投稿。",
                  displayName: "Mina Rei",
                  handle: "@minarei",
                  id: "11111111-1111-4111-8111-111111111111",
                },
                short: {
                  caption: "quiet rooftop preview。",
                  canonicalMainId: "22222222-2222-4222-8222-222222222222",
                  creatorId: "11111111-1111-4111-8111-111111111111",
                  id: "33333333-3333-4333-8333-333333333333",
                  media: {
                    durationSeconds: 16,
                    id: "asset_short_1",
                    kind: "video",
                    posterUrl: null,
                    url: "https://cdn.example.com/short.mp4",
                  },
                  previewDurationSeconds: 16,
                  title: "quiet rooftop preview",
                },
                unlockCta: {
                  mainDurationSeconds: 480,
                  priceJpy: 1800,
                  resumePositionSeconds: null,
                  state: "unlock_available",
                },
                viewer: {
                  isPinned: false,
                },
              },
            ],
            tab: "recommended",
          },
          error: null,
          meta: {
            page: {
              hasNext: true,
              nextCursor: "feed:recommended:cursor:001",
            },
            requestId: "req_feed_001",
          },
        }),
        { status: 200 },
      ),
    );

    const state = await fetchRecommendedFeedShellState({
      baseUrl: "https://api.example.com",
      fetcher,
    });

    expect(state.kind).toBe("ready");
    if (state.kind !== "ready") {
      throw new Error("fixture missing");
    }

    expect(state.tab).toBe("recommended");
    expect(state.detailHref).toBeUndefined();
    expect(state.page).toEqual({
      hasNext: true,
      nextCursor: "feed:recommended:cursor:001",
    });
    expect(state.surface.short.id).toBe("33333333-3333-4333-8333-333333333333");
    expect(state.surface.creator.displayName).toBe("Mina Rei");
  });

  it("maps an empty recommended response to an empty shell state", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [],
            tab: "recommended",
          },
          error: null,
          meta: {
            page: {
              hasNext: false,
              nextCursor: null,
            },
            requestId: "req_feed_002",
          },
        }),
        { status: 200 },
      ),
    );

    const state = await fetchRecommendedFeedShellState({
      baseUrl: "https://api.example.com",
      fetcher,
    });

    expect(state).toEqual({
      kind: "empty",
      tab: "recommended",
    });
  });
});

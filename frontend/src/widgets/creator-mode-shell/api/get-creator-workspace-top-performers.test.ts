import { getCreatorWorkspaceTopPerformers } from "./get-creator-workspace-top-performers";

describe("getCreatorWorkspaceTopPerformers", () => {
  it("parses the top performers response", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            topPerformers: {
              topMain: {
                id: "main_quiet_rooftop",
                media: {
                  durationSeconds: 720,
                  id: "asset_main_quiet_rooftop",
                  kind: "video",
                  posterUrl: "https://signed.example.com/mains/top.jpg",
                },
                unlockCount: 238,
              },
              topShort: {
                attributedUnlockCount: 238,
                id: "short_quiet_rooftop",
                media: {
                  durationSeconds: 16,
                  id: "asset_short_quiet_rooftop",
                  kind: "video",
                  posterUrl: "https://cdn.example.com/shorts/top.jpg",
                },
              },
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_workspace_top_performers_001",
          },
        }),
        {
          status: 200,
        },
      ),
    );

    await expect(
      getCreatorWorkspaceTopPerformers({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toEqual({
      topMain: {
        id: "main_quiet_rooftop",
        media: {
          durationSeconds: 720,
          id: "asset_main_quiet_rooftop",
          kind: "video",
          posterUrl: "https://signed.example.com/mains/top.jpg",
        },
        unlockCount: 238,
      },
      topShort: {
        attributedUnlockCount: 238,
        id: "short_quiet_rooftop",
        media: {
          durationSeconds: 16,
          id: "asset_short_quiet_rooftop",
          kind: "video",
          posterUrl: "https://cdn.example.com/shorts/top.jpg",
        },
      },
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/creator/workspace/top-performers"),
      expect.objectContaining({
        cache: "no-store",
        credentials: "include",
      }),
    );
  });

  it("accepts null performers", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            topPerformers: {
              topMain: null,
              topShort: null,
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_workspace_top_performers_002",
          },
        }),
        {
          status: 200,
        },
      ),
    );

    await expect(
      getCreatorWorkspaceTopPerformers({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toEqual({
      topMain: null,
      topShort: null,
    });
  });
});

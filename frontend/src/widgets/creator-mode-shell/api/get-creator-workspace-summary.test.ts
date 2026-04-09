import { getCreatorWorkspaceSummary } from "./get-creator-workspace-summary";

describe("getCreatorWorkspaceSummary", () => {
  it("parses the creator workspace summary response", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            workspace: {
              creator: {
                avatar: {
                  durationSeconds: null,
                  id: "asset_creator_api_avatar",
                  kind: "image",
                  posterUrl: null,
                  url: "https://cdn.example.com/creator/api/avatar.jpg",
                },
                bio: "contract-backed bio",
                displayName: "API Creator",
                handle: "@apicreator",
                id: "creator_api_001",
              },
              overviewMetrics: {
                grossUnlockRevenueJpy: 82000,
                unlockCount: 91,
                uniquePurchaserCount: 74,
              },
              revisionRequestedSummary: {
                mainCount: 1,
                shortCount: 0,
                totalCount: 1,
              },
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_workspace_001",
          },
        }),
        {
          status: 200,
        },
      ),
    );

    await expect(
      getCreatorWorkspaceSummary({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toEqual({
      creator: {
        avatar: {
          durationSeconds: null,
          id: "asset_creator_api_avatar",
          kind: "image",
          posterUrl: null,
          url: "https://cdn.example.com/creator/api/avatar.jpg",
        },
        bio: "contract-backed bio",
        displayName: "API Creator",
        handle: "@apicreator",
        id: "creator_api_001",
      },
      overviewMetrics: {
        grossUnlockRevenueJpy: 82000,
        unlockCount: 91,
        uniquePurchaserCount: 74,
      },
      revisionRequestedSummary: {
        mainCount: 1,
        shortCount: 0,
        totalCount: 1,
      },
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/creator/workspace"),
      expect.objectContaining({
        cache: "no-store",
        credentials: "include",
      }),
    );
  });

  it("accepts a null revision summary", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            workspace: {
              creator: {
                avatar: null,
                bio: "after rain",
                displayName: "Sora Vale",
                handle: "@soravale",
                id: "creator_sora_vale",
              },
              overviewMetrics: {
                grossUnlockRevenueJpy: 12000,
                unlockCount: 18,
                uniquePurchaserCount: 16,
              },
              revisionRequestedSummary: null,
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_workspace_002",
          },
        }),
        {
          status: 200,
        },
      ),
    );

    await expect(
      getCreatorWorkspaceSummary({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toMatchObject({
      revisionRequestedSummary: null,
    });
  });

  it("surfaces non-success statuses as API errors", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response("server error", {
        status: 500,
      }),
    );

    await expect(
      getCreatorWorkspaceSummary({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).rejects.toMatchObject({
      code: "http",
      status: 500,
    });
  });
});

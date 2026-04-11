import {
  getCreatorWorkspacePreviewMains,
  getCreatorWorkspacePreviewShorts,
} from "./get-creator-workspace-preview-collections";

describe("creator workspace preview collection fetchers", () => {
  it("parses the owner preview short list response", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [
              {
                canonicalMainId: "main_quiet_rooftop",
                id: "short_quiet_rooftop",
                media: {
                  durationSeconds: 16,
                  id: "asset_short_quiet_rooftop",
                  kind: "video",
                  posterUrl: "https://cdn.example.com/creator/preview/shorts/quiet-rooftop-poster.jpg",
                },
                previewDurationSeconds: 16,
              },
            ],
          },
          error: null,
          meta: {
            page: {
              hasNext: true,
              nextCursor: "next-short-cursor",
            },
            requestId: "req_creator_workspace_shorts_001",
          },
        }),
        {
          status: 200,
        },
      ),
    );

    await expect(
      getCreatorWorkspacePreviewShorts({
        baseUrl: "https://api.example.com",
        cursor: "cursor-short",
        fetcher,
      }),
    ).resolves.toEqual({
      items: [
        {
          canonicalMainId: "main_quiet_rooftop",
          id: "short_quiet_rooftop",
          media: {
            durationSeconds: 16,
            id: "asset_short_quiet_rooftop",
            kind: "video",
            posterUrl: "https://cdn.example.com/creator/preview/shorts/quiet-rooftop-poster.jpg",
          },
          previewDurationSeconds: 16,
        },
      ],
      page: {
        hasNext: true,
        nextCursor: "next-short-cursor",
      },
      requestId: "req_creator_workspace_shorts_001",
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/creator/workspace/shorts?cursor=cursor-short"),
      expect.objectContaining({
        cache: "no-store",
        credentials: "include",
      }),
    );
  });

  it("parses the owner preview main list response", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [
              {
                durationSeconds: 720,
                id: "main_quiet_rooftop",
                leadShortId: "short_quiet_rooftop",
                media: {
                  durationSeconds: 720,
                  id: "asset_main_quiet_rooftop",
                  kind: "video",
                  posterUrl: "https://cdn.example.com/creator/preview/mains/quiet-rooftop-poster.jpg",
                },
                priceJpy: 1800,
              },
            ],
          },
          error: null,
          meta: {
            page: {
              hasNext: false,
              nextCursor: null,
            },
            requestId: "req_creator_workspace_mains_001",
          },
        }),
        {
          status: 200,
        },
      ),
    );

    await expect(
      getCreatorWorkspacePreviewMains({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toMatchObject({
      items: [
        {
          durationSeconds: 720,
          id: "main_quiet_rooftop",
          leadShortId: "short_quiet_rooftop",
          priceJpy: 1800,
        },
      ],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      requestId: "req_creator_workspace_mains_001",
    });
  });

  it("surfaces non-success statuses as API errors", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response("server error", {
        status: 500,
      }),
    );

    await expect(
      getCreatorWorkspacePreviewShorts({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).rejects.toMatchObject({
      code: "http",
      status: 500,
    });
  });
});

import {
  getCreatorWorkspacePreviewMainDetail,
  getCreatorWorkspacePreviewShortDetail,
} from "./get-creator-workspace-preview-detail";

describe("creator workspace preview detail fetchers", () => {
  it("parses the owner preview short detail response", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            preview: {
              access: {
                mainId: "main_quiet_rooftop",
                reason: "owner_preview",
                status: "owner",
              },
              creator: {
                avatar: null,
                bio: "quiet rooftop と hotel light の preview を軸に投稿。",
                displayName: "Mina Rei",
                handle: "@minarei",
                id: "creator_mina_rei",
              },
              short: {
                caption: "quiet rooftop preview.",
                canonicalMainId: "main_quiet_rooftop",
                creatorId: "creator_mina_rei",
                id: "short_quiet_rooftop",
                media: {
                  durationSeconds: 16,
                  id: "asset_short_quiet_rooftop",
                  kind: "video",
                  posterUrl: "https://cdn.example.com/creator/preview/shorts/quiet-rooftop-poster.jpg",
                  url: "https://cdn.example.com/creator/preview/shorts/quiet-rooftop.mp4",
                },
                previewDurationSeconds: 16,
              },
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_workspace_short_detail_001",
          },
        }),
        {
          status: 200,
        },
      ),
    );

    await expect(
      getCreatorWorkspacePreviewShortDetail("short_quiet_rooftop", {
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toEqual({
      access: {
        mainId: "main_quiet_rooftop",
        reason: "owner_preview",
        status: "owner",
      },
      creator: {
        avatar: null,
        bio: "quiet rooftop と hotel light の preview を軸に投稿。",
        displayName: "Mina Rei",
        handle: "@minarei",
        id: "creator_mina_rei",
      },
      kind: "preview-short",
      requestId: "req_creator_workspace_short_detail_001",
      short: {
        caption: "quiet rooftop preview.",
        canonicalMainId: "main_quiet_rooftop",
        creatorId: "creator_mina_rei",
        id: "short_quiet_rooftop",
        media: {
          durationSeconds: 16,
          id: "asset_short_quiet_rooftop",
          kind: "video",
          posterUrl: "https://cdn.example.com/creator/preview/shorts/quiet-rooftop-poster.jpg",
          url: "https://cdn.example.com/creator/preview/shorts/quiet-rooftop.mp4",
        },
        previewDurationSeconds: 16,
      },
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/creator/workspace/shorts/short_quiet_rooftop/preview"),
      expect.objectContaining({
        cache: "no-store",
        credentials: "include",
      }),
    );
  });

  it("parses the owner preview main detail response", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            preview: {
              access: {
                mainId: "main_quiet_rooftop",
                reason: "owner_preview",
                status: "owner",
              },
              creator: {
                avatar: null,
                bio: "quiet rooftop と hotel light の preview を軸に投稿。",
                displayName: "Mina Rei",
                handle: "@minarei",
                id: "creator_mina_rei",
              },
              entryShort: {
                caption: "quiet rooftop preview.",
                canonicalMainId: "main_quiet_rooftop",
                creatorId: "creator_mina_rei",
                id: "short_quiet_rooftop",
                media: {
                  durationSeconds: 16,
                  id: "asset_short_quiet_rooftop",
                  kind: "video",
                  posterUrl: "https://cdn.example.com/creator/preview/shorts/quiet-rooftop-poster.jpg",
                  url: "https://cdn.example.com/creator/preview/shorts/quiet-rooftop.mp4",
                },
                previewDurationSeconds: 16,
              },
              main: {
                durationSeconds: 720,
                id: "main_quiet_rooftop",
                media: {
                  durationSeconds: 720,
                  id: "asset_main_quiet_rooftop",
                  kind: "video",
                  posterUrl: "https://cdn.example.com/creator/preview/mains/quiet-rooftop-poster.jpg",
                  url: "https://cdn.example.com/creator/preview/mains/quiet-rooftop.mp4",
                },
                priceJpy: 1800,
              },
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_workspace_main_detail_001",
          },
        }),
        {
          status: 200,
        },
      ),
    );

    await expect(
      getCreatorWorkspacePreviewMainDetail("main_quiet_rooftop", {
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toMatchObject({
      access: {
        mainId: "main_quiet_rooftop",
        reason: "owner_preview",
        status: "owner",
      },
      entryShort: {
        id: "short_quiet_rooftop",
      },
      kind: "preview-main",
      main: {
        durationSeconds: 720,
        id: "main_quiet_rooftop",
        media: {
          url: "https://cdn.example.com/creator/preview/mains/quiet-rooftop.mp4",
        },
        priceJpy: 1800,
      },
      requestId: "req_creator_workspace_main_detail_001",
    });
  });

  it("surfaces non-success statuses as API errors", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response("server error", {
        status: 500,
      }),
    );

    await expect(
      getCreatorWorkspacePreviewMainDetail("main_quiet_rooftop", {
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).rejects.toMatchObject({
      code: "http",
      status: 500,
    });
  });
});

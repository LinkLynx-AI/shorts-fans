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
                avatar: {
                  durationSeconds: null,
                  id: "asset_creator_avatar",
                  kind: "image",
                  posterUrl: null,
                  url: "https://cdn.example.com/creator/avatar.jpg",
                },
                bio: "owner preview bio",
                displayName: "Mina Rei",
                handle: "@minarei",
                id: "creator_mina_rei",
              },
              short: {
                caption: "blue tone の balcony preview。",
                canonicalMainId: "main_quiet_rooftop",
                creatorId: "creator_mina_rei",
                id: "short_quiet_rooftop",
                media: {
                  durationSeconds: 16,
                  id: "asset_short_quiet_rooftop",
                  kind: "video",
                  posterUrl: "https://cdn.example.com/shorts/poster.jpg",
                  url: "https://cdn.example.com/shorts/playback.mp4",
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
        { status: 200 },
      ),
    );

    await expect(
      getCreatorWorkspacePreviewShortDetail("short_quiet_rooftop", {
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toMatchObject({
      access: {
        mainId: "main_quiet_rooftop",
        reason: "owner_preview",
        status: "owner",
      },
      short: {
        caption: "blue tone の balcony preview。",
        id: "short_quiet_rooftop",
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
                avatar: {
                  durationSeconds: null,
                  id: "asset_creator_avatar",
                  kind: "image",
                  posterUrl: null,
                  url: "https://cdn.example.com/creator/avatar.jpg",
                },
                bio: "owner preview bio",
                displayName: "Mina Rei",
                handle: "@minarei",
                id: "creator_mina_rei",
              },
              entryShort: {
                caption: "quiet rooftop preview。",
                canonicalMainId: "main_quiet_rooftop",
                creatorId: "creator_mina_rei",
                id: "short_quiet_rooftop",
                media: {
                  durationSeconds: 16,
                  id: "asset_short_quiet_rooftop",
                  kind: "video",
                  posterUrl: "https://cdn.example.com/shorts/poster.jpg",
                  url: "https://cdn.example.com/shorts/playback.mp4",
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
                  posterUrl: "https://signed.example.com/mains/poster.jpg",
                  url: "https://signed.example.com/mains/playback.mp4",
                },
                priceJpy: 2200,
              },
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_workspace_main_detail_001",
          },
        }),
        { status: 200 },
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
      main: {
        id: "main_quiet_rooftop",
        priceJpy: 2200,
      },
    });
  });
});

import {
  getCreatorWorkspaceMainReviewSurface,
  getCreatorWorkspaceReviewSurface,
  getCreatorWorkspaceShortReviewSurface,
} from "./get-creator-workspace-review-surface";

describe("creator workspace review surface fetchers", () => {
  it("parses the dashboard review surface response", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            reviewSurface: {
              mains: [
                {
                  id: "main_aoi_blue_balcony",
                  state: "revision_requested",
                },
              ],
              packages: [
                {
                  blockers: [],
                  canonicalMainId: "main_aoi_blue_balcony",
                  linkedShortCount: 2,
                  readiness: "ready",
                  reviewStatus: "changes_requested",
                  submitAction: "resubmit",
                },
              ],
              shorts: [
                {
                  canonicalMainId: "main_aoi_blue_balcony",
                  id: "short_aoi_balcony_close",
                  state: "revision_requested",
                },
              ],
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_workspace_review_surface_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getCreatorWorkspaceReviewSurface({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toEqual({
      mains: [
        {
          id: "main_aoi_blue_balcony",
          state: "revision_requested",
        },
      ],
      packages: [
        {
          blockers: [],
          canonicalMainId: "main_aoi_blue_balcony",
          linkedShortCount: 2,
          readiness: "ready",
          reviewStatus: "changes_requested",
          submitAction: "resubmit",
        },
      ],
      requestId: "req_creator_workspace_review_surface_001",
      shorts: [
        {
          canonicalMainId: "main_aoi_blue_balcony",
          id: "short_aoi_balcony_close",
          state: "revision_requested",
        },
      ],
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/creator/workspace/review-surface"),
      expect.objectContaining({
        cache: "no-store",
        credentials: "include",
      }),
    );
  });

  it("parses the main detail review surface response", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            reviewSurface: {
              package: {
                blockers: [],
                canonicalMainId: "main_aoi_blue_balcony",
                linkedShortCount: 2,
                readiness: "ready",
                reviewStatus: "changes_requested",
                submitAction: "resubmit",
              },
              review: {
                reasonCode: "thumbnail_policy_adjustment",
                state: "revision_requested",
              },
              target: {
                canonicalMainId: "main_aoi_blue_balcony",
                id: "main_aoi_blue_balcony",
                kind: "main",
              },
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_workspace_main_review_surface_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getCreatorWorkspaceMainReviewSurface("main_aoi_blue_balcony", {
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toMatchObject({
      package: {
        canonicalMainId: "main_aoi_blue_balcony",
        submitAction: "resubmit",
      },
      requestId: "req_creator_workspace_main_review_surface_001",
      review: {
        reasonCode: "thumbnail_policy_adjustment",
      },
      target: {
        id: "main_aoi_blue_balcony",
        kind: "main",
      },
    });
  });

  it("parses the short detail review surface response", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            reviewSurface: {
              package: {
                blockers: [
                  "main_price_missing",
                ],
                canonicalMainId: "main_aoi_blue_balcony",
                linkedShortCount: 2,
                readiness: "blocked",
                reviewStatus: "draft",
                submitAction: "none",
              },
              review: {
                reasonCode: null,
                state: "draft",
              },
              target: {
                canonicalMainId: "main_aoi_blue_balcony",
                id: "short_aoi_balcony_close",
                kind: "short",
              },
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_workspace_short_review_surface_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getCreatorWorkspaceShortReviewSurface("short_aoi_balcony_close", {
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toMatchObject({
      package: {
        blockers: ["main_price_missing"],
        readiness: "blocked",
      },
      target: {
        id: "short_aoi_balcony_close",
        kind: "short",
      },
    });
  });

  it("accepts opaque object states from the contract", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            reviewSurface: {
              package: {
                blockers: [],
                canonicalMainId: "main_future_contract_state",
                linkedShortCount: 1,
                readiness: "none",
                reviewStatus: "approved",
                submitAction: "none",
              },
              review: {
                reasonCode: null,
                state: "future_review_state",
              },
              target: {
                canonicalMainId: "main_future_contract_state",
                id: "main_future_contract_state",
                kind: "main",
              },
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_creator_workspace_main_review_surface_opaque_state",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getCreatorWorkspaceMainReviewSurface("main_future_contract_state", {
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toMatchObject({
      review: {
        state: "future_review_state",
      },
    });
  });
});

import { getCreatorReviewQueue } from "./get-creator-review-queue";

describe("getCreatorReviewQueue", () => {
  it("requests the admin queue for the given state", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [
              {
                creatorBio: "quiet rooftop",
                legalName: "Mina Rei",
                review: {
                  approvedAt: null,
                  rejectedAt: null,
                  submittedAt: "2026-04-18T09:00:00Z",
                  suspendedAt: null,
                },
                sharedProfile: {
                  avatar: null,
                  displayName: "Mina Rei",
                  handle: "@minarei",
                },
                state: "submitted",
                userId: "11111111-1111-1111-1111-111111111111",
              },
            ],
            state: "submitted",
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_admin_creator_review_queue_001",
          },
        }),
        {
          status: 200,
        },
      ),
    );

    await expect(
      getCreatorReviewQueue({
        baseUrl: "https://api.example.com",
        fetcher,
        state: "submitted",
      }),
    ).resolves.toEqual({
      items: [
        expect.objectContaining({
          creatorBio: "quiet rooftop",
          state: "submitted",
          userId: "11111111-1111-1111-1111-111111111111",
        }),
      ],
      state: "submitted",
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/admin/creator-reviews?state=submitted"),
      expect.objectContaining({
        cache: "no-store",
        credentials: "include",
      }),
    );
  });

  it("surfaces non-success responses as API errors", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response("bad request", {
        status: 400,
      }),
    );

    await expect(
      getCreatorReviewQueue({
        baseUrl: "https://api.example.com",
        fetcher,
        state: "approved",
      }),
    ).rejects.toMatchObject({
      code: "http",
      status: 400,
    });
  });
});

import { getSubmissionReviewQueue } from "./get-submission-review-queue";

describe("getSubmissionReviewQueue", () => {
  it("requests the pending submission review queue", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            items: [
              {
                creator: {
                  avatar: null,
                  bio: "quiet rooftop",
                  displayName: "Mina Rei",
                  handle: "@minarei",
                  id: "creator_11111111111111111111111111111111",
                },
                intakeId: "11111111-1111-1111-1111-111111111111",
                mainDecisionRequired: true,
                pendingShortCount: 2,
                shortCount: 3,
                submitKind: "initial_submit",
                submittedAt: "2026-04-20T09:00:00Z",
              },
            ],
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_admin_submission_review_queue_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getSubmissionReviewQueue({
        baseUrl: "https://api.example.com",
        fetcher,
      }),
    ).resolves.toEqual({
      items: [
        expect.objectContaining({
          intakeId: "11111111-1111-1111-1111-111111111111",
          mainDecisionRequired: true,
        }),
      ],
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/admin/submission-reviews"),
      expect.objectContaining({
        cache: "no-store",
        credentials: "include",
      }),
    );
  });
});

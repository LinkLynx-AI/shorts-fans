import { applySubmissionReviewDecision } from "./apply-submission-review-decision";

describe("applySubmissionReviewDecision", () => {
  it("posts object-level decisions and accepts 204 no content", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }));

    await expect(
      applySubmissionReviewDecision({
        baseUrl: "https://api.example.com",
        fetcher,
        intakeId: "11111111-1111-1111-1111-111111111111",
        mainDecision: {
          decision: "approved",
          reviewNote: "unlock ready",
        },
        shortDecisions: [
          {
            decision: "revision_requested",
            reasonCode: "quality_issue",
            reviewNote: "reframe intro",
            shortId: "55555555-5555-5555-5555-555555555555",
          },
        ],
      }),
    ).resolves.toBeUndefined();

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/admin/submission-reviews/11111111-1111-1111-1111-111111111111/decision"),
      expect.objectContaining({
        body: JSON.stringify({
          mainDecision: {
            decision: "approved",
            reasonCode: "",
            reviewNote: "unlock ready",
          },
          shortDecisions: [
            {
              decision: "revision_requested",
              reasonCode: "quality_issue",
              reviewNote: "reframe intro",
              shortId: "55555555-5555-5555-5555-555555555555",
            },
          ],
        }),
        credentials: "include",
        method: "POST",
      }),
    );
  });
});

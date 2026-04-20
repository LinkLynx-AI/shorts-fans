import { getSubmissionReviewCase } from "./get-submission-review-case";

describe("getSubmissionReviewCase", () => {
  it("parses the intake detail response", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            case: {
              creator: {
                avatar: null,
                bio: "quiet rooftop",
                displayName: "Mina Rei",
                handle: "@minarei",
                id: "creator_11111111111111111111111111111111",
              },
              intake: {
                canonicalMainId: "22222222-2222-2222-2222-222222222222",
                consentConfirmed: true,
                creatorUserId: "33333333-3333-3333-3333-333333333333",
                id: "11111111-1111-1111-1111-111111111111",
                mainMediaAssetId: "44444444-4444-4444-4444-444444444444",
                mainPriceJpy: 1800,
                ownershipConfirmed: true,
                previousIntakeId: null,
                status: "pending_review",
                submitKind: "initial_submit",
                submittedAt: "2026-04-20T09:00:00Z",
              },
              main: {
                currencyCode: "JPY",
                decisionRequired: true,
                id: "22222222-2222-2222-2222-222222222222",
                intakeDecisionLog: null,
                media: {
                  durationSeconds: 540,
                  id: "asset_main_1",
                  kind: "video",
                  posterUrl: "https://cdn.example.com/mains/poster.jpg",
                  url: "https://cdn.example.com/mains/video.mp4",
                },
                priceJpy: 1800,
                review: {
                  decisionSource: null,
                  decisionedAt: null,
                  reasonCode: null,
                  reviewNote: null,
                },
                state: "pending_review",
              },
              shorts: [
                {
                  caption: "quiet rooftop cut",
                  decisionRequired: true,
                  id: "55555555-5555-5555-5555-555555555555",
                  intakeDecisionLog: null,
                  media: {
                    durationSeconds: 18,
                    id: "asset_short_1",
                    kind: "video",
                    posterUrl: "https://cdn.example.com/shorts/poster.jpg",
                    url: "https://cdn.example.com/shorts/video.mp4",
                  },
                  review: {
                    decisionSource: null,
                    decisionedAt: null,
                    reasonCode: null,
                    reviewNote: null,
                  },
                  state: "pending_review",
                },
              ],
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_admin_submission_review_case_001",
          },
        }),
        { status: 200 },
      ),
    );

    await expect(
      getSubmissionReviewCase({
        baseUrl: "https://api.example.com",
        fetcher,
        intakeId: "11111111-1111-1111-1111-111111111111",
      }),
    ).resolves.toMatchObject({
      intake: expect.objectContaining({
        id: "11111111-1111-1111-1111-111111111111",
      }),
      main: expect.objectContaining({
        state: "pending_review",
      }),
    });
  });
});

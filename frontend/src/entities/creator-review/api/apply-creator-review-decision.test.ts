import { applyCreatorReviewDecision } from "./apply-creator-review-decision";

describe("applyCreatorReviewDecision", () => {
  it("posts the selected decision and returns the updated case", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            case: {
              creatorBio: "quiet rooftop",
              evidences: [],
              intake: {
                acceptsConsentResponsibility: true,
                birthDate: "1999-04-02",
                declaresNoProhibitedCategory: true,
                legalName: "Mina Rei",
                payoutRecipientName: "Mina Rei",
                payoutRecipientType: "self",
              },
              rejection: {
                isResubmitEligible: false,
                isSupportReviewRequired: false,
                reasonCode: "documents_blurry",
                selfServeResubmitCount: 1,
                selfServeResubmitRemaining: 1,
              },
              review: {
                approvedAt: null,
                rejectedAt: "2026-04-18T10:00:00Z",
                submittedAt: "2026-04-18T09:00:00Z",
                suspendedAt: null,
              },
              sharedProfile: {
                avatar: null,
                displayName: "Mina Rei",
                handle: "@minarei",
              },
              state: "rejected",
              userId: "11111111-1111-1111-1111-111111111111",
            },
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_admin_creator_review_decision_001",
          },
        }),
        {
          status: 200,
        },
      ),
    );

    await expect(
      applyCreatorReviewDecision({
        baseUrl: "https://api.example.com",
        decision: "rejected",
        fetcher,
        reasonCode: "documents_blurry",
        userId: "11111111-1111-1111-1111-111111111111",
      }),
    ).resolves.toMatchObject({
      rejection: expect.objectContaining({
        reasonCode: "documents_blurry",
      }),
      state: "rejected",
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/admin/creator-reviews/11111111-1111-1111-1111-111111111111/decision"),
      expect.objectContaining({
        body: JSON.stringify({
          decision: "rejected",
          isResubmitEligible: false,
          isSupportReviewRequired: false,
          reasonCode: "documents_blurry",
        }),
        credentials: "include",
        method: "POST",
      }),
    );
  });

  it("surfaces conflicts as API errors", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response("conflict", {
        status: 409,
      }),
    );

    await expect(
      applyCreatorReviewDecision({
        baseUrl: "https://api.example.com",
        decision: "approved",
        fetcher,
        userId: "11111111-1111-1111-1111-111111111111",
      }),
    ).rejects.toMatchObject({
      code: "http",
      status: 409,
    });
  });
});

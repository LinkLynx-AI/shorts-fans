import { getCreatorReviewCase } from "./get-creator-review-case";

describe("getCreatorReviewCase", () => {
  it("parses the admin review detail response", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            case: {
              creatorBio: "quiet rooftop",
              evidences: [
                {
                  accessUrl: "https://signed.example.com/government-id",
                  fileName: "government-id.png",
                  fileSizeBytes: 183442,
                  kind: "government_id",
                  mimeType: "image/png",
                  uploadedAt: "2026-04-18T08:45:00Z",
                },
              ],
              intake: {
                acceptsConsentResponsibility: true,
                birthDate: "1999-04-02",
                declaresNoProhibitedCategory: true,
                legalName: "Mina Rei",
                payoutRecipientName: "Mina Rei",
                payoutRecipientType: "self",
              },
              rejection: null,
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
          },
          error: null,
          meta: {
            page: null,
            requestId: "req_admin_creator_review_case_001",
          },
        }),
        {
          status: 200,
        },
      ),
    );

    await expect(
      getCreatorReviewCase({
        baseUrl: "https://api.example.com",
        fetcher,
        userId: "11111111-1111-1111-1111-111111111111",
      }),
    ).resolves.toMatchObject({
      state: "submitted",
      userId: "11111111-1111-1111-1111-111111111111",
    });

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/admin/creator-reviews/11111111-1111-1111-1111-111111111111"),
      expect.objectContaining({
        cache: "no-store",
        credentials: "include",
      }),
    );
  });

  it("surfaces 404 responses as API errors", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(
      new Response("not found", {
        status: 404,
      }),
    );

    await expect(
      getCreatorReviewCase({
        baseUrl: "https://api.example.com",
        fetcher,
        userId: "11111111-1111-1111-1111-111111111111",
      }),
    ).rejects.toMatchObject({
      code: "http",
      status: 404,
    });
  });
});

import { ApiError } from "@/shared/api";

import {
  createCreatorWorkspaceSubmissionReview,
  CreatorWorkspaceSubmissionReviewApiError,
} from "./create-creator-workspace-submission-review";

describe("createCreatorWorkspaceSubmissionReview", () => {
  it("posts the current package submit with credentials included", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }));

    await expect(createCreatorWorkspaceSubmissionReview({
      baseUrl: "https://api.example.com",
      fetcher,
      mainId: "main_quiet_rooftop",
    })).resolves.toBeUndefined();

    expect(fetcher).toHaveBeenCalledWith(
      new URL("https://api.example.com/api/creator/workspace/mains/main_quiet_rooftop/review-submissions"),
      expect.objectContaining({
        credentials: "include",
        method: "POST",
      }),
    );
  });

  it("surfaces contract errors as CreatorWorkspaceSubmissionReviewApiError", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({
      data: null,
      error: {
        code: "review_state_conflict",
        message: "creator workspace submission package could not be submitted from the current review state",
      },
      meta: {
        page: null,
        requestId: "req_creator_workspace_submission_review_post_001",
      },
    }), { status: 409 }));

    await expect(createCreatorWorkspaceSubmissionReview({
      baseUrl: "https://api.example.com",
      fetcher,
      mainId: "main_quiet_rooftop",
    })).rejects.toEqual(new CreatorWorkspaceSubmissionReviewApiError(
      "review_state_conflict",
      "creator workspace submission package could not be submitted from the current review state",
      {
        requestId: "req_creator_workspace_submission_review_post_001",
        status: 409,
      },
    ));
  });

  it("surfaces malformed error payloads as ApiError", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(new Response("server error", { status: 500 }));

    await expect(createCreatorWorkspaceSubmissionReview({
      baseUrl: "https://api.example.com",
      fetcher,
      mainId: "main_quiet_rooftop",
    })).rejects.toBeInstanceOf(ApiError);
  });
});

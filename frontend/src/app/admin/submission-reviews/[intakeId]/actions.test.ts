import { applySubmissionReviewDecision } from "@/entities/submission-review";

import { assertAdminUiAccess } from "../../_lib/admin-ui-access";
import { applySubmissionReviewDecisionFromAdmin } from "./actions";

vi.mock("../../_lib/admin-ui-access", () => ({
  assertAdminUiAccess: vi.fn(),
}));

vi.mock("@/entities/submission-review", async () => {
  const actual = await vi.importActual<typeof import("@/entities/submission-review")>("@/entities/submission-review");

  return {
    ...actual,
    applySubmissionReviewDecision: vi.fn(),
  };
});

describe("applySubmissionReviewDecisionFromAdmin", () => {
  beforeEach(() => {
    vi.mocked(assertAdminUiAccess).mockReset();
    vi.mocked(applySubmissionReviewDecision).mockReset();
  });

  it("checks admin UI access before applying object decisions", async () => {
    vi.mocked(applySubmissionReviewDecision).mockResolvedValue(undefined);

    await applySubmissionReviewDecisionFromAdmin({
      intakeId: "11111111-1111-1111-1111-111111111111",
      mainDecision: {
        decision: "approved",
      },
      shortDecisions: [],
    });

    expect(assertAdminUiAccess).toHaveBeenCalledTimes(1);
    expect(applySubmissionReviewDecision).toHaveBeenCalledWith(expect.objectContaining({
      fetcher: expect.any(Function),
      intakeId: "11111111-1111-1111-1111-111111111111",
    }));
  });

  it("does not call the backend mutation when admin UI access is rejected", async () => {
    vi.mocked(assertAdminUiAccess).mockRejectedValue(new Error("blocked"));

    await expect(applySubmissionReviewDecisionFromAdmin({
      intakeId: "11111111-1111-1111-1111-111111111111",
      mainDecision: {
        decision: "approved",
      },
      shortDecisions: [],
    })).rejects.toThrow("blocked");

    expect(applySubmissionReviewDecision).not.toHaveBeenCalled();
  });
});

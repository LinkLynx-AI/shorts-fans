import { applyCreatorReviewDecision } from "@/entities/creator-review";

import { assertAdminUiAccess } from "../../_lib/admin-ui-access";
import { applyCreatorReviewDecisionFromAdmin } from "./actions";

vi.mock("../../_lib/admin-ui-access", () => ({
  assertAdminUiAccess: vi.fn(),
}));

vi.mock("@/entities/creator-review", async () => {
  const actual = await vi.importActual<typeof import("@/entities/creator-review")>("@/entities/creator-review");

  return {
    ...actual,
    applyCreatorReviewDecision: vi.fn(),
  };
});

describe("applyCreatorReviewDecisionFromAdmin", () => {
  beforeEach(() => {
    vi.mocked(assertAdminUiAccess).mockReset();
    vi.mocked(applyCreatorReviewDecision).mockReset();
  });

  it("checks admin UI access before applying the decision", async () => {
    vi.mocked(applyCreatorReviewDecision).mockResolvedValue({} as Awaited<ReturnType<typeof applyCreatorReviewDecision>>);

    await applyCreatorReviewDecisionFromAdmin({
      decision: "approved",
      isResubmitEligible: false,
      isSupportReviewRequired: false,
      reasonCode: "",
      userId: "11111111-1111-1111-1111-111111111111",
    });

    expect(assertAdminUiAccess).toHaveBeenCalledTimes(1);
    expect(applyCreatorReviewDecision).toHaveBeenCalledWith(expect.objectContaining({
      decision: "approved",
      fetcher: expect.any(Function),
      userId: "11111111-1111-1111-1111-111111111111",
    }));
  });

  it("does not call the backend mutation when admin UI access is rejected", async () => {
    vi.mocked(assertAdminUiAccess).mockRejectedValue(new Error("blocked"));

    await expect(applyCreatorReviewDecisionFromAdmin({
      decision: "approved",
      isResubmitEligible: false,
      isSupportReviewRequired: false,
      reasonCode: "",
      userId: "11111111-1111-1111-1111-111111111111",
    })).rejects.toThrow("blocked");

    expect(applyCreatorReviewDecision).not.toHaveBeenCalled();
  });
});

import { isSubmissionReviewIntakeId } from "./contracts";

describe("submission review contracts", () => {
  it("shares intake id validation across admin surfaces", () => {
    expect(isSubmissionReviewIntakeId("11111111-1111-1111-1111-111111111111")).toBe(true);
    expect(isSubmissionReviewIntakeId("not-a-uuid")).toBe(false);
  });
});

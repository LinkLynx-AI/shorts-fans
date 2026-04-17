import { isCreatorReviewUserId } from "./contracts";

describe("creator review contracts", () => {
  it("shares user id validation across admin review surfaces", () => {
    expect(isCreatorReviewUserId("11111111-1111-1111-1111-111111111111")).toBe(true);
    expect(isCreatorReviewUserId("not-a-uuid")).toBe(false);
  });
});

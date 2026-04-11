import {
  createMockSessionProof,
  issueMockSignedToken,
  readMockSignedToken,
  verifyMockSignedToken,
} from "@/shared/lib/mock-signed-token";

describe("mock signed token", () => {
  it("verifies a token for the same context", () => {
    const token = issueMockSignedToken("main_aoi_blue_balcony::softlight", {
      nowMs: 1_000,
      ttlMs: 10_000,
    });

    expect(
      verifyMockSignedToken("main_aoi_blue_balcony::softlight", token, {
        nowMs: 5_000,
      }),
    ).toBe(true);
  });

  it("rejects a token for a different context", () => {
    const token = issueMockSignedToken("main_aoi_blue_balcony::softlight", {
      nowMs: 1_000,
      ttlMs: 10_000,
    });

    expect(
      verifyMockSignedToken("main_aoi_blue_balcony::balcony", token, {
        nowMs: 5_000,
      }),
    ).toBe(false);
  });

  it("rejects an expired token", () => {
    const token = issueMockSignedToken("main_aoi_blue_balcony::softlight", {
      nowMs: 1_000,
      ttlMs: 10,
    });

    expect(
      verifyMockSignedToken("main_aoi_blue_balcony::softlight", token, {
        nowMs: 2_000,
      }),
    ).toBe(false);
  });

  it("returns the payload for a valid token", () => {
    const token = issueMockSignedToken("main_aoi_blue_balcony::softlight", {
      nowMs: 1_000,
      ttlMs: 10_000,
    });

    expect(
      readMockSignedToken(token, {
        nowMs: 5_000,
      }),
    ).toMatchObject({
      context: "main_aoi_blue_balcony::softlight",
    });
  });

  it("verifies a token only for the bound session proof", () => {
    const sessionProof = createMockSessionProof("viewer-session");
    const token = issueMockSignedToken("main_aoi_blue_balcony::softlight", {
      nowMs: 1_000,
      sessionProof,
      ttlMs: 10_000,
    });

    expect(
      verifyMockSignedToken("main_aoi_blue_balcony::softlight", token, {
        nowMs: 5_000,
        sessionProof,
      }),
    ).toBe(true);
    expect(
      verifyMockSignedToken("main_aoi_blue_balcony::softlight", token, {
        nowMs: 5_000,
        sessionProof: createMockSessionProof("other-session"),
      }),
    ).toBe(false);
  });
});

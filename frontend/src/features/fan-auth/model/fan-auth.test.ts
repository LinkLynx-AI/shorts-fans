import {
  buildFanLoginHref,
  isAuthRequiredResponse,
} from "@/features/fan-auth";

describe("fan auth helpers", () => {
  it("returns the fan login entry path", () => {
    expect(buildFanLoginHref()).toBe("/login");
  });

  it("recognizes auth_required payloads", () => {
    expect(
      isAuthRequiredResponse({
        error: {
          code: "auth_required",
          message: "login required",
        },
      }),
    ).toBe(true);
    expect(
      isAuthRequiredResponse({
        error: {
          code: "not_found",
          message: "missing",
        },
      }),
    ).toBe(false);
  });
});

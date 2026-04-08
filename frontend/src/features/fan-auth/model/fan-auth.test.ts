import {
  buildFanLoginHref,
  getFanAuthErrorMessage,
  isAuthRequiredApiError,
  isAuthRequiredResponse,
} from "@/features/fan-auth";
import { ApiError } from "@/shared/api";

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

  it("recognizes auth_required api errors", () => {
    expect(
      isAuthRequiredApiError(
        new ApiError("unauthorized", {
          code: "http",
          details: JSON.stringify({
            error: {
              code: "auth_required",
              message: "login required",
            },
          }),
          status: 401,
        }),
      ),
    ).toBe(true);

    expect(
      isAuthRequiredApiError(
        new ApiError("forbidden", {
          code: "http",
          details: JSON.stringify({
            error: {
              code: "auth_required",
              message: "login required",
            },
          }),
          status: 403,
        }),
      ),
    ).toBe(false);
  });

  it("maps fan auth contract errors to UI copy", () => {
    expect(getFanAuthErrorMessage("invalid_email")).toBe("メールアドレスの形式を確認してください。");
  });
});

import {
  buildFanLoginHref,
  getFanAuthErrorMessage,
  getFanAuthSubmitLabel,
  getFanLogoutErrorMessage,
  isAuthRequiredApiError,
  isAuthRequiredResponse,
  isFreshAuthRequiredApiError,
  isFreshAuthRequiredResponse,
  mapFanAuthNextStepToMode,
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

  it("recognizes fresh_auth_required payloads", () => {
    expect(
      isFreshAuthRequiredResponse({
        error: {
          code: "fresh_auth_required",
          message: "recent auth required",
        },
      }),
    ).toBe(true);
    expect(
      isFreshAuthRequiredResponse({
        error: {
          code: "auth_required",
          message: "login required",
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

  it("recognizes fresh_auth_required api errors", () => {
    expect(
      isFreshAuthRequiredApiError(
        new ApiError("forbidden", {
          code: "http",
          details: JSON.stringify({
            error: {
              code: "fresh_auth_required",
              message: "recent auth required",
            },
          }),
          status: 403,
        }),
      ),
    ).toBe(true);
    expect(
      isFreshAuthRequiredApiError(
        new ApiError("unauthorized", {
          code: "http",
          details: JSON.stringify({
            error: {
              code: "fresh_auth_required",
              message: "recent auth required",
            },
          }),
          status: 401,
        }),
      ),
    ).toBe(false);
  });

  it("maps accepted next steps into modal modes", () => {
    expect(mapFanAuthNextStepToMode("confirm_sign_up")).toBe("confirm-sign-up");
    expect(mapFanAuthNextStepToMode("confirm_password_reset")).toBe("confirm-password-reset");
  });

  it("maps fan auth contract errors to UI copy", () => {
    expect(getFanAuthErrorMessage("confirmation_required")).toBe(
      "確認コードを入力して登録を完了してください。",
    );
    expect(getFanAuthErrorMessage("fresh_auth_required")).toBe(
      "続けるには、もう一度パスワードを入力して認証を確認してください。",
    );
  });

  it("returns mode-aware submit labels", () => {
    expect(getFanAuthSubmitLabel("sign-up", false)).toBe("確認コードを送る");
    expect(getFanAuthSubmitLabel("confirm-sign-up", true)).toBe("登録中...");
  });

  it("maps API logout failures to network-oriented copy", () => {
    expect(
      getFanLogoutErrorMessage(
        new ApiError("API request failed before a response was received.", {
          code: "network",
        }),
      ),
    ).toBe("ログアウトできませんでした。通信状態を確認してから再度お試しください。");
  });

  it("maps unexpected logout failures to fallback copy", () => {
    expect(getFanLogoutErrorMessage(new Error("boom"))).toBe(
      "ログアウトできませんでした。少し時間を置いてからやり直してください。",
    );
  });
});

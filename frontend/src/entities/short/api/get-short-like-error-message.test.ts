import { ApiError } from "@/shared/api";

import { getShortLikeErrorMessage } from "./get-short-like-error-message";
import { ShortLikeApiError } from "./update-short-like";

describe("getShortLikeErrorMessage", () => {
  it("returns a not-found message for unavailable shorts", () => {
    expect(getShortLikeErrorMessage(new ShortLikeApiError("not_found", "missing"))).toBe(
      "この short は現在利用できません。",
    );
  });

  it("returns a network message for transport failures", () => {
    expect(
      getShortLikeErrorMessage(
        new ApiError("network failed", {
          code: "network",
        }),
      ),
    ).toBe("like 状態を更新できませんでした。通信状態を確認してから再度お試しください。");
  });

  it("returns a retry message for generic failures", () => {
    expect(getShortLikeErrorMessage(new Error("unknown"))).toBe(
      "like 状態を更新できませんでした。少し時間を置いてから再度お試しください。",
    );
  });
});

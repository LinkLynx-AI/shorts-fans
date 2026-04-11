import { ApiError } from "@/shared/api";

import { getShortPinErrorMessage } from "./get-short-pin-error-message";
import { ShortPinApiError } from "./update-short-pin";

describe("getShortPinErrorMessage", () => {
  it("returns a not-found message for unavailable shorts", () => {
    expect(getShortPinErrorMessage(new ShortPinApiError("not_found", "missing"))).toBe(
      "この short は現在利用できません。",
    );
  });

  it("returns a network message for transport failures", () => {
    expect(
      getShortPinErrorMessage(
        new ApiError("network failed", {
          code: "network",
        }),
      ),
    ).toBe("pin 状態を更新できませんでした。通信状態を確認してから再度お試しください。");
  });

  it("returns a retry message for generic failures", () => {
    expect(getShortPinErrorMessage(new Error("unknown"))).toBe(
      "pin 状態を更新できませんでした。少し時間を置いてから再度お試しください。",
    );
  });
});

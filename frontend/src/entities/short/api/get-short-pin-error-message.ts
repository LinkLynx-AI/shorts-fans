import { ShortPinApiError } from "./update-short-pin";

const pinUpdateRetryMessage =
  "pin 状態を更新できませんでした。少し時間を置いてから再度お試しください。";
const pinUpdateNetworkMessage =
  "pin 状態を更新できませんでした。通信状態を確認してから再度お試しください。";
const shortNotAvailableMessage = "この short は現在利用できません。";

function getPinErrorMessageFromApiError(error: { code: string }): string {
  if (error.code === "network") {
    return pinUpdateNetworkMessage;
  }

  return pinUpdateRetryMessage;
}

function isCodeError(error: unknown): error is { code: string } {
  return typeof error === "object" && error !== null && "code" in error && typeof error.code === "string";
}

/**
 * short pin mutation の error を UI 表示文言に変換する。
 */
export function getShortPinErrorMessage(error: unknown): string {
  if (error instanceof ShortPinApiError) {
    if (error.code === "not_found") {
      return shortNotAvailableMessage;
    }

    return getPinErrorMessageFromApiError(error);
  }

  if (isCodeError(error)) {
    return getPinErrorMessageFromApiError(error);
  }

  return pinUpdateRetryMessage;
}

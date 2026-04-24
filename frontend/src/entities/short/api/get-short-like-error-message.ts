import { ShortLikeApiError } from "./update-short-like";

const likeUpdateRetryMessage =
  "like 状態を更新できませんでした。少し時間を置いてから再度お試しください。";
const likeUpdateNetworkMessage =
  "like 状態を更新できませんでした。通信状態を確認してから再度お試しください。";
const shortNotAvailableMessage = "この short は現在利用できません。";

function getLikeErrorMessageFromApiError(error: { code: string }): string {
  if (error.code === "network") {
    return likeUpdateNetworkMessage;
  }

  return likeUpdateRetryMessage;
}

function isCodeError(error: unknown): error is { code: string } {
  return typeof error === "object" && error !== null && "code" in error && typeof error.code === "string";
}

/**
 * short like mutation の error を UI 表示文言に変換する。
 */
export function getShortLikeErrorMessage(error: unknown): string {
  if (error instanceof ShortLikeApiError) {
    if (error.code === "not_found") {
      return shortNotAvailableMessage;
    }

    return getLikeErrorMessageFromApiError(error);
  }

  if (isCodeError(error)) {
    return getLikeErrorMessageFromApiError(error);
  }

  return likeUpdateRetryMessage;
}

import { sendRecommendationSignal } from "../api/send-recommendation-signal";

export const recommendationShortEventKinds = [
  "impression",
  "view_start",
  "view_completion",
  "rewatch_loop",
  "main_click",
] as const;

export type RecommendationShortEventKind = (typeof recommendationShortEventKinds)[number];
export type RecommendationEventKind = RecommendationShortEventKind | "profile_click";

export type RecommendationSignalInput =
  | {
      eventKind: RecommendationShortEventKind;
      idempotencyKey: string;
      shortId: string;
    }
  | {
      creatorId: string;
      eventKind: "profile_click";
      idempotencyKey: string;
    };

const publicCreatorIdPattern = /^creator_[0-9a-f]{32}$/;
const publicShortIdPattern = /^short_[0-9a-f]{32}$/;

export function isRecommendationPublicCreatorId(value: string): boolean {
  return publicCreatorIdPattern.test(value.trim().toLowerCase());
}

export function isRecommendationPublicShortId(value: string): boolean {
  return publicShortIdPattern.test(value.trim().toLowerCase());
}

export function createRecommendationSignalNonce(): string {
  const randomUUID = globalThis.crypto?.randomUUID?.();

  if (randomUUID) {
    return randomUUID;
  }

  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}

export function createRecommendationSignalIdempotencyKey(...parts: string[]): string {
  return parts
    .map((part) => part.trim())
    .filter((part) => part.length > 0)
    .join(":");
}

/**
 * UI を止めずに recommendation signal を送る。
 */
export function fireRecommendationSignal(input: RecommendationSignalInput): void {
  void sendRecommendationSignal({ input }).catch(() => undefined);
}

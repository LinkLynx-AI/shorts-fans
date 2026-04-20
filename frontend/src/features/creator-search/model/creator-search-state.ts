import type { CreatorSummary } from "@/entities/creator";

export type CreatorSearchState =
  | {
      items: readonly [];
      kind: "empty";
      query: string;
    }
  | {
      items: readonly [];
      kind: "error";
      message: string;
      query: string;
    }
  | {
      items: readonly [];
      kind: "loading";
      query: string;
    }
  | {
      items: readonly CreatorSummary[];
      kind: "ready";
      query: string;
    };

export const creatorSearchErrorMessage = "検索結果を読み込めませんでした。もう一度お試しください。";

/**
 * creator search query を描画用に正規化する。
 */
export function normalizeCreatorSearchQuery(query: string): string {
  return query.trim();
}

/**
 * creator search の empty state を組み立てる。
 */
export function buildEmptyCreatorSearchState(query: string): CreatorSearchState {
  return {
    items: [],
    kind: "empty",
    query,
  };
}

/**
 * creator search の loading state を組み立てる。
 */
export function buildLoadingCreatorSearchState(query: string): CreatorSearchState {
  return {
    items: [],
    kind: "loading",
    query,
  };
}

/**
 * creator search の error state を組み立てる。
 */
export function buildErrorCreatorSearchState(query: string): CreatorSearchState {
  return {
    items: [],
    kind: "error",
    message: creatorSearchErrorMessage,
    query,
  };
}

/**
 * creator search の ready state を組み立てる。
 */
export function buildReadyCreatorSearchState(
  query: string,
  items: readonly CreatorSummary[],
): CreatorSearchState {
  return {
    items,
    kind: "ready",
    query,
  };
}

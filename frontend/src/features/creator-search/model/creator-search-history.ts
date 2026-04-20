import { z } from "zod";

import { creatorSummarySchema, type CreatorSummary } from "@/entities/creator";

import { normalizeCreatorSearchQuery } from "./creator-search-state";

export const creatorSearchHistoryStorageKeyPrefix = "shorts_fans.creator_search_history.v1";
export const creatorSearchHistoryLimit = 10;
export const creatorSearchHistoryUpdatedEventName = "shorts_fans:creator_search_history_updated";
export const creatorSearchHistoryPendingNavigationKey = "shorts_fans.creator_search_history_pending.v1";

const creatorSearchHistorySchema = creatorSummarySchema.array();
const creatorSearchHistoryPendingNavigationSchema = z.object({
  id: z.string().min(1),
  query: z.string(),
  recordedAtMs: z.number().int().nonnegative(),
});
const creatorSearchHistoryPendingNavigationMaxAgeMs = 5 * 60 * 1000;

export type CreatorSearchHistoryScope = {
  storageKey: string;
  storageKind: "local" | "session";
};

type CreatorSearchHistoryOptions = {
  scope: CreatorSearchHistoryScope | null;
  storage?: Storage;
};

type CreatorSearchHistoryPendingNavigation = {
  creatorId: string;
  query: string;
};

/**
 * viewer 状態に応じた creator search history の保存先を解決する。
 */
export function createCreatorSearchHistoryScope({
  hasViewerSession,
  viewerId,
}: {
  hasViewerSession: boolean;
  viewerId: string | null | undefined;
}): CreatorSearchHistoryScope | null {
  const normalizedViewerId = viewerId?.trim().toLowerCase();

  if (hasViewerSession) {
    if (!normalizedViewerId) {
      return null;
    }

    return {
      storageKey: `${creatorSearchHistoryStorageKeyPrefix}:viewer:${normalizedViewerId}`,
      storageKind: "session",
    };
  }

  return {
    storageKey: `${creatorSearchHistoryStorageKeyPrefix}:guest`,
    storageKind: "session",
  };
}

/**
 * search result からの遷移意図を同一 tab 内で一時記録する。
 */
export function markPendingCreatorSearchHistorySelection(
  navigation: CreatorSearchHistoryPendingNavigation,
  storage?: Storage,
): void {
  const resolvedStorage = resolveSessionStorage(storage);

  if (!resolvedStorage) {
    return;
  }

  try {
    resolvedStorage.setItem(
      creatorSearchHistoryPendingNavigationKey,
      JSON.stringify({
        id: navigation.creatorId,
        query: normalizeCreatorSearchQuery(navigation.query),
        recordedAtMs: Date.now(),
      }),
    );
  } catch {
    return;
  }
}

/**
 * search result からの遷移 marker を検証して一度だけ消費する。
 */
export function consumePendingCreatorSearchHistorySelection(
  navigation: CreatorSearchHistoryPendingNavigation,
  storage?: Storage,
): boolean {
  const resolvedStorage = resolveSessionStorage(storage);

  if (!resolvedStorage) {
    return false;
  }

  try {
    const rawValue = resolvedStorage.getItem(creatorSearchHistoryPendingNavigationKey);

    if (!rawValue) {
      return false;
    }

    const parsedValue = JSON.parse(rawValue) as unknown;
    const result = creatorSearchHistoryPendingNavigationSchema.safeParse(parsedValue);

    if (!result.success) {
      resolvedStorage.removeItem(creatorSearchHistoryPendingNavigationKey);
      return false;
    }

    const isExpired = Date.now() - result.data.recordedAtMs > creatorSearchHistoryPendingNavigationMaxAgeMs;
    const isMatch =
      result.data.id === navigation.creatorId &&
      result.data.query === normalizeCreatorSearchQuery(navigation.query);

    resolvedStorage.removeItem(creatorSearchHistoryPendingNavigationKey);

    if (isExpired) {
      return false;
    }

    return isMatch;
  } catch {
    return false;
  }
}

function resolveStorage(options: CreatorSearchHistoryOptions): {
  storage: Storage;
  storageKey: string;
} | null {
  if (!options.scope) {
    return null;
  }

  if (options.storage) {
    return {
      storage: options.storage,
      storageKey: options.scope.storageKey,
    };
  }

  if (typeof window === "undefined") {
    return null;
  }

  try {
    return {
      storage: options.scope.storageKind === "local" ? window.localStorage : window.sessionStorage,
      storageKey: options.scope.storageKey,
    };
  } catch {
    return null;
  }
}

function resolveSessionStorage(storage?: Storage): Storage | null {
  if (storage) {
    return storage;
  }

  if (typeof window === "undefined") {
    return null;
  }

  try {
    return window.sessionStorage;
  } catch {
    return null;
  }
}

function dispatchCreatorSearchHistoryUpdated(): void {
  if (typeof window === "undefined") {
    return;
  }

  window.dispatchEvent(new Event(creatorSearchHistoryUpdatedEventName));
}

function parseCreatorSearchHistory(rawValue: string | null): readonly CreatorSummary[] {
  if (!rawValue) {
    return [];
  }

  try {
    const parsedValue = JSON.parse(rawValue) as unknown;
    const result = creatorSearchHistorySchema.safeParse(parsedValue);

    if (!result.success) {
      return [];
    }

    return result.data.slice(0, creatorSearchHistoryLimit);
  } catch {
    return [];
  }
}

/**
 * creator search history snapshot を副作用なしで parse する。
 */
export function parseCreatorSearchHistorySnapshot(rawValue: string): readonly CreatorSummary[] {
  return parseCreatorSearchHistory(rawValue);
}

/**
 * creator search history の raw snapshot を取得する。
 */
export function readCreatorSearchHistorySnapshot(options: CreatorSearchHistoryOptions): string {
  const resolvedStorage = resolveStorage(options);

  if (!resolvedStorage) {
    return "";
  }

  try {
    return resolvedStorage.storage.getItem(resolvedStorage.storageKey) ?? "";
  } catch {
    return "";
  }
}

/**
 * creator search history を browser storage から読み出す。
 */
export function readCreatorSearchHistory(options: CreatorSearchHistoryOptions): readonly CreatorSummary[] {
  const resolvedStorage = resolveStorage(options);

  if (!resolvedStorage) {
    return [];
  }

  try {
    const rawValue = readCreatorSearchHistorySnapshot(options);
    const history = parseCreatorSearchHistory(rawValue);

    if (rawValue && history.length === 0) {
      resolvedStorage.storage.removeItem(resolvedStorage.storageKey);
      dispatchCreatorSearchHistoryUpdated();
    }

    return history;
  } catch {
    return [];
  }
}

function buildNextCreatorSearchHistory(
  currentHistory: readonly CreatorSummary[],
  creator: CreatorSummary,
): readonly CreatorSummary[] {
  return [creator, ...currentHistory.filter((historyCreator) => historyCreator.id !== creator.id)].slice(
    0,
    creatorSearchHistoryLimit,
  );
}

/**
 * creator search history に creator を記録する。
 */
export function recordCreatorSearchHistory(
  creator: CreatorSummary,
  options: CreatorSearchHistoryOptions,
): readonly CreatorSummary[] {
  const resolvedStorage = resolveStorage(options);
  const nextHistory = buildNextCreatorSearchHistory(readCreatorSearchHistory(options), creator);

  if (!resolvedStorage) {
    return nextHistory;
  }

  try {
    resolvedStorage.storage.setItem(resolvedStorage.storageKey, JSON.stringify(nextHistory));
    dispatchCreatorSearchHistoryUpdated();
  } catch {
    return nextHistory;
  }

  return nextHistory;
}

/**
 * creator search history 更新を購読する。
 */
export function subscribeCreatorSearchHistory(
  scope: CreatorSearchHistoryScope | null,
  onStoreChange: () => void,
): () => void {
  if (!scope) {
    return () => undefined;
  }

  if (typeof window === "undefined") {
    return () => undefined;
  }

  const handleStorage = (event: StorageEvent) => {
    if (event.key !== null && event.key !== scope.storageKey) {
      return;
    }

    onStoreChange();
  };

  window.addEventListener("storage", handleStorage);
  window.addEventListener(creatorSearchHistoryUpdatedEventName, onStoreChange);

  return () => {
    window.removeEventListener("storage", handleStorage);
    window.removeEventListener(creatorSearchHistoryUpdatedEventName, onStoreChange);
  };
}

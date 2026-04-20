"use client";

import { useEffect, useRef, useState, useSyncExternalStore } from "react";

import type { CreatorSummary } from "@/entities/creator";
import { useCurrentViewer, useHasViewerSession } from "@/entities/viewer";

import {
  buildEmptyCreatorSearchState,
  buildLoadingCreatorSearchState,
  buildReadyCreatorSearchState,
  normalizeCreatorSearchQuery,
  type CreatorSearchState,
} from "./creator-search-state";
import {
  createCreatorSearchHistoryScope,
  markPendingCreatorSearchHistorySelection,
  readCreatorSearchHistory,
  parseCreatorSearchHistorySnapshot,
  readCreatorSearchHistorySnapshot,
  subscribeCreatorSearchHistory,
} from "./creator-search-history";
import { loadCreatorSearchState } from "./load-creator-search-state";
import { useDebouncedValue } from "./use-debounced-value";

type UseCreatorSearchOptions = {
  initialQuery: string;
  initialState: CreatorSearchState;
};

const searchDelayMs = 250;
const subscribeHydrationStatus: (onStoreChange: () => void) => () => void = () => () => undefined;
const readDisabledHistorySnapshot = () => "";

function buildHistoryCreatorSearchState(history: readonly CreatorSummary[]): CreatorSearchState {
  if (history.length === 0) {
    return buildEmptyCreatorSearchState("");
  }

  return buildReadyCreatorSearchState("", history);
}

/**
 * creator search panel 用の query / async state を管理する。
 */
export function useCreatorSearch({
  initialQuery,
  initialState,
}: UseCreatorSearchOptions) {
  const [query, setQuery] = useState(initialQuery);
  const [retryCount, setRetryCount] = useState(0);
  const [state, setState] = useState(initialState);
  const isFirstLoadRef = useRef(true);
  const resolvedQuery = useDebouncedValue(query, searchDelayMs);
  const currentViewer = useCurrentViewer();
  const hasViewerSession = useHasViewerSession();
  const usesHistory = normalizeCreatorSearchQuery(query).length === 0;
  const historyScope = createCreatorSearchHistoryScope({
    hasViewerSession,
    viewerId: currentViewer?.id,
  });
  const isHydrated = useSyncExternalStore(subscribeHydrationStatus, () => true, () => false);
  const historySnapshot = useSyncExternalStore(
    (onStoreChange) =>
      usesHistory ? subscribeCreatorSearchHistory(historyScope, onStoreChange) : subscribeHydrationStatus(onStoreChange),
    () => (usesHistory ? readCreatorSearchHistorySnapshot({ scope: historyScope }) : readDisabledHistorySnapshot()),
    () => "",
  );
  const historyState = buildHistoryCreatorSearchState(parseCreatorSearchHistorySnapshot(historySnapshot));

  function setLoadingForQuery(nextQuery: string) {
    setState(buildLoadingCreatorSearchState(normalizeCreatorSearchQuery(nextQuery)));
  }

  useEffect(() => {
    if (!usesHistory || !isHydrated) {
      return;
    }

    readCreatorSearchHistory({
      scope: historyScope,
    });
  }, [historyScope, historySnapshot, isHydrated, usesHistory]);

  useEffect(() => {
    if (isFirstLoadRef.current) {
      isFirstLoadRef.current = false;
      if (
        normalizeCreatorSearchQuery(resolvedQuery).length > 0 &&
        normalizeCreatorSearchQuery(resolvedQuery) === normalizeCreatorSearchQuery(initialState.query)
      ) {
        return;
      }
    }

    const normalizedQuery = normalizeCreatorSearchQuery(resolvedQuery);

    if (normalizedQuery.length === 0) {
      return;
    }

    const controller = new AbortController();

    void loadCreatorSearchState(normalizedQuery, {
      signal: controller.signal,
    }).then((nextState) => {
      if (controller.signal.aborted) {
        return;
      }
      setState(nextState);
    });

    return () => {
      controller.abort();
    };
  }, [initialState.query, resolvedQuery, retryCount]);

  return {
    isHistoryHydrating: usesHistory && !isHydrated,
    markCreatorSelectionPending: (creator: CreatorSummary, selectedQuery: string) => {
      if (!historyScope) {
        return;
      }

      markPendingCreatorSearchHistorySelection({
        creatorId: creator.id,
        query: selectedQuery,
      });
    },
    query,
    retry: () => {
      if (normalizeCreatorSearchQuery(query).length === 0) {
        return;
      }

      setLoadingForQuery(query);
      setRetryCount((currentCount) => currentCount + 1);
    },
    setQuery: (nextQuery: string) => {
      if (normalizeCreatorSearchQuery(nextQuery).length > 0) {
        setLoadingForQuery(nextQuery);
      }
      setQuery(nextQuery);
    },
    state: usesHistory ? (isHydrated ? historyState : initialState) : state,
  };
}

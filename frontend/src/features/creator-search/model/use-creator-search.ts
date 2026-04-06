"use client";

import { useEffect, useRef, useState } from "react";

import {
  buildLoadingCreatorSearchState,
  normalizeCreatorSearchQuery,
  type CreatorSearchState,
} from "./creator-search-state";
import { loadCreatorSearchState } from "./load-creator-search-state";
import { useDebouncedValue } from "./use-debounced-value";

type UseCreatorSearchOptions = {
  initialQuery: string;
  initialState: CreatorSearchState;
};

const searchDelayMs = 250;

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

  function setLoadingForQuery(nextQuery: string) {
    setState(buildLoadingCreatorSearchState(normalizeCreatorSearchQuery(nextQuery)));
  }

  useEffect(() => {
    if (isFirstLoadRef.current) {
      isFirstLoadRef.current = false;
      if (normalizeCreatorSearchQuery(resolvedQuery) === normalizeCreatorSearchQuery(initialState.query)) {
        return;
      }
    }

    const controller = new AbortController();
    const normalizedQuery = normalizeCreatorSearchQuery(resolvedQuery);

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
    query,
    retry: () => {
      setLoadingForQuery(query);
      setRetryCount((currentCount) => currentCount + 1);
    },
    setQuery: (nextQuery: string) => {
      setLoadingForQuery(nextQuery);
      setQuery(nextQuery);
    },
    state,
  };
}

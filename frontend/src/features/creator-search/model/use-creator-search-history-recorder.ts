"use client";

import type { CreatorSummary } from "@/entities/creator";
import {
  useCurrentViewer,
  useHasViewerSession,
} from "@/entities/viewer";

import {
  consumePendingCreatorSearchHistorySelection,
  createCreatorSearchHistoryScope,
  recordCreatorSearchHistory,
} from "./creator-search-history";

/**
 * search 起点の creator profile 遷移完了を履歴に確定する。
 */
export function useCreatorSearchHistoryRecorder() {
  const currentViewer = useCurrentViewer();
  const hasViewerSession = useHasViewerSession();
  const historyScope = createCreatorSearchHistoryScope({
    hasViewerSession,
    viewerId: currentViewer?.id,
  });

  return {
    recordVisitedCreatorFromSearch: ({
      creator,
      query,
    }: {
      creator: CreatorSummary;
      query: string | undefined;
    }): boolean => {
      if (!historyScope) {
        return false;
      }

      const canRecord = consumePendingCreatorSearchHistorySelection({
        creatorId: creator.id,
        query: query ?? "",
      });

      if (!canRecord) {
        return false;
      }

      recordCreatorSearchHistory(creator, {
        scope: historyScope,
      });

      return true;
    },
  };
}

"use client";

import { useEffect, useState } from "react";

import { resolveCreatorWorkspaceBlockedState } from "./creator-workspace-blocked-state";
import {
  buildErrorCreatorWorkspacePreviewCollectionsState,
  buildLoadingCreatorWorkspacePreviewCollectionsState,
  type CreatorWorkspacePreviewCollectionsState,
} from "./creator-workspace-preview-collections";
import { loadCreatorWorkspacePreviewCollections } from "./load-creator-workspace-preview-collections";
import { type CreatorModeShellBlockedState } from "./creator-mode-shell";

type UseCreatorWorkspacePreviewCollectionsResult = {
  blockedState: CreatorModeShellBlockedState | null;
  retry: () => void;
  state: CreatorWorkspacePreviewCollectionsState;
};

/**
 * creator workspace 下側一覧の取得状態を client で管理する。
 */
export function useCreatorWorkspacePreviewCollections(): UseCreatorWorkspacePreviewCollectionsResult {
  const [blockedState, setBlockedState] = useState<CreatorModeShellBlockedState | null>(null);
  const [retryCount, setRetryCount] = useState(0);
  const [state, setState] = useState<CreatorWorkspacePreviewCollectionsState>(
    buildLoadingCreatorWorkspacePreviewCollectionsState(),
  );

  useEffect(() => {
    const controller = new AbortController();

    void loadCreatorWorkspacePreviewCollections(controller.signal).then((collections) => {
      if (controller.signal.aborted) {
        return;
      }

      setState({
        collections,
        kind: "ready",
      });
    }).catch((error: unknown) => {
      if (controller.signal.aborted) {
        return;
      }

      const nextBlockedState = resolveCreatorWorkspaceBlockedState(error);

      if (nextBlockedState) {
        setBlockedState(nextBlockedState);
        return;
      }

      setState(buildErrorCreatorWorkspacePreviewCollectionsState(error));
    });

    return () => {
      controller.abort();
    };
  }, [retryCount]);

  return {
    blockedState,
    retry: () => {
      setBlockedState(null);
      setState(buildLoadingCreatorWorkspacePreviewCollectionsState());
      setRetryCount((currentCount) => currentCount + 1);
    },
    state,
  };
}

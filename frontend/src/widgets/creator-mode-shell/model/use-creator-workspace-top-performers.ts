"use client";

import { useEffect, useState } from "react";

import { getCreatorWorkspaceTopPerformers } from "../api/get-creator-workspace-top-performers";
import { resolveCreatorWorkspaceBlockedState } from "./creator-workspace-blocked-state";
import {
  buildErrorCreatorWorkspaceTopPerformersState,
  buildLoadingCreatorWorkspaceTopPerformersState,
  type CreatorWorkspaceTopPerformersState,
} from "./creator-workspace-top-performers";
import { type CreatorModeShellBlockedState } from "./creator-mode-shell";

type UseCreatorWorkspaceTopPerformersResult = {
  blockedState: CreatorModeShellBlockedState | null;
  retry: () => void;
  state: CreatorWorkspaceTopPerformersState;
};

/**
 * creator workspace top performers の取得状態を client で管理する。
 */
export function useCreatorWorkspaceTopPerformers(): UseCreatorWorkspaceTopPerformersResult {
  const [blockedState, setBlockedState] = useState<CreatorModeShellBlockedState | null>(null);
  const [retryCount, setRetryCount] = useState(0);
  const [state, setState] = useState<CreatorWorkspaceTopPerformersState>(
    buildLoadingCreatorWorkspaceTopPerformersState(),
  );

  useEffect(() => {
    const controller = new AbortController();

    void getCreatorWorkspaceTopPerformers({
      credentials: "include",
      signal: controller.signal,
    }).then((topPerformers) => {
      if (controller.signal.aborted) {
        return;
      }

      setState({
        kind: "ready",
        topPerformers,
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

      setState(buildErrorCreatorWorkspaceTopPerformersState(error));
    });

    return () => {
      controller.abort();
    };
  }, [retryCount]);

  return {
    blockedState,
    retry: () => {
      setBlockedState(null);
      setState(buildLoadingCreatorWorkspaceTopPerformersState());
      setRetryCount((currentCount) => currentCount + 1);
    },
    state,
  };
}

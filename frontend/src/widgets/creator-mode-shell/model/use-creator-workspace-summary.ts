"use client";

import { useEffect, useState } from "react";

import { ApiError } from "@/shared/api";

import { getCreatorWorkspaceSummary } from "../api/get-creator-workspace-summary";
import {
  buildErrorCreatorWorkspaceSummaryState,
  buildLoadingCreatorWorkspaceSummaryState,
  type CreatorWorkspaceSummaryState,
} from "./creator-workspace-summary";
import {
  getCreatorModeCapabilityRequiredState,
  getCreatorModeUnauthenticatedState,
  type CreatorModeShellBlockedState,
} from "./creator-mode-shell";

type UseCreatorWorkspaceSummaryResult = {
  blockedState: CreatorModeShellBlockedState | null;
  retry: () => void;
  state: CreatorWorkspaceSummaryState;
};

function resolveCreatorWorkspaceSummaryBlockedState(error: unknown): CreatorModeShellBlockedState | null {
  if (!(error instanceof ApiError) || error.code !== "http") {
    return null;
  }

  if (error.status === 401) {
    return getCreatorModeUnauthenticatedState();
  }

  if (error.status === 403) {
    return getCreatorModeCapabilityRequiredState();
  }

  return null;
}

/**
 * creator workspace summary の取得状態を client で管理する。
 */
export function useCreatorWorkspaceSummary(): UseCreatorWorkspaceSummaryResult {
  const [blockedState, setBlockedState] = useState<CreatorModeShellBlockedState | null>(null);
  const [retryCount, setRetryCount] = useState(0);
  const [state, setState] = useState<CreatorWorkspaceSummaryState>(buildLoadingCreatorWorkspaceSummaryState());

  useEffect(() => {
    const controller = new AbortController();

    void getCreatorWorkspaceSummary({
      credentials: "include",
      signal: controller.signal,
    }).then((summary) => {
      if (controller.signal.aborted) {
        return;
      }

      setState({
        kind: "ready",
        summary,
      });
    }).catch((error: unknown) => {
      if (controller.signal.aborted) {
        return;
      }

      const nextBlockedState = resolveCreatorWorkspaceSummaryBlockedState(error);

      if (nextBlockedState) {
        setBlockedState(nextBlockedState);
        return;
      }

      setState(buildErrorCreatorWorkspaceSummaryState(error));
    });

    return () => {
      controller.abort();
    };
  }, [retryCount]);

  return {
    blockedState,
    retry: () => {
      setBlockedState(null);
      setState(buildLoadingCreatorWorkspaceSummaryState());
      setRetryCount((currentCount) => currentCount + 1);
    },
    state,
  };
}

"use client";

import { useRef, useState } from "react";

import {
  getShortLikeErrorMessage,
  ShortLikeApiError,
  updateShortLike,
} from "@/entities/short";
import { useHasViewerSession } from "@/entities/viewer";
import { useFanAuthDialogControls } from "@/features/fan-auth";

type UseShortLikeStateOptions = {
  enabled?: boolean;
  initialHasLiked: boolean;
  initialLikeCount: number;
  shortId: string;
};

type UseShortLikeStateResult = {
  errorMessage: string | null;
  hasLiked: boolean;
  isPending: boolean;
  likeCount: number;
  onToggle: () => void;
};

/**
 * 単一 short detail 用の like pending / success / error state を管理する。
 */
export function useShortLikeState({
  enabled = true,
  initialHasLiked,
  initialLikeCount,
  shortId,
}: UseShortLikeStateOptions): UseShortLikeStateResult {
  const hasViewerSession = useHasViewerSession();
  const { openFanAuthDialog } = useFanAuthDialogControls();
  const isPendingRef = useRef(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [hasLiked, setHasLiked] = useState(initialHasLiked);
  const [isPending, setIsPending] = useState(false);
  const [likeCount, setLikeCount] = useState(initialLikeCount);

  const onToggle = () => {
    if (!enabled || isPendingRef.current) {
      return;
    }

    if (!hasViewerSession) {
      setErrorMessage(null);
      openFanAuthDialog({
        postAuthNavigation: "none",
      });
      return;
    }

    const nextAction = hasLiked ? "unlike" : "like";

    isPendingRef.current = true;
    setErrorMessage(null);
    setIsPending(true);

    void updateShortLike({
      action: nextAction,
      shortId,
    }).then((result) => {
      setHasLiked(result.viewer.hasLiked);
      setLikeCount(result.engagement.likeCount);
    }).catch((error: unknown) => {
      if (error instanceof ShortLikeApiError && error.code === "auth_required") {
        openFanAuthDialog({
          postAuthNavigation: "none",
        });
        return;
      }

      setErrorMessage(getShortLikeErrorMessage(error));
    }).finally(() => {
      isPendingRef.current = false;
      setIsPending(false);
    });
  };

  return {
    errorMessage,
    hasLiked,
    isPending,
    likeCount,
    onToggle,
  };
}

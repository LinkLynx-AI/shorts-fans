"use client";

import { useRef, useState } from "react";
import { useRouter } from "next/navigation";

import {
  getShortPinErrorMessage,
  ShortPinApiError,
  updateShortPin,
} from "@/entities/short";
import { useHasViewerSession } from "@/entities/viewer";
import { buildFanLoginHref } from "@/features/fan-auth";

type UseShortPinStateOptions = {
  enabled?: boolean;
  initialIsPinned: boolean;
  shortId: string;
};

type UseShortPinStateResult = {
  errorMessage: string | null;
  isPending: boolean;
  isPinned: boolean;
  onToggle: () => void;
};

/**
 * 単一 short detail 用の pin pending / success / error state を管理する。
 */
export function useShortPinState({
  enabled = true,
  initialIsPinned,
  shortId,
}: UseShortPinStateOptions): UseShortPinStateResult {
  const hasViewerSession = useHasViewerSession();
  const isPendingRef = useRef(false);
  const router = useRouter();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isPending, setIsPending] = useState(false);
  const [isPinned, setIsPinned] = useState(initialIsPinned);

  const onToggle = () => {
    if (!enabled || isPendingRef.current) {
      return;
    }

    if (!hasViewerSession) {
      setErrorMessage(null);
      router.push(buildFanLoginHref());
      return;
    }

    const nextAction = isPinned ? "unpin" : "pin";

    isPendingRef.current = true;
    setErrorMessage(null);
    setIsPending(true);

    void updateShortPin({
      action: nextAction,
      shortId,
    }).then((result) => {
      setIsPinned(result.viewer.isPinned);
    }).catch((error: unknown) => {
      if (error instanceof ShortPinApiError && error.code === "auth_required") {
        router.push(buildFanLoginHref());
        return;
      }

      setErrorMessage(getShortPinErrorMessage(error));
    }).finally(() => {
      isPendingRef.current = false;
      setIsPending(false);
    });
  };

  return {
    errorMessage,
    isPending,
    isPinned,
    onToggle,
  };
}

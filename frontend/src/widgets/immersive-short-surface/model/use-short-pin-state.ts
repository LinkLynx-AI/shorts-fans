"use client";

import { useRef, useState } from "react";
import { useRouter } from "next/navigation";

import {
  ShortPinApiError,
  updateShortPin,
} from "@/entities/short";
import { useHasViewerSession } from "@/entities/viewer";
import { buildFanLoginHref } from "@/features/fan-auth";
import { ApiError } from "@/shared/api";

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

function getShortPinErrorMessage(error: unknown): string {
  if (error instanceof ShortPinApiError) {
    if (error.code === "not_found") {
      return "この short は現在利用できません。";
    }

    return "pin 状態を更新できませんでした。少し時間を置いてから再度お試しください。";
  }

  if (error instanceof ApiError) {
    if (error.code === "network") {
      return "pin 状態を更新できませんでした。通信状態を確認してから再度お試しください。";
    }

    return "pin 状態を更新できませんでした。少し時間を置いてから再度お試しください。";
  }

  return "pin 状態を更新できませんでした。少し時間を置いてから再度お試しください。";
}

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

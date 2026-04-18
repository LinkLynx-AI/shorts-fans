"use client";

import { useCallback, useLayoutEffect, useRef } from "react";

import {
  createRecommendationSignalIdempotencyKey,
  createRecommendationSignalNonce,
  fireRecommendationSignal,
  isRecommendationPublicCreatorId,
  isRecommendationPublicShortId,
} from "./recommendation-signal";

const completionThreshold = 0.95;
const loopCompletionFloor = 0.9;
const loopRestartCeiling = 0.1;
const manualSeekSuppressWindowMs = 1500;

type UseShortRecommendationSignalsOptions = {
  creatorId: string;
  isActive: boolean;
  isSurfaceReady?: boolean;
  shortId: string;
  viewerId: string | null;
};

type ShortSignalSessionState = {
  activationNonce: string | null;
  completionSent: boolean;
  impressionSent: boolean;
  lastProgress: number;
  lastSeekAt: number;
  loopCount: number;
  viewStartSent: boolean;
};

type ActiveShortSignalSessionState = ShortSignalSessionState & {
  activationNonce: string;
};

function createShortSignalSessionState(): ShortSignalSessionState {
  return {
    activationNonce: null,
    completionSent: false,
    impressionSent: false,
    lastProgress: 0,
    lastSeekAt: 0,
    loopCount: 0,
    viewStartSent: false,
  };
}

function normalizeProgress(currentTime: number, duration: number): number {
  if (!Number.isFinite(currentTime) || !Number.isFinite(duration) || duration <= 0) {
    return 0;
  }

  const progress = currentTime / duration;

  if (progress <= 0) {
    return 0;
  }

  if (progress >= 1) {
    return 1;
  }

  return progress;
}

/**
 * short playback 中の recommendation signal 発火を管理する。
 */
export function useShortRecommendationSignals({
  creatorId,
  isActive,
  isSurfaceReady = true,
  shortId,
  viewerId,
}: UseShortRecommendationSignalsOptions) {
  const sessionStateRef = useRef<ShortSignalSessionState>(createShortSignalSessionState());

  const canRecordShortSignals =
    viewerId !== null && isSurfaceReady && isRecommendationPublicShortId(shortId);
  const canRecordProfileClick =
    viewerId !== null && isSurfaceReady && isRecommendationPublicCreatorId(creatorId);

  const ensureActiveSession = useCallback((): ActiveShortSignalSessionState | null => {
    if (!isActive || !canRecordShortSignals) {
      return null;
    }

    const sessionState = sessionStateRef.current;

    if (sessionState.activationNonce !== null) {
      return sessionState as ActiveShortSignalSessionState;
    }

    sessionState.activationNonce = createRecommendationSignalNonce();
    sessionState.impressionSent = true;

    fireRecommendationSignal({
      eventKind: "impression",
      idempotencyKey: createRecommendationSignalIdempotencyKey(
        "impression",
        shortId,
        sessionState.activationNonce,
      ),
      shortId,
    });

    return sessionState as ActiveShortSignalSessionState;
  }, [canRecordShortSignals, isActive, shortId]);

  useLayoutEffect(() => {
    sessionStateRef.current = createShortSignalSessionState();
    ensureActiveSession();
  }, [canRecordShortSignals, ensureActiveSession, isActive, shortId, viewerId]);

  const markManualSeek = () => {
    sessionStateRef.current.lastSeekAt = Date.now();
  };

  const handleVideoPlay = () => {
    const sessionState = ensureActiveSession();
    if (sessionState === null) {
      return;
    }

    if (sessionState.viewStartSent) {
      return;
    }

    sessionState.viewStartSent = true;

    fireRecommendationSignal({
      eventKind: "view_start",
      idempotencyKey: createRecommendationSignalIdempotencyKey(
        "view_start",
        shortId,
        sessionState.activationNonce,
      ),
      shortId,
    });
  };

  const handleTimeUpdate = (currentTime: number, duration: number) => {
    const sessionState = ensureActiveSession();
    if (sessionState === null) {
      return;
    }

    const nextProgress = normalizeProgress(currentTime, duration);

    if (!sessionState.completionSent && nextProgress >= completionThreshold) {
      sessionState.completionSent = true;

      fireRecommendationSignal({
        eventKind: "view_completion",
        idempotencyKey: createRecommendationSignalIdempotencyKey(
          "view_completion",
          shortId,
          sessionState.activationNonce,
        ),
        shortId,
      });
    }

    const wrappedAround =
      sessionState.lastProgress >= loopCompletionFloor && nextProgress <= loopRestartCeiling;
    const manualSeekSuppressed = Date.now() - sessionState.lastSeekAt < manualSeekSuppressWindowMs;

    if (wrappedAround && !manualSeekSuppressed) {
      sessionState.loopCount += 1;

      fireRecommendationSignal({
        eventKind: "rewatch_loop",
        idempotencyKey: createRecommendationSignalIdempotencyKey(
          "rewatch_loop",
          shortId,
          sessionState.activationNonce,
          String(sessionState.loopCount),
        ),
        shortId,
      });
    }

    sessionState.lastProgress = nextProgress;
  };

  const recordMainClick = () => {
    if (!canRecordShortSignals) {
      return;
    }

    fireRecommendationSignal({
      eventKind: "main_click",
      idempotencyKey: createRecommendationSignalIdempotencyKey(
        "main_click",
        shortId,
        createRecommendationSignalNonce(),
      ),
      shortId,
    });
  };

  const recordProfileClick = () => {
    if (!canRecordProfileClick) {
      return;
    }

    fireRecommendationSignal({
      creatorId,
      eventKind: "profile_click",
      idempotencyKey: createRecommendationSignalIdempotencyKey(
        "profile_click",
        creatorId,
        createRecommendationSignalNonce(),
      ),
    });
  };

  return {
    handleTimeUpdate,
    handleVideoPlay,
    markManualSeek,
    recordMainClick,
    recordProfileClick,
  };
}

"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import {
  createShortComment,
  getShortComments,
  ShortCommentApiError,
  type ShortComment,
} from "@/entities/short";
import { ApiError } from "@/shared/api";

export const maxCommentBodyLength = 500;
export const maxCommentDraftInputLength = maxCommentBodyLength + 100;

type UseShortCommentsOptions = {
  createCommentApi?: typeof createShortComment | undefined;
  fetchComments?: typeof getShortComments | undefined;
  hasViewerSession: boolean;
  onAuthRequired: () => void;
  open: boolean;
  shortId: string;
};

type CommentPageState = {
  hasNext: boolean;
  nextCursor: string | null;
};

function isAbortError(error: unknown): boolean {
  if (error instanceof DOMException && error.name === "AbortError") {
    return true;
  }

  return error instanceof ApiError && error.cause instanceof DOMException && error.cause.name === "AbortError";
}

function getListErrorMessage(error: unknown): string | null {
  if (isAbortError(error)) {
    return null;
  }

  if (error instanceof ShortCommentApiError && error.code === "not_found") {
    return "この short は表示できません。";
  }

  return "コメントを読み込めませんでした。";
}

function getSubmitErrorMessage(error: unknown): string {
  if (error instanceof ShortCommentApiError) {
    if (error.code === "validation_error") {
      return "1〜500文字で入力してください。";
    }

    if (error.code === "not_found") {
      return "この short にはコメントできません。";
    }
  }

  return "コメントを投稿できませんでした。";
}

function mergeCreatedComment(comments: readonly ShortComment[], createdComment: ShortComment): ShortComment[] {
  return [
    createdComment,
    ...comments.filter((comment) => comment.id !== createdComment.id),
  ];
}

function mergeInitialPage(currentComments: readonly ShortComment[], initialComments: readonly ShortComment[]): ShortComment[] {
  if (currentComments.length === 0) {
    return [...initialComments];
  }

  const initialCommentIds = new Set(initialComments.map((comment) => comment.id));

  return [
    ...currentComments.filter((comment) => !initialCommentIds.has(comment.id)),
    ...initialComments,
  ];
}

/**
 * comment body の文字数を上限超過が分かる時点で打ち切って数える。
 */
export function countCommentBodyCharacters(rawBody: string): number {
  const startIndex = findFirstNonWhitespaceIndex(rawBody);

  if (startIndex >= rawBody.length) {
    return 0;
  }

  const endIndex = findLastNonWhitespaceEndIndex(rawBody, startIndex);
  let count = 0;
  let index = startIndex;

  while (index < endIndex) {
    count += 1;

    if (count > maxCommentBodyLength) {
      return count;
    }

    index = getNextCodePointIndex(rawBody, index);
  }

  return count;
}

/**
 * comment draft が極端に大きくならないよう入力 state の長さを制限する。
 */
export function limitCommentDraftInput(rawBody: string): string {
  let count = 0;
  let index = 0;

  while (index < rawBody.length) {
    count += 1;

    if (count > maxCommentDraftInputLength) {
      return rawBody.slice(0, index);
    }

    index = getNextCodePointIndex(rawBody, index);
  }

  return rawBody;
}

function findFirstNonWhitespaceIndex(value: string): number {
  let index = 0;

  while (index < value.length) {
    const nextIndex = getNextCodePointIndex(value, index);
    if (value.slice(index, nextIndex).trim() !== "") {
      return index;
    }

    index = nextIndex;
  }

  return value.length;
}

function findLastNonWhitespaceEndIndex(value: string, minIndex: number): number {
  let index = value.length;

  while (index > minIndex) {
    const previousIndex = getPreviousCodePointIndex(value, index);
    if (value.slice(previousIndex, index).trim() !== "") {
      return index;
    }

    index = previousIndex;
  }

  return minIndex;
}

function getNextCodePointIndex(value: string, index: number): number {
  const codePoint = value.codePointAt(index);

  if (codePoint === undefined) {
    return value.length;
  }

  return index + (codePoint > 0xffff ? 2 : 1);
}

function getPreviousCodePointIndex(value: string, index: number): number {
  const previous = value.charCodeAt(index - 1);

  if (previous >= 0xdc00 && previous <= 0xdfff && index > 1) {
    const beforePrevious = value.charCodeAt(index - 2);
    if (beforePrevious >= 0xd800 && beforePrevious <= 0xdbff) {
      return index - 2;
    }
  }

  return index - 1;
}

/**
 * short comment bottom sheet の取得・投稿状態を管理する。
 */
export function useShortComments({
  createCommentApi = createShortComment,
  fetchComments = getShortComments,
  hasViewerSession,
  onAuthRequired,
  open,
  shortId,
}: UseShortCommentsOptions) {
  const [comments, setComments] = useState<readonly ShortComment[]>([]);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isLoadingInitial, setIsLoadingInitial] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [page, setPage] = useState<CommentPageState>({
    hasNext: false,
    nextCursor: null,
  });
  const [submitErrorMessage, setSubmitErrorMessage] = useState<string | null>(null);
  const loadMoreAbortControllerRef = useRef<AbortController | null>(null);
  const loadMoreInFlightRef = useRef(false);
  const submitAbortControllerRef = useRef<AbortController | null>(null);
  const currentShortIdRef = useRef(shortId);
  const requestKeyRef = useRef(0);
  currentShortIdRef.current = shortId;

  useEffect(() => {
    if (!open) {
      requestKeyRef.current += 1;
      loadMoreAbortControllerRef.current?.abort();
      loadMoreAbortControllerRef.current = null;
      loadMoreInFlightRef.current = false;
      submitAbortControllerRef.current?.abort();
      submitAbortControllerRef.current = null;
      setIsLoadingInitial(false);
      setIsLoadingMore(false);
      setIsSubmitting(false);
      return;
    }

    const abortController = new AbortController();
    const requestKey = requestKeyRef.current + 1;
    requestKeyRef.current = requestKey;
    loadMoreInFlightRef.current = false;
    setComments([]);
    setErrorMessage(null);
    setIsLoadingInitial(true);
    setIsLoadingMore(false);
    setIsSubmitting(false);
    setPage({
      hasNext: false,
      nextCursor: null,
    });
    setSubmitErrorMessage(null);

    void fetchComments({
      shortId,
      signal: abortController.signal,
    })
      .then((nextPage) => {
        if (requestKeyRef.current !== requestKey || currentShortIdRef.current !== shortId) {
          return;
        }

        setComments((currentComments) => mergeInitialPage(currentComments, nextPage.items));
        setPage(nextPage.page);
      })
      .catch((error: unknown) => {
        if (requestKeyRef.current !== requestKey || currentShortIdRef.current !== shortId) {
          return;
        }

        const nextMessage = getListErrorMessage(error);
        if (nextMessage !== null) {
          setErrorMessage(nextMessage);
        }
      })
      .finally(() => {
        if (requestKeyRef.current === requestKey && currentShortIdRef.current === shortId) {
          setIsLoadingInitial(false);
        }
      });

    return () => {
      abortController.abort();
      loadMoreAbortControllerRef.current?.abort();
      loadMoreAbortControllerRef.current = null;
      loadMoreInFlightRef.current = false;
      submitAbortControllerRef.current?.abort();
      submitAbortControllerRef.current = null;
    };
  }, [fetchComments, open, shortId]);

  const loadMore = useCallback(async () => {
    if (isLoadingInitial || isLoadingMore || loadMoreInFlightRef.current || !page.hasNext || !page.nextCursor) {
      return;
    }

    const requestKey = requestKeyRef.current;
    loadMoreAbortControllerRef.current?.abort();
    const abortController = new AbortController();
    loadMoreAbortControllerRef.current = abortController;
    loadMoreInFlightRef.current = true;
    setErrorMessage(null);
    setIsLoadingMore(true);

    try {
      const nextPage = await fetchComments({
        cursor: page.nextCursor,
        signal: abortController.signal,
        shortId,
      });

      if (requestKeyRef.current !== requestKey || currentShortIdRef.current !== shortId) {
        return;
      }

      setComments((currentComments) => [
        ...currentComments,
        ...nextPage.items,
      ]);
      setPage(nextPage.page);
    } catch (error) {
      if (requestKeyRef.current !== requestKey || currentShortIdRef.current !== shortId) {
        return;
      }

      const nextMessage = getListErrorMessage(error);
      if (nextMessage !== null) {
        setErrorMessage(nextMessage);
      }
    } finally {
      if (loadMoreAbortControllerRef.current === abortController) {
        loadMoreAbortControllerRef.current = null;
        loadMoreInFlightRef.current = false;
      }

      if (requestKeyRef.current === requestKey && currentShortIdRef.current === shortId) {
        setIsLoadingMore(false);
      }
    }
  }, [fetchComments, isLoadingInitial, isLoadingMore, page.hasNext, page.nextCursor, shortId]);

  const submitComment = useCallback(async (rawBody: string): Promise<boolean> => {
    const bodyLength = countCommentBodyCharacters(rawBody);

    setSubmitErrorMessage(null);

    if (bodyLength === 0 || bodyLength > maxCommentBodyLength) {
      setSubmitErrorMessage("1〜500文字で入力してください。");
      return false;
    }

    if (!hasViewerSession) {
      onAuthRequired();
      return false;
    }

    const body = rawBody.trim();
    const requestKey = requestKeyRef.current;
    submitAbortControllerRef.current?.abort();
    const abortController = new AbortController();
    submitAbortControllerRef.current = abortController;
    setIsSubmitting(true);

    try {
      const createdComment = await createCommentApi({
        body,
        signal: abortController.signal,
        shortId,
      });

      if (
        requestKeyRef.current !== requestKey
        || submitAbortControllerRef.current !== abortController
        || currentShortIdRef.current !== shortId
      ) {
        return false;
      }

      setComments((currentComments) => mergeCreatedComment(currentComments, createdComment));
      setErrorMessage(null);
      return true;
    } catch (error) {
      if (
        isAbortError(error)
        || requestKeyRef.current !== requestKey
        || submitAbortControllerRef.current !== abortController
        || currentShortIdRef.current !== shortId
      ) {
        return false;
      }

      if (error instanceof ShortCommentApiError && error.code === "auth_required") {
        onAuthRequired();
        return false;
      }

      setSubmitErrorMessage(getSubmitErrorMessage(error));
      return false;
    } finally {
      if (submitAbortControllerRef.current === abortController) {
        submitAbortControllerRef.current = null;
      }

      if (requestKeyRef.current === requestKey && currentShortIdRef.current === shortId) {
        setIsSubmitting(false);
      }
    }
  }, [createCommentApi, hasViewerSession, onAuthRequired, shortId]);

  return {
    comments,
    errorMessage,
    hasNextPage: page.hasNext,
    isLoadingInitial,
    isLoadingMore,
    isSubmitting,
    loadMore,
    submitComment,
    submitErrorMessage,
  };
}

import { ApiError, createApiUrl } from "@/shared/api";
import { getClientEnv } from "@/shared/config";

import {
  shortCommentCreateResponseSchema,
  type ShortComment,
} from "./contracts";
import { parseShortCommentError } from "./short-comment-error";

export { ShortCommentApiError } from "./short-comment-error";
export type { ShortCommentApiErrorCode } from "./short-comment-error";

type CreateShortCommentOptions = {
  baseUrl?: string | undefined;
  body: string;
  credentials?: RequestCredentials | undefined;
  fetcher?: typeof fetch | undefined;
  shortId: string;
  signal?: AbortSignal | undefined;
};

function buildShortCommentPath(shortId: string): `/${string}` {
  return `/api/fan/shorts/${encodeURIComponent(shortId)}/comments`;
}

function getShortCommentBaseUrl(baseUrl?: string): string {
  return baseUrl ?? getClientEnv().NEXT_PUBLIC_API_BASE_URL;
}

/**
 * public short に comment を投稿する。
 */
export async function createShortComment({
  baseUrl,
  body,
  credentials = "include",
  fetcher = fetch,
  shortId,
  signal,
}: CreateShortCommentOptions): Promise<ShortComment> {
  let response: Response;

  try {
    response = await fetcher(createApiUrl(getShortCommentBaseUrl(baseUrl), buildShortCommentPath(shortId)), {
      body: JSON.stringify({ body }),
      credentials,
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      method: "POST",
      ...(signal ? { signal } : {}),
    });
  } catch (error) {
    throw new ApiError("API request failed before a response was received.", {
      cause: error,
      code: "network",
    });
  }

  if (!response.ok) {
    throw await parseShortCommentError(response);
  }

  let payload: unknown;

  try {
    payload = await response.json();
  } catch (error) {
    throw new ApiError("API response body was not valid JSON.", {
      cause: error,
      code: "parse",
      status: response.status,
    });
  }

  const parsed = shortCommentCreateResponseSchema.safeParse(payload);

  if (!parsed.success) {
    throw new ApiError("API response body did not match the expected schema.", {
      cause: parsed.error,
      code: "parse",
      details: parsed.error.message,
      status: response.status,
    });
  }

  return parsed.data.data.comment;
}

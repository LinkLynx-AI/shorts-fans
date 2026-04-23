import { ApiError, createApiUrl } from "@/shared/api";
import { getClientEnv } from "@/shared/config";

import { shortCommentsResponseSchema, type ShortComment } from "./contracts";
import { parseShortCommentError } from "./short-comment-error";

export type ShortCommentsPage = {
  items: readonly ShortComment[];
  page: {
    hasNext: boolean;
    nextCursor: string | null;
  };
  requestId: string;
};

type GetShortCommentsOptions = {
  baseUrl?: string | undefined;
  credentials?: RequestCredentials | undefined;
  cursor?: string | null | undefined;
  fetcher?: typeof fetch | undefined;
  shortId: string;
  signal?: AbortSignal | undefined;
};

function buildShortCommentsPath(shortId: string, cursor?: string | null): `/${string}` {
  const trimmedCursor = cursor?.trim();
  const encodedShortID = encodeURIComponent(shortId);

  if (!trimmedCursor) {
    return `/api/fan/shorts/${encodedShortID}/comments`;
  }

  const searchParams = new URLSearchParams({
    cursor: trimmedCursor,
  });

  return `/api/fan/shorts/${encodedShortID}/comments?${searchParams.toString()}` as `/${string}`;
}

function getShortCommentsBaseUrl(baseUrl?: string): string {
  return baseUrl ?? getClientEnv().NEXT_PUBLIC_API_BASE_URL;
}

/**
 * public short comment の 1 page を取得する。
 */
export async function getShortComments({
  baseUrl,
  credentials = "include",
  cursor,
  fetcher,
  shortId,
  signal,
}: GetShortCommentsOptions): Promise<ShortCommentsPage> {
  let response: Response;
  const resolvedFetcher = fetcher ?? fetch;
  const requestUrl = createApiUrl(getShortCommentsBaseUrl(baseUrl), buildShortCommentsPath(shortId, cursor));

  try {
    response = await resolvedFetcher(requestUrl, {
      cache: "no-store",
      credentials,
      headers: {
        Accept: "application/json",
      },
      method: "GET",
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

  const parsed = shortCommentsResponseSchema.safeParse(payload);

  if (!parsed.success) {
    throw new ApiError("API response body did not match the expected schema.", {
      cause: parsed.error,
      code: "parse",
      details: parsed.error.message,
      status: response.status,
    });
  }

  return {
    items: parsed.data.data.items,
    page: parsed.data.meta.page,
    requestId: parsed.data.meta.requestId,
  };
}

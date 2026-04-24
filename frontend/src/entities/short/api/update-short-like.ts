import { z } from "zod";

import { ApiError, createApiUrl } from "@/shared/api";
import { getClientEnv } from "@/shared/config";

import {
  shortLikeErrorCodeSchema,
  shortLikeMutationResultSchema,
} from "./contracts";

const shortLikeSuccessResponseSchema = z.object({
  data: shortLikeMutationResultSchema,
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

const shortLikeErrorResponseSchema = z.object({
  data: z.null(),
  error: z.object({
    code: shortLikeErrorCodeSchema,
    message: z.string().min(1),
  }),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type ShortLikeAction = "like" | "unlike";
export type ShortLikeApiErrorCode = z.infer<typeof shortLikeErrorCodeSchema>;
export type ShortLikeMutationResult = z.output<typeof shortLikeMutationResultSchema>;

type ShortLikeApiErrorOptions = {
  requestId?: string;
  status?: number;
};

type UpdateShortLikeOptions = {
  action: ShortLikeAction;
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
  shortId: string;
};

/**
 * short like mutation の contract error を表す。
 */
export class ShortLikeApiError extends Error {
  readonly code: ShortLikeApiErrorCode;
  readonly requestId: string | undefined;
  readonly status: number | undefined;

  constructor(code: ShortLikeApiErrorCode, message: string, options: ShortLikeApiErrorOptions = {}) {
    super(message);
    this.name = "ShortLikeApiError";
    this.code = code;
    this.requestId = options.requestId;
    this.status = options.status;
  }
}

function buildShortLikePath(shortId: string): `/${string}` {
  return `/api/fan/shorts/${encodeURIComponent(shortId)}/like`;
}

function getShortLikeBaseUrl(baseUrl?: string): string {
  return baseUrl ?? getClientEnv().NEXT_PUBLIC_API_BASE_URL;
}

function resolveShortLikeMethod(action: ShortLikeAction): "DELETE" | "PUT" {
  return action === "like" ? "PUT" : "DELETE";
}

async function parseShortLikeError(response: Response): Promise<ApiError | ShortLikeApiError> {
  let payload: unknown;

  try {
    payload = await response.json();
  } catch (error) {
    return new ApiError("API response body was not valid JSON.", {
      cause: error,
      code: "parse",
      status: response.status,
    });
  }

  const parsed = shortLikeErrorResponseSchema.safeParse(payload);

  if (!parsed.success) {
    return new ApiError("API response body did not match the expected schema.", {
      cause: parsed.error,
      code: "parse",
      details: parsed.error.message,
      status: response.status,
    });
  }

  return new ShortLikeApiError(parsed.data.error.code, parsed.data.error.message, {
    requestId: parsed.data.meta.requestId,
    status: response.status,
  });
}

/**
 * feed/detail short の like / unlike mutation を実行する。
 */
export async function updateShortLike({
  action,
  baseUrl,
  credentials = "include",
  fetcher = fetch,
  shortId,
}: UpdateShortLikeOptions): Promise<ShortLikeMutationResult> {
  let response: Response;

  try {
    response = await fetcher(createApiUrl(getShortLikeBaseUrl(baseUrl), buildShortLikePath(shortId)), {
      credentials,
      headers: {
        Accept: "application/json",
      },
      method: resolveShortLikeMethod(action),
    });
  } catch (error) {
    throw new ApiError("API request failed before a response was received.", {
      cause: error,
      code: "network",
    });
  }

  if (!response.ok) {
    throw await parseShortLikeError(response);
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

  const parsed = shortLikeSuccessResponseSchema.safeParse(payload);

  if (!parsed.success) {
    throw new ApiError("API response body did not match the expected schema.", {
      cause: parsed.error,
      code: "parse",
      details: parsed.error.message,
      status: response.status,
    });
  }

  return parsed.data.data;
}

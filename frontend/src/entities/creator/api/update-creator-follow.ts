import { z } from "zod";

import { ApiError, createApiUrl } from "@/shared/api";
import { getClientEnv } from "@/shared/config";

import {
  creatorFollowErrorCodeSchema,
  creatorFollowMutationResultSchema,
} from "./contracts";

const creatorFollowSuccessResponseSchema = z.object({
  data: creatorFollowMutationResultSchema,
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

const creatorFollowErrorResponseSchema = z.object({
  data: z.null(),
  error: z.object({
    code: creatorFollowErrorCodeSchema,
    message: z.string().min(1),
  }),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type CreatorFollowAction = "follow" | "unfollow";
export type CreatorFollowApiErrorCode = z.infer<typeof creatorFollowErrorCodeSchema>;
export type CreatorFollowMutationResult = z.output<typeof creatorFollowMutationResultSchema>;

type CreatorFollowApiErrorOptions = {
  requestId?: string;
  status?: number;
};

type UpdateCreatorFollowOptions = {
  action: CreatorFollowAction;
  baseUrl?: string;
  creatorId: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
};

/**
 * creator follow mutation の contract error を表す。
 */
export class CreatorFollowApiError extends Error {
  readonly code: CreatorFollowApiErrorCode;
  readonly requestId: string | undefined;
  readonly status: number | undefined;

  constructor(code: CreatorFollowApiErrorCode, message: string, options: CreatorFollowApiErrorOptions = {}) {
    super(message);
    this.name = "CreatorFollowApiError";
    this.code = code;
    this.requestId = options.requestId;
    this.status = options.status;
  }
}

function buildCreatorFollowPath(creatorId: string): `/${string}` {
  return `/api/fan/creators/${encodeURIComponent(creatorId)}/follow`;
}

function getCreatorFollowBaseUrl(baseUrl?: string): string {
  return baseUrl ?? getClientEnv().NEXT_PUBLIC_API_BASE_URL;
}

function resolveCreatorFollowMethod(action: CreatorFollowAction): "DELETE" | "PUT" {
  return action === "follow" ? "PUT" : "DELETE";
}

async function parseCreatorFollowError(response: Response): Promise<ApiError | CreatorFollowApiError> {
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

  const parsed = creatorFollowErrorResponseSchema.safeParse(payload);

  if (!parsed.success) {
    return new ApiError("API response body did not match the expected schema.", {
      cause: parsed.error,
      code: "parse",
      details: parsed.error.message,
      status: response.status,
    });
  }

  return new CreatorFollowApiError(parsed.data.error.code, parsed.data.error.message, {
    requestId: parsed.data.meta.requestId,
    status: response.status,
  });
}

/**
 * creator profile follow / unfollow mutation を実行する。
 *
 * Contract:
 * - `PUT / DELETE /api/fan/creators/{creatorId}/follow` を呼ぶ
 * - success では post-condition の `viewer.isFollowing` と `stats.fanCount` を返す
 * - browser cookie を送るため `credentials: "include"` を既定にする
 *
 * Errors:
 * - contract error は `CreatorFollowApiError`
 * - schema / network / malformed response は `ApiError`
 *
 * Side effects:
 * - backend creator follow mutation endpoint へ network request を送る
 */
export async function updateCreatorFollow({
  action,
  baseUrl,
  creatorId,
  credentials = "include",
  fetcher = fetch,
}: UpdateCreatorFollowOptions): Promise<CreatorFollowMutationResult> {
  let response: Response;

  try {
    response = await fetcher(createApiUrl(getCreatorFollowBaseUrl(baseUrl), buildCreatorFollowPath(creatorId)), {
      credentials,
      headers: {
        Accept: "application/json",
      },
      method: resolveCreatorFollowMethod(action),
    });
  } catch (error) {
    throw new ApiError("API request failed before a response was received.", {
      cause: error,
      code: "network",
    });
  }

  if (!response.ok) {
    throw await parseCreatorFollowError(response);
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

  const parsed = creatorFollowSuccessResponseSchema.safeParse(payload);

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

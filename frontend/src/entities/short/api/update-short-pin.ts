import { z } from "zod";

import { ApiError, createApiUrl } from "@/shared/api";
import { getClientEnv } from "@/shared/config";

import {
  shortPinErrorCodeSchema,
  shortPinMutationResultSchema,
} from "./contracts";

const shortPinSuccessResponseSchema = z.object({
  data: shortPinMutationResultSchema,
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

const shortPinErrorResponseSchema = z.object({
  data: z.null(),
  error: z.object({
    code: shortPinErrorCodeSchema,
    message: z.string().min(1),
  }),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type ShortPinAction = "pin" | "unpin";
export type ShortPinApiErrorCode = z.infer<typeof shortPinErrorCodeSchema>;
export type ShortPinMutationResult = z.output<typeof shortPinMutationResultSchema>;

type ShortPinApiErrorOptions = {
  requestId?: string;
  status?: number;
};

type UpdateShortPinOptions = {
  action: ShortPinAction;
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
  shortId: string;
};

/**
 * short pin mutation の contract error を表す。
 */
export class ShortPinApiError extends Error {
  readonly code: ShortPinApiErrorCode;
  readonly requestId: string | undefined;
  readonly status: number | undefined;

  constructor(code: ShortPinApiErrorCode, message: string, options: ShortPinApiErrorOptions = {}) {
    super(message);
    this.name = "ShortPinApiError";
    this.code = code;
    this.requestId = options.requestId;
    this.status = options.status;
  }
}

function buildShortPinPath(shortId: string): `/${string}` {
  return `/api/fan/shorts/${encodeURIComponent(shortId)}/pin`;
}

function getShortPinBaseUrl(baseUrl?: string): string {
  return baseUrl ?? getClientEnv().NEXT_PUBLIC_API_BASE_URL;
}

function resolveShortPinMethod(action: ShortPinAction): "DELETE" | "PUT" {
  return action === "pin" ? "PUT" : "DELETE";
}

async function parseShortPinError(response: Response): Promise<ApiError | ShortPinApiError> {
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

  const parsed = shortPinErrorResponseSchema.safeParse(payload);

  if (!parsed.success) {
    return new ApiError("API response body did not match the expected schema.", {
      cause: parsed.error,
      code: "parse",
      details: parsed.error.message,
      status: response.status,
    });
  }

  return new ShortPinApiError(parsed.data.error.code, parsed.data.error.message, {
    requestId: parsed.data.meta.requestId,
    status: response.status,
  });
}

/**
 * feed short の pin / unpin mutation を実行する。
 */
export async function updateShortPin({
  action,
  baseUrl,
  credentials = "include",
  fetcher = fetch,
  shortId,
}: UpdateShortPinOptions): Promise<ShortPinMutationResult> {
  let response: Response;

  try {
    response = await fetcher(createApiUrl(getShortPinBaseUrl(baseUrl), buildShortPinPath(shortId)), {
      credentials,
      headers: {
        Accept: "application/json",
      },
      method: resolveShortPinMethod(action),
    });
  } catch (error) {
    throw new ApiError("API request failed before a response was received.", {
      cause: error,
      code: "network",
    });
  }

  if (!response.ok) {
    throw await parseShortPinError(response);
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

  const parsed = shortPinSuccessResponseSchema.safeParse(payload);

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

import { z } from "zod";

import { ApiError, createApiUrl } from "@/shared/api";
import { getClientEnv } from "@/shared/config";

import {
  creatorWorkspaceMainPriceErrorCodeSchema,
  creatorWorkspaceMainPriceMutationResultSchema,
} from "./contracts";

const creatorWorkspaceMainPriceSuccessResponseSchema = z.object({
  data: creatorWorkspaceMainPriceMutationResultSchema,
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

const creatorWorkspaceMainPriceErrorResponseSchema = z.object({
  data: z.null(),
  error: z.object({
    code: creatorWorkspaceMainPriceErrorCodeSchema,
    message: z.string().min(1),
  }),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type CreatorWorkspaceMainPriceApiErrorCode = z.infer<typeof creatorWorkspaceMainPriceErrorCodeSchema>;
export type CreatorWorkspaceMainPriceMutationResult = z.output<typeof creatorWorkspaceMainPriceMutationResultSchema>;

type CreatorWorkspaceMainPriceApiErrorOptions = {
  requestId?: string;
  status?: number;
};

type UpdateCreatorWorkspaceMainPriceOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
  mainId: string;
  priceJpy: number;
};

/**
 * creator workspace main price mutation の contract error を表す。
 */
export class CreatorWorkspaceMainPriceApiError extends Error {
  readonly code: CreatorWorkspaceMainPriceApiErrorCode;
  readonly requestId: string | undefined;
  readonly status: number | undefined;

  constructor(code: CreatorWorkspaceMainPriceApiErrorCode, message: string, options: CreatorWorkspaceMainPriceApiErrorOptions = {}) {
    super(message);
    this.name = "CreatorWorkspaceMainPriceApiError";
    this.code = code;
    this.requestId = options.requestId;
    this.status = options.status;
  }
}

function buildCreatorWorkspaceMainPricePath(mainId: string): `/${string}` {
  return `/api/creator/workspace/mains/${encodeURIComponent(mainId.trim())}/price`;
}

function getCreatorWorkspaceMainPriceBaseUrl(baseUrl?: string): string {
  return baseUrl ?? getClientEnv().NEXT_PUBLIC_API_BASE_URL;
}

async function parseCreatorWorkspaceMainPriceError(response: Response): Promise<ApiError | CreatorWorkspaceMainPriceApiError> {
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

  const parsed = creatorWorkspaceMainPriceErrorResponseSchema.safeParse(payload);

  if (!parsed.success) {
    return new ApiError("API response body did not match the expected schema.", {
      cause: parsed.error,
      code: "parse",
      details: parsed.error.message,
      status: response.status,
    });
  }

  return new CreatorWorkspaceMainPriceApiError(parsed.data.error.code, parsed.data.error.message, {
    requestId: parsed.data.meta.requestId,
    status: response.status,
  });
}

/**
 * creator workspace owner preview 上の本編価格を更新する。
 */
export async function updateCreatorWorkspaceMainPrice({
  baseUrl,
  credentials = "include",
  fetcher = fetch,
  mainId,
  priceJpy,
}: UpdateCreatorWorkspaceMainPriceOptions): Promise<CreatorWorkspaceMainPriceMutationResult> {
  let response: Response;

  try {
    response = await fetcher(createApiUrl(getCreatorWorkspaceMainPriceBaseUrl(baseUrl), buildCreatorWorkspaceMainPricePath(mainId)), {
      body: JSON.stringify({
        priceJpy,
      }),
      credentials,
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      method: "PUT",
    });
  } catch (error) {
    throw new ApiError("API request failed before a response was received.", {
      cause: error,
      code: "network",
    });
  }

  if (!response.ok) {
    throw await parseCreatorWorkspaceMainPriceError(response);
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

  const parsed = creatorWorkspaceMainPriceSuccessResponseSchema.safeParse(payload);

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

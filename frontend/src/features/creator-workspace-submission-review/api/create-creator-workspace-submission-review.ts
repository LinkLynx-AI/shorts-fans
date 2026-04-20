import { z } from "zod";

import { ApiError, createApiUrl } from "@/shared/api";
import { getClientEnv } from "@/shared/config";

const creatorWorkspaceSubmissionReviewErrorCodeSchema = z.enum([
  "auth_required",
  "creator_mode_unavailable",
  "not_found",
  "review_state_conflict",
  "submission_not_ready",
  "internal_error",
]);

const creatorWorkspaceSubmissionReviewErrorResponseSchema = z.object({
  data: z.null(),
  error: z.object({
    code: creatorWorkspaceSubmissionReviewErrorCodeSchema,
    message: z.string().min(1),
  }),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type CreatorWorkspaceSubmissionReviewApiErrorCode = z.infer<typeof creatorWorkspaceSubmissionReviewErrorCodeSchema>;

type CreatorWorkspaceSubmissionReviewApiErrorOptions = {
  requestId?: string;
  status?: number;
};

type CreateCreatorWorkspaceSubmissionReviewOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
  mainId: string;
};

/**
 * creator workspace submission review mutation の contract error を表す。
 */
export class CreatorWorkspaceSubmissionReviewApiError extends Error {
  readonly code: CreatorWorkspaceSubmissionReviewApiErrorCode;
  readonly requestId: string | undefined;
  readonly status: number | undefined;

  constructor(
    code: CreatorWorkspaceSubmissionReviewApiErrorCode,
    message: string,
    options: CreatorWorkspaceSubmissionReviewApiErrorOptions = {},
  ) {
    super(message);
    this.name = "CreatorWorkspaceSubmissionReviewApiError";
    this.code = code;
    this.requestId = options.requestId;
    this.status = options.status;
  }
}

function buildCreatorWorkspaceSubmissionReviewPath(mainId: string): `/${string}` {
  return `/api/creator/workspace/mains/${encodeURIComponent(mainId.trim())}/review-submissions`;
}

function getCreatorWorkspaceSubmissionReviewBaseUrl(baseUrl?: string): string {
  return baseUrl ?? getClientEnv().NEXT_PUBLIC_API_BASE_URL;
}

async function parseCreatorWorkspaceSubmissionReviewError(
  response: Response,
): Promise<ApiError | CreatorWorkspaceSubmissionReviewApiError> {
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

  const parsed = creatorWorkspaceSubmissionReviewErrorResponseSchema.safeParse(payload);

  if (!parsed.success) {
    return new ApiError("API response body did not match the expected schema.", {
      cause: parsed.error,
      code: "parse",
      details: parsed.error.message,
      status: response.status,
    });
  }

  return new CreatorWorkspaceSubmissionReviewApiError(
    parsed.data.error.code,
    parsed.data.error.message,
    {
      requestId: parsed.data.meta.requestId,
      status: response.status,
    },
  );
}

/**
 * creator workspace から current package を submit / resubmit する。
 */
export async function createCreatorWorkspaceSubmissionReview({
  baseUrl,
  credentials = "include",
  fetcher = fetch,
  mainId,
}: CreateCreatorWorkspaceSubmissionReviewOptions): Promise<void> {
  let response: Response;

  try {
    response = await fetcher(
      createApiUrl(
        getCreatorWorkspaceSubmissionReviewBaseUrl(baseUrl),
        buildCreatorWorkspaceSubmissionReviewPath(mainId),
      ),
      {
        credentials,
        headers: {
          Accept: "application/json",
        },
        method: "POST",
      },
    );
  } catch (error) {
    throw new ApiError("API request failed before a response was received.", {
      cause: error,
      code: "network",
    });
  }

  if (!response.ok) {
    throw await parseCreatorWorkspaceSubmissionReviewError(response);
  }
}

import { z } from "zod";

import { ApiError } from "@/shared/api";

import {
  shortCommentErrorCodeSchema,
  shortCommentErrorResponseSchema,
} from "./contracts";

export type ShortCommentApiErrorCode = z.infer<typeof shortCommentErrorCodeSchema>;

type ShortCommentApiErrorOptions = {
  requestId?: string;
  status?: number;
};

/**
 * short comment API の contract error を表す。
 */
export class ShortCommentApiError extends Error {
  readonly code: ShortCommentApiErrorCode;
  readonly requestId: string | undefined;
  readonly status: number | undefined;

  constructor(code: ShortCommentApiErrorCode, message: string, options: ShortCommentApiErrorOptions = {}) {
    super(message);
    this.name = "ShortCommentApiError";
    this.code = code;
    this.requestId = options.requestId;
    this.status = options.status;
  }
}

/**
 * short comment API の error response を contract error へ変換する。
 */
export async function parseShortCommentError(response: Response): Promise<ApiError | ShortCommentApiError> {
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

  const parsed = shortCommentErrorResponseSchema.safeParse(payload);

  if (!parsed.success) {
    return new ApiError("API response body did not match the expected schema.", {
      cause: parsed.error,
      code: "parse",
      details: parsed.error.message,
      status: response.status,
    });
  }

  return new ShortCommentApiError(parsed.data.error.code, parsed.data.error.message, {
    requestId: parsed.data.meta.requestId,
    status: response.status,
  });
}

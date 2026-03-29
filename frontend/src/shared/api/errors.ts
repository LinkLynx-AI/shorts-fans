export type ApiErrorCode = "http" | "network" | "parse";

type ApiErrorOptions = {
  cause?: unknown;
  code: ApiErrorCode;
  details?: string;
  status?: number;
};

/**
 * API 境界で扱う失敗を正規化する。
 */
export class ApiError extends Error {
  readonly code: ApiErrorCode;
  readonly details: string | undefined;
  readonly status: number | undefined;

  constructor(message: string, options: ApiErrorOptions) {
    super(message, { cause: options.cause });
    this.name = "ApiError";
    this.code = options.code;
    this.details = options.details;
    this.status = options.status;
  }
}

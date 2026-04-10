import { ApiError, createApiUrl } from "@/shared/api";
import { getClientEnv } from "@/shared/config";

import {
  creatorUploadCompleteResponseSchema,
  creatorUploadCreateResponseSchema,
  creatorUploadErrorResponseSchema,
  type CreatorUploadCompleteResponse,
  type CreatorUploadCreateResponse,
  type CreatorUploadErrorCode,
  type CreatorUploadTarget,
} from "./contracts";

type CreatorUploadApiErrorOptions = {
  requestId?: string;
  status?: number;
};

type CreateCreatorUploadPackageOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
  mainFile: File;
  shortFiles: readonly File[];
};

type CompleteCreatorUploadPackageOptions = {
  baseUrl?: string;
  consentConfirmed: boolean;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
  mainUploadEntryId: string;
  ownershipConfirmed: boolean;
  packageToken: string;
  priceJpy: number;
  shorts: readonly {
    caption: string | null;
    uploadEntryId: string;
  }[];
};

type UploadCreatorUploadTargetOptions = {
  fetcher?: typeof fetch;
  file: File;
  target: CreatorUploadTarget;
};

/**
 * creator upload endpoint の contract error を表す。
 */
export class CreatorUploadApiError extends Error {
  readonly code: CreatorUploadErrorCode;
  readonly requestId: string | undefined;
  readonly status: number | undefined;

  constructor(code: CreatorUploadErrorCode, message: string, options: CreatorUploadApiErrorOptions = {}) {
    super(message);
    this.name = "CreatorUploadApiError";
    this.code = code;
    this.requestId = options.requestId;
    this.status = options.status;
  }
}

function getCreatorUploadBaseUrl(baseUrl?: string): string {
  return baseUrl ?? getClientEnv().NEXT_PUBLIC_API_BASE_URL;
}

async function parseCreatorUploadError(response: Response): Promise<ApiError | CreatorUploadApiError> {
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

  const parsed = creatorUploadErrorResponseSchema.safeParse(payload);

  if (!parsed.success) {
    return new ApiError("API response body did not match the expected schema.", {
      cause: parsed.error,
      code: "parse",
      details: parsed.error.message,
      status: response.status,
    });
  }

  return new CreatorUploadApiError(parsed.data.error.code, parsed.data.error.message, {
    requestId: parsed.data.meta.requestId,
    status: response.status,
  });
}

function buildCreatorUploadFileMetadata(file: File) {
  return {
    fileName: file.name,
    fileSizeBytes: file.size,
    mimeType: file.type,
  };
}

async function parseCreatorUploadSuccess<TResponse extends CreatorUploadCreateResponse | CreatorUploadCompleteResponse>(
  response: Response,
  schema: { safeParse: (payload: unknown) => { success: true; data: TResponse } | { success: false; error: Error } },
): Promise<TResponse> {
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

  const parsed = schema.safeParse(payload);

  if (!parsed.success) {
    throw new ApiError("API response body did not match the expected schema.", {
      cause: parsed.error,
      code: "parse",
      details: parsed.error.message,
      status: response.status,
    });
  }

  return parsed.data;
}

/**
 * creator upload package 作成 endpoint を呼ぶ。
 */
export async function createCreatorUploadPackage({
  baseUrl,
  credentials = "include",
  fetcher = fetch,
  mainFile,
  shortFiles,
}: CreateCreatorUploadPackageOptions): Promise<CreatorUploadCreateResponse["data"]> {
  let response: Response;

  try {
    response = await fetcher(createApiUrl(getCreatorUploadBaseUrl(baseUrl), "/api/creator/upload-packages"), {
      body: JSON.stringify({
        main: buildCreatorUploadFileMetadata(mainFile),
        shorts: shortFiles.map((file) => buildCreatorUploadFileMetadata(file)),
      }),
      credentials,
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      method: "POST",
    });
  } catch (error) {
    throw new ApiError("API request failed before a response was received.", {
      cause: error,
      code: "network",
    });
  }

  if (!response.ok) {
    throw await parseCreatorUploadError(response);
  }

  const parsed = await parseCreatorUploadSuccess(response, creatorUploadCreateResponseSchema);
  return parsed.data;
}

/**
 * creator upload completion endpoint を呼ぶ。
 */
export async function completeCreatorUploadPackage({
  baseUrl,
  consentConfirmed,
  credentials = "include",
  fetcher = fetch,
  mainUploadEntryId,
  ownershipConfirmed,
  packageToken,
  priceJpy,
  shorts,
}: CompleteCreatorUploadPackageOptions): Promise<CreatorUploadCompleteResponse["data"]> {
  let response: Response;

  try {
    response = await fetcher(createApiUrl(getCreatorUploadBaseUrl(baseUrl), "/api/creator/upload-packages/complete"), {
      body: JSON.stringify({
        main: {
          consentConfirmed,
          ownershipConfirmed,
          priceJpy,
          uploadEntryId: mainUploadEntryId,
        },
        packageToken,
        shorts: shorts.map((short) => ({
          caption: short.caption,
          uploadEntryId: short.uploadEntryId,
        })),
      }),
      credentials,
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      method: "POST",
    });
  } catch (error) {
    throw new ApiError("API request failed before a response was received.", {
      cause: error,
      code: "network",
    });
  }

  if (!response.ok) {
    throw await parseCreatorUploadError(response);
  }

  const parsed = await parseCreatorUploadSuccess(response, creatorUploadCompleteResponseSchema);
  return parsed.data;
}

/**
 * presigned target へ creator upload file を直接送る。
 */
export async function uploadCreatorUploadTarget({
  fetcher = fetch,
  file,
  target,
}: UploadCreatorUploadTargetOptions): Promise<void> {
  let response: Response;

  try {
    response = await fetcher(target.upload.url, {
      body: file,
      headers: target.upload.headers,
      method: target.upload.method,
    });
  } catch (error) {
    throw new ApiError("Direct upload request failed before a response was received.", {
      cause: error,
      code: "network",
    });
  }

  if (!response.ok) {
    const details = await response.text().catch(() => response.statusText);

    throw new ApiError("Direct upload request failed with a non-success status.", {
      code: "http",
      details,
      status: response.status,
    });
  }
}

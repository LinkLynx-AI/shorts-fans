import { ApiError, requestJson } from "@/shared/api";
import { getClientEnv } from "@/shared/config";

import {
  creatorRegistrationAvatarCompleteResponseSchema,
  creatorRegistrationAvatarCreateResponseSchema,
  type CreatorRegistrationAvatarCompleteResponse,
  type CreatorRegistrationAvatarCreateResponse,
  type CreatorRegistrationAvatarUploadTarget,
} from "./avatar-upload-contracts";

type CreatorRegistrationAvatarUploadOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
};

type UploadCreatorRegistrationAvatarTargetOptions = {
  fetcher?: typeof fetch;
  file: File;
  target: CreatorRegistrationAvatarUploadTarget;
};

function getCreatorRegistrationAvatarBaseURL(baseUrl?: string): string {
  return baseUrl ?? getClientEnv().NEXT_PUBLIC_API_BASE_URL;
}

function buildCreatorRegistrationAvatarFileMetadata(file: File) {
  return {
    fileName: file.name,
    fileSizeBytes: file.size,
    mimeType: file.type,
  };
}

/**
 * creator registration avatar upload target を要求する。
 */
export async function createCreatorRegistrationAvatarUpload(
  file: File,
  {
    baseUrl,
    credentials = "include",
    fetcher = fetch,
  }: CreatorRegistrationAvatarUploadOptions = {},
): Promise<CreatorRegistrationAvatarCreateResponse["data"]> {
  const response = await requestJson({
    baseUrl: getCreatorRegistrationAvatarBaseURL(baseUrl),
    fetcher,
    init: {
      body: JSON.stringify(buildCreatorRegistrationAvatarFileMetadata(file)),
      credentials,
      headers: {
        "Content-Type": "application/json",
      },
      method: "POST",
    },
    path: "/api/viewer/creator-registration/avatar-uploads",
    schema: creatorRegistrationAvatarCreateResponseSchema,
  });

  return response.data;
}

/**
 * uploaded avatar を registration で使える completed token に変換する。
 */
export async function completeCreatorRegistrationAvatarUpload(
  avatarUploadToken: string,
  {
    baseUrl,
    credentials = "include",
    fetcher = fetch,
  }: CreatorRegistrationAvatarUploadOptions = {},
): Promise<CreatorRegistrationAvatarCompleteResponse["data"]> {
  const response = await requestJson({
    baseUrl: getCreatorRegistrationAvatarBaseURL(baseUrl),
    fetcher,
    init: {
      body: JSON.stringify({
        avatarUploadToken,
      }),
      credentials,
      headers: {
        "Content-Type": "application/json",
      },
      method: "POST",
    },
    path: "/api/viewer/creator-registration/avatar-uploads/complete",
    schema: creatorRegistrationAvatarCompleteResponseSchema,
  });

  return response.data;
}

/**
 * presigned target へ avatar file を直接送る。
 */
export async function uploadCreatorRegistrationAvatarTarget({
  fetcher = fetch,
  file,
  target,
}: UploadCreatorRegistrationAvatarTargetOptions): Promise<void> {
  let response: Response;

  try {
    response = await fetcher(new URL(target.upload.url), {
      body: file,
      headers: target.upload.headers,
      method: target.upload.method,
    });
  } catch (error) {
    throw new ApiError("Direct avatar upload request failed before a response was received.", {
      cause: error,
      code: "network",
    });
  }

  if (!response.ok) {
    const details = await response.text().catch(() => response.statusText);

    throw new ApiError("Direct avatar upload request failed with a non-success status.", {
      code: "http",
      details,
      status: response.status,
    });
  }
}

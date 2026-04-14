import { ApiError, requestJson } from "@/shared/api";
import { getClientEnv } from "@/shared/config";

import {
  viewerProfileAvatarCompleteResponseSchema,
  viewerProfileAvatarCreateResponseSchema,
  type ViewerProfileAvatarCompleteResponse,
  type ViewerProfileAvatarCreateResponse,
  type ViewerProfileAvatarUploadTarget,
} from "./contracts";

type ViewerProfileAvatarUploadOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
};

type UploadViewerProfileAvatarTargetOptions = {
  fetcher?: typeof fetch;
  file: File;
  target: ViewerProfileAvatarUploadTarget;
};

function getViewerProfileBaseURL(baseUrl?: string): string {
  return baseUrl ?? getClientEnv().NEXT_PUBLIC_API_BASE_URL;
}

function buildViewerProfileAvatarFileMetadata(file: File) {
  return {
    fileName: file.name,
    fileSizeBytes: file.size,
    mimeType: file.type,
  };
}

export async function createViewerProfileAvatarUpload(
  file: File,
  {
    baseUrl,
    credentials = "include",
    fetcher = fetch,
  }: ViewerProfileAvatarUploadOptions = {},
): Promise<ViewerProfileAvatarCreateResponse["data"]> {
  const response = await requestJson({
    baseUrl: getViewerProfileBaseURL(baseUrl),
    fetcher,
    init: {
      body: JSON.stringify(buildViewerProfileAvatarFileMetadata(file)),
      credentials,
      headers: {
        "Content-Type": "application/json",
      },
      method: "POST",
    },
    path: "/api/viewer/profile/avatar-uploads",
    schema: viewerProfileAvatarCreateResponseSchema,
  });

  return response.data;
}

export async function completeViewerProfileAvatarUpload(
  avatarUploadToken: string,
  {
    baseUrl,
    credentials = "include",
    fetcher = fetch,
  }: ViewerProfileAvatarUploadOptions = {},
): Promise<ViewerProfileAvatarCompleteResponse["data"]> {
  const response = await requestJson({
    baseUrl: getViewerProfileBaseURL(baseUrl),
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
    path: "/api/viewer/profile/avatar-uploads/complete",
    schema: viewerProfileAvatarCompleteResponseSchema,
  });

  return response.data;
}

export async function uploadViewerProfileAvatarTarget({
  fetcher = fetch,
  file,
  target,
}: UploadViewerProfileAvatarTargetOptions): Promise<void> {
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

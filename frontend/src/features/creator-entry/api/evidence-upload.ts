import { ApiError, requestJson } from "@/shared/api";
import { getClientEnv } from "@/shared/config";

import {
  creatorRegistrationEvidenceUploadCompleteResponseSchema,
  creatorRegistrationEvidenceUploadCreateResponseSchema,
  type CreatorRegistrationEvidence,
  type CreatorRegistrationEvidenceKind,
  type CreatorRegistrationEvidenceUploadTarget,
} from "./contracts";

type CreatorRegistrationEvidenceUploadOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
};

type UploadCreatorRegistrationEvidenceTargetOptions = {
  fetcher?: typeof fetch;
  file: File;
  target: CreatorRegistrationEvidenceUploadTarget;
};

function getCreatorRegistrationEvidenceBaseURL(baseUrl?: string): string {
  return baseUrl ?? getClientEnv().NEXT_PUBLIC_API_BASE_URL;
}

/**
 * creator registration evidence upload target を要求する。
 */
export async function createCreatorRegistrationEvidenceUpload(
  kind: CreatorRegistrationEvidenceKind,
  file: File,
  {
    baseUrl,
    credentials = "include",
    fetcher = fetch,
  }: CreatorRegistrationEvidenceUploadOptions = {},
): Promise<{
  evidenceKind: CreatorRegistrationEvidenceKind;
  evidenceUploadToken: string;
  expiresAt: string;
  uploadTarget: CreatorRegistrationEvidenceUploadTarget;
}> {
  const response = await requestJson({
    baseUrl: getCreatorRegistrationEvidenceBaseURL(baseUrl),
    fetcher,
    init: {
      body: JSON.stringify({
        fileName: file.name,
        fileSizeBytes: file.size,
        kind,
        mimeType: file.type,
      }),
      credentials,
      headers: {
        "Content-Type": "application/json",
      },
      method: "POST",
    },
    path: "/api/viewer/creator-registration/evidence-uploads",
    schema: creatorRegistrationEvidenceUploadCreateResponseSchema,
  });

  return response.data;
}

/**
 * presigned target へ evidence file を直接送る。
 */
export async function uploadCreatorRegistrationEvidenceTarget({
  fetcher = fetch,
  file,
  target,
}: UploadCreatorRegistrationEvidenceTargetOptions): Promise<void> {
  let response: Response;

  try {
    response = await fetcher(new URL(target.upload.url), {
      body: file,
      headers: target.upload.headers,
      method: target.upload.method,
    });
  } catch (error) {
    throw new ApiError("Direct evidence upload request failed before a response was received.", {
      cause: error,
      code: "network",
    });
  }

  if (!response.ok) {
    const details = await response.text().catch(() => response.statusText);

    throw new ApiError("Direct evidence upload request failed with a non-success status.", {
      code: "http",
      details,
      status: response.status,
    });
  }
}

/**
 * uploaded evidence を registration intake で使える completed state に変換する。
 */
export async function completeCreatorRegistrationEvidenceUpload(
  evidenceUploadToken: string,
  {
    baseUrl,
    credentials = "include",
    fetcher = fetch,
  }: CreatorRegistrationEvidenceUploadOptions = {},
): Promise<{
  evidence: CreatorRegistrationEvidence;
  evidenceKind: CreatorRegistrationEvidenceKind;
  evidenceUploadToken: string;
}> {
  const response = await requestJson({
    baseUrl: getCreatorRegistrationEvidenceBaseURL(baseUrl),
    fetcher,
    init: {
      body: JSON.stringify({
        evidenceUploadToken,
      }),
      credentials,
      headers: {
        "Content-Type": "application/json",
      },
      method: "POST",
    },
    path: "/api/viewer/creator-registration/evidence-uploads/complete",
    schema: creatorRegistrationEvidenceUploadCompleteResponseSchema,
  });

  return response.data;
}

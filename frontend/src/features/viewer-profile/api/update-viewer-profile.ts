import { z } from "zod";

import { requestJson } from "@/shared/api";

type UpdateViewerProfileInput = {
  avatarUploadToken?: string;
  displayName: string;
  handle: string;
};

type UpdateViewerProfileOptions = {
  baseUrl?: string;
  fetcher?: typeof fetch;
};

export async function updateViewerProfile(
  input: UpdateViewerProfileInput,
  options: UpdateViewerProfileOptions = {},
): Promise<void> {
  await requestJson({
    ...(options.baseUrl ? { baseUrl: options.baseUrl } : {}),
    ...(options.fetcher ? { fetcher: options.fetcher } : {}),
    init: {
      body: JSON.stringify({
        displayName: input.displayName,
        handle: input.handle,
        ...(input.avatarUploadToken ? { avatarUploadToken: input.avatarUploadToken } : {}),
      }),
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      method: "PUT",
    },
    path: "/api/viewer/profile",
    schema: z.undefined(),
  });
}

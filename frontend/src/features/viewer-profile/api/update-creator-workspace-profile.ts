import { z } from "zod";

import { requestJson } from "@/shared/api";

type UpdateCreatorWorkspaceProfileInput = {
  avatarUploadToken?: string;
  bio: string;
  displayName: string;
  handle: string;
};

type UpdateCreatorWorkspaceProfileOptions = {
  baseUrl?: string;
  fetcher?: typeof fetch;
};

export async function updateCreatorWorkspaceProfile(
  input: UpdateCreatorWorkspaceProfileInput,
  options: UpdateCreatorWorkspaceProfileOptions = {},
): Promise<void> {
  await requestJson({
    ...(options.baseUrl ? { baseUrl: options.baseUrl } : {}),
    ...(options.fetcher ? { fetcher: options.fetcher } : {}),
    init: {
      body: JSON.stringify({
        bio: input.bio,
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
    path: "/api/creator/workspace/profile",
    schema: z.undefined(),
  });
}

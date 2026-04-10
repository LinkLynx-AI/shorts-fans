import { z } from "zod";

import { requestJson } from "@/shared/api";

type RegisterCreatorOptions = {
  baseUrl?: string;
  fetcher?: typeof fetch;
};

type RegisterCreatorInput = {
  avatarUploadToken?: string;
  bio: string;
  displayName: string;
  handle: string;
};

/**
 * creator self-serve registration を送信する。
 */
export async function registerCreator(
  input: RegisterCreatorInput,
  options: RegisterCreatorOptions = {},
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
      method: "POST",
    },
    path: "/api/viewer/creator-registration",
    schema: z.undefined(),
  });
}

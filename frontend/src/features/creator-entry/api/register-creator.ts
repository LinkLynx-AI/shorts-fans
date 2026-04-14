import { z } from "zod";

import { requestJson } from "@/shared/api";

type RegisterCreatorOptions = {
  baseUrl?: string;
  fetcher?: typeof fetch;
};

/**
 * creator self-serve registration を送信する。
 */
export async function registerCreator(
  options: RegisterCreatorOptions = {},
): Promise<void> {
  await requestJson({
    ...(options.baseUrl ? { baseUrl: options.baseUrl } : {}),
    ...(options.fetcher ? { fetcher: options.fetcher } : {}),
    init: {
      body: JSON.stringify({}),
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

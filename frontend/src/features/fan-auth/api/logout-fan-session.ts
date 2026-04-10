import { z } from "zod";

import { requestJson } from "@/shared/api";

type LogoutFanSessionOptions = {
  baseUrl?: string;
  fetcher?: typeof fetch;
};

/**
 * fan session を logout して cookie を clear する。
 */
export async function logoutFanSession(options: LogoutFanSessionOptions = {}): Promise<void> {
  await requestJson({
    ...(options.baseUrl ? { baseUrl: options.baseUrl } : {}),
    ...(options.fetcher ? { fetcher: options.fetcher } : {}),
    init: {
      credentials: "include",
      method: "DELETE",
    },
    path: "/api/fan/auth/session",
    schema: z.undefined(),
  });
}

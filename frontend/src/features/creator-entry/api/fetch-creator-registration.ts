import { requestJson } from "@/shared/api";
import { viewerSessionCookieName } from "@/entities/viewer";

import {
  creatorRegistrationStatusResponseSchema,
  type CreatorRegistrationStatus,
} from "./contracts";

type FetchCreatorRegistrationOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
  sessionToken?: string;
};

/**
 * current viewer の creator registration status を取得する。
 */
export async function fetchCreatorRegistration(
  {
    baseUrl,
    credentials = "include",
    fetcher,
    sessionToken,
  }: FetchCreatorRegistrationOptions = {},
): Promise<CreatorRegistrationStatus | null> {
  const headers = new Headers();

  if (sessionToken) {
    headers.set("Cookie", `${viewerSessionCookieName}=${sessionToken}`);
  }

  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      credentials,
      headers,
      method: "GET",
    },
    path: "/api/viewer/creator-registration",
    schema: creatorRegistrationStatusResponseSchema,
  });

  return response.data.registration;
}

import { viewerSessionCookieName } from "@/entities/viewer";
import { requestJson } from "@/shared/api";

import { mainAccessEntryResponseSchema } from "./contracts";

type RequestMainAccessEntryOptions = {
  acceptedAge: boolean;
  acceptedTerms: boolean;
  baseUrl?: string | undefined;
  credentials?: RequestCredentials | undefined;
  entryToken: string;
  fetcher?: typeof fetch | undefined;
  fromShortId: string;
  mainId: string;
  routePath?: `/${string}` | undefined;
  sessionToken?: string | undefined;
  signal?: AbortSignal | undefined;
};

function buildMainAccessEntryPath(mainId: string): `/${string}` {
  return `/api/fan/mains/${encodeURIComponent(mainId)}/access-entry`;
}

/**
 * main access entry を発行して playback href を返す。
 */
export async function requestMainAccessEntry({
  acceptedAge,
  acceptedTerms,
  baseUrl,
  credentials = "include",
  entryToken,
  fetcher,
  fromShortId,
  mainId,
  routePath,
  sessionToken,
  signal,
}: RequestMainAccessEntryOptions) {
  const headers = new Headers();

  if (sessionToken) {
    headers.set("Cookie", `${viewerSessionCookieName}=${sessionToken}`);
  }

  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      body: JSON.stringify({
        acceptedAge,
        acceptedTerms,
        entryToken,
        fromShortId,
      }),
      credentials,
      headers: {
        "Content-Type": "application/json",
        ...Object.fromEntries(headers.entries()),
      },
      method: "POST",
      ...(signal ? { signal } : {}),
    },
    path: routePath ?? buildMainAccessEntryPath(mainId),
    schema: mainAccessEntryResponseSchema,
  });

  return response.data;
}

import { viewerSessionCookieName } from "@/entities/viewer";
import { requestJson } from "@/shared/api";

import { cardSetupTokenResponseSchema } from "./contracts";

type RequestCardSetupTokenOptions = {
  baseUrl?: string | undefined;
  credentials?: RequestCredentials | undefined;
  entryToken: string;
  fetcher?: typeof fetch | undefined;
  fromShortId: string;
  mainId: string;
  paymentTokenId: string;
  cardSetupSessionToken: string;
  sessionToken?: string | undefined;
  signal?: AbortSignal | undefined;
};

function buildCardSetupTokenPath(mainId: string): `/${string}` {
  return `/api/fan/mains/${encodeURIComponent(mainId)}/card-setup-token`;
}

/**
 * provider payment token を purchase 用 opaque token に交換する。
 */
export async function requestCardSetupToken({
  baseUrl,
  cardSetupSessionToken,
  credentials = "include",
  entryToken,
  fetcher,
  fromShortId,
  mainId,
  paymentTokenId,
  sessionToken,
  signal,
}: RequestCardSetupTokenOptions) {
  const headers = new Headers();

  if (sessionToken) {
    headers.set("Cookie", `${viewerSessionCookieName}=${sessionToken}`);
  }

  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      body: JSON.stringify({
        entryToken,
        fromShortId,
        paymentTokenId,
        sessionToken: cardSetupSessionToken,
      }),
      credentials,
      headers: {
        "Content-Type": "application/json",
        ...Object.fromEntries(headers.entries()),
      },
      method: "POST",
      ...(signal ? { signal } : {}),
    },
    path: buildCardSetupTokenPath(mainId),
    schema: cardSetupTokenResponseSchema,
  });

  return response.data;
}

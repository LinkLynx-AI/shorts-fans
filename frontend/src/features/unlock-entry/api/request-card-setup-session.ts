import { viewerSessionCookieName } from "@/entities/viewer";
import { requestJson } from "@/shared/api";

import { cardSetupSessionResponseSchema } from "./contracts";

type RequestCardSetupSessionOptions = {
  baseUrl?: string | undefined;
  credentials?: RequestCredentials | undefined;
  entryToken: string;
  fetcher?: typeof fetch | undefined;
  fromShortId: string;
  mainId: string;
  sessionToken?: string | undefined;
  signal?: AbortSignal | undefined;
};

function buildCardSetupSessionPath(mainId: string): `/${string}` {
  return `/api/fan/mains/${encodeURIComponent(mainId)}/card-setup-session`;
}

/**
 * new card 用の CCBill widget 初期化情報を取得する。
 */
export async function requestCardSetupSession({
  baseUrl,
  credentials = "include",
  entryToken,
  fetcher,
  fromShortId,
  mainId,
  sessionToken,
  signal,
}: RequestCardSetupSessionOptions) {
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
      }),
      credentials,
      headers: {
        "Content-Type": "application/json",
        ...Object.fromEntries(headers.entries()),
      },
      method: "POST",
      ...(signal ? { signal } : {}),
    },
    path: buildCardSetupSessionPath(mainId),
    schema: cardSetupSessionResponseSchema,
  });

  return response.data;
}

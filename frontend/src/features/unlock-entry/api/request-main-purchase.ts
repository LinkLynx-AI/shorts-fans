import { viewerSessionCookieName } from "@/entities/viewer";
import { requestJson } from "@/shared/api";

import { mainPurchaseResponseSchema } from "./contracts";

export type RequestMainPurchaseOptions = {
  acceptedAge: boolean;
  acceptedTerms: boolean;
  baseUrl?: string | undefined;
  credentials?: RequestCredentials | undefined;
  entryToken: string;
  fetcher?: typeof fetch | undefined;
  fromShortId: string;
  mainId: string;
  paymentMethod:
    | {
        mode: "saved_card";
        paymentMethodId: string;
      }
    | {
        cardSetupToken: string;
        mode: "new_card";
      };
  purchasePath?: `/${string}` | undefined;
  sessionToken?: string | undefined;
  signal?: AbortSignal | undefined;
};

function buildMainPurchasePath(mainId: string): `/${string}` {
  return `/api/fan/mains/${encodeURIComponent(mainId)}/purchase`;
}

/**
 * main purchase を実行し、purchase outcome を返す。
 */
export async function requestMainPurchase({
  acceptedAge,
  acceptedTerms,
  baseUrl,
  credentials = "include",
  entryToken,
  fetcher,
  fromShortId,
  mainId,
  paymentMethod,
  purchasePath,
  sessionToken,
  signal,
}: RequestMainPurchaseOptions) {
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
        paymentMethod,
      }),
      credentials,
      headers: {
        "Content-Type": "application/json",
        ...Object.fromEntries(headers.entries()),
      },
      method: "POST",
      ...(signal ? { signal } : {}),
    },
    path: purchasePath ?? buildMainPurchasePath(mainId),
    schema: mainPurchaseResponseSchema,
  });

  return response.data;
}

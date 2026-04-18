import { viewerSessionCookieName } from "@/entities/viewer";
import { requestJson } from "@/shared/api";

import { publicShortDetailResponseSchema, type PublicShortDetail } from "./contracts";

type GetPublicShortDetailOptions = {
  baseUrl?: string | undefined;
  credentials?: RequestCredentials | undefined;
  fetcher?: typeof fetch | undefined;
  sessionToken?: string | undefined;
  shortId: string;
  signal?: AbortSignal | undefined;
};

/**
 * public short detail を取得する。
 */
export async function getPublicShortDetail({
  baseUrl,
  credentials = "include",
  fetcher,
  sessionToken,
  shortId,
  signal,
}: GetPublicShortDetailOptions): Promise<PublicShortDetail> {
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
      ...(signal ? { signal } : {}),
    },
    path: `/api/fan/shorts/${encodeURIComponent(shortId)}`,
    schema: publicShortDetailResponseSchema,
  });

  return response.data.detail;
}

import { viewerSessionCookieName } from "@/entities/viewer";
import { requestJson } from "@/shared/api";

import { unlockSurfaceResponseSchema } from "./contracts";
import { normalizeUnlockSurface } from "../model/unlock-entry";

type RequestUnlockSurfaceByShortIdOptions = {
  baseUrl?: string | undefined;
  fetcher?: typeof fetch | undefined;
  sessionToken?: string | undefined;
  shortId: string;
  signal?: AbortSignal | undefined;
};

/**
 * short ごとの unlock surface を取得する。
 */
export async function requestUnlockSurfaceByShortId({
  baseUrl,
  sessionToken,
  fetcher,
  shortId,
  signal,
}: RequestUnlockSurfaceByShortIdOptions) {
  const headers = new Headers();

  if (sessionToken) {
    headers.set("Cookie", `${viewerSessionCookieName}=${sessionToken}`);
  }

  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      credentials: "include",
      headers,
      ...(signal ? { signal } : {}),
    },
    path: `/api/fan/shorts/${encodeURIComponent(shortId)}/unlock`,
    schema: unlockSurfaceResponseSchema,
  });

  return normalizeUnlockSurface(response.data);
}

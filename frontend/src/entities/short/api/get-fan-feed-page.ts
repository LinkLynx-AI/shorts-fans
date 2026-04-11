import { viewerSessionCookieName } from "@/entities/viewer";
import { requestJson } from "@/shared/api";

import { fanFeedResponseSchema, type FanFeedItem, type FanFeedTab } from "./contracts";

export type FanFeedPage = {
  items: readonly FanFeedItem[];
  page: {
    hasNext: boolean;
    nextCursor: string | null;
  };
  requestId: string;
  tab: FanFeedTab;
};

type GetFanFeedPageOptions = {
  baseUrl?: string | undefined;
  cursor?: string | undefined;
  fetcher?: typeof fetch | undefined;
  sessionToken?: string | undefined;
  signal?: AbortSignal | undefined;
  tab: FanFeedTab;
};

function buildFanFeedPath(tab: FanFeedTab, cursor?: string): `/${string}` {
  const searchParams = new URLSearchParams({
    tab,
  });
  const trimmedCursor = cursor?.trim();

  if (trimmedCursor) {
    searchParams.set("cursor", trimmedCursor);
  }

  return `/api/fan/feed?${searchParams.toString()}` as `/${string}`;
}

/**
 * fan public feed の 1 page を取得する。
 */
export async function getFanFeedPage({
  baseUrl,
  cursor,
  fetcher,
  sessionToken,
  signal,
  tab,
}: GetFanFeedPageOptions): Promise<FanFeedPage> {
  const headers = new Headers();

  if (sessionToken) {
    headers.set("Cookie", `${viewerSessionCookieName}=${sessionToken}`);
  }

  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      headers,
      ...(signal ? { signal } : {}),
    },
    path: buildFanFeedPath(tab, cursor),
    schema: fanFeedResponseSchema,
  });

  return {
    items: response.data.items,
    page: response.meta.page,
    requestId: response.meta.requestId,
    tab: response.data.tab,
  };
}

import { z } from "zod";

import { viewerSessionCookieName } from "@/entities/viewer";

export const fanProfileCursorPageSchema = z.object({
  hasNext: z.boolean(),
  nextCursor: z.string().min(1).nullable(),
});

export const fanProfileCursorMetaSchema = z.object({
  page: fanProfileCursorPageSchema,
  requestId: z.string().min(1),
});

export type FanProfileCursorPage = z.output<typeof fanProfileCursorPageSchema>;

/**
 * fan profile cursor endpoint の path を組み立てる。
 */
export function buildFanProfileCursorPath(basePath: `/${string}`, cursor?: string): `/${string}` {
  const trimmedCursor = cursor?.trim();

  if (!trimmedCursor) {
    return basePath;
  }

  const searchParams = new URLSearchParams();
  searchParams.set("cursor", trimmedCursor);

  return `${basePath}?${searchParams.toString()}` as `/${string}`;
}

/**
 * fan profile request 用の headers を組み立てる。
 */
export function buildFanProfileRequestHeaders(sessionToken?: string): Headers {
  const headers = new Headers();

  if (sessionToken) {
    headers.set("Cookie", `${viewerSessionCookieName}=${sessionToken}`);
  }

  return headers;
}

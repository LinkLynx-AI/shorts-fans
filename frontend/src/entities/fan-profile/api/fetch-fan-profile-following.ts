import { z } from "zod";

import { viewerSessionCookieName } from "@/entities/viewer";
import { requestJson } from "@/shared/api";

import type { FanFollowingItem } from "../model/fan-profile";

const creatorAvatarAssetSchema = z.object({
  durationSeconds: z.null(),
  id: z.string().min(1),
  kind: z.literal("image"),
  posterUrl: z.string().min(1).nullable(),
  url: z.string().min(1),
});

const creatorSummarySchema = z.object({
  avatar: creatorAvatarAssetSchema.nullable(),
  bio: z.string(),
  displayName: z.string().min(1),
  handle: z.custom<`@${string}`>((value) => typeof value === "string" && value.startsWith("@")),
  id: z.string().min(1),
});

const fanFollowingItemSchema = z.object({
  creator: creatorSummarySchema,
  viewer: z.object({
    isFollowing: z.literal(true),
  }),
});

const fanProfileFollowingResponseSchema = z.object({
  data: z.object({
    items: z.array(fanFollowingItemSchema),
  }),
  error: z.null(),
  meta: z.object({
    page: z.object({
      hasNext: z.boolean(),
      nextCursor: z.string().min(1).nullable(),
    }),
    requestId: z.string().min(1),
  }),
});

export type FanProfileFollowingPage = {
  items: readonly FanFollowingItem[];
  page: {
    hasNext: boolean;
    nextCursor: string | null;
  };
  requestId: string;
};

type FetchFanProfileFollowingPageOptions = {
  baseUrl?: string | undefined;
  cursor?: string | undefined;
  fetcher?: typeof fetch | undefined;
  sessionToken?: string | undefined;
  signal?: AbortSignal | undefined;
};

function buildFanProfileFollowingPath(cursor?: string): `/${string}` {
  const trimmedCursor = cursor?.trim();

  if (!trimmedCursor) {
    return "/api/fan/profile/following";
  }

  const searchParams = new URLSearchParams();
  searchParams.set("cursor", trimmedCursor);

  return `/api/fan/profile/following?${searchParams.toString()}` as `/${string}`;
}

/**
 * fan profile following API から creator 一覧 1 page を取得する。
 */
export async function fetchFanProfileFollowingPage({
  baseUrl,
  cursor,
  fetcher,
  sessionToken,
  signal,
}: FetchFanProfileFollowingPageOptions = {}): Promise<FanProfileFollowingPage> {
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
    path: buildFanProfileFollowingPath(cursor),
    schema: fanProfileFollowingResponseSchema,
  });

  return {
    items: response.data.items,
    page: response.meta.page,
    requestId: response.meta.requestId,
  };
}

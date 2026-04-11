import { z } from "zod";

import { viewerSessionCookieName } from "@/entities/viewer";
import { requestJson } from "@/shared/api";

import type { MainPlaybackPayload } from "../model/main-playback-surface";

const videoAssetSchema = z.object({
  durationSeconds: z.number().int().positive(),
  id: z.string().min(1),
  kind: z.literal("video"),
  posterUrl: z.string().nullable(),
  url: z.string().url(),
});

const creatorSchema = z.object({
  avatar: z.object({
    durationSeconds: z.null(),
    id: z.string().min(1),
    kind: z.literal("image"),
    posterUrl: z.null(),
    url: z.string().min(1),
  }),
  bio: z.string().min(1),
  displayName: z.string().min(1),
  handle: z.custom<`@${string}`>((value) => typeof value === "string" && value.startsWith("@")),
  id: z.string().min(1),
});

const shortSchema = z
  .object({
    canonicalMainId: z.string().min(1),
    caption: z.string().min(1),
    creatorId: z.string().min(1),
    id: z.string().min(1),
    media: videoAssetSchema.extend({
      durationSeconds: z.number().int().nonnegative().nullable(),
    }),
    previewDurationSeconds: z.number().int().nonnegative(),
    title: z.string().min(1),
  })
  .nullable();

const mainPlaybackPayloadSchema = z.object({
  access: z.object({
    mainId: z.string().min(1),
    reason: z.enum(["owner_preview", "session_unlocked", "unlock_required"]),
    status: z.enum(["locked", "owner", "unlocked"]),
  }),
  creator: creatorSchema,
  entryShort: shortSchema,
  main: z.object({
    durationSeconds: z.number().int().positive(),
    id: z.string().min(1),
    media: videoAssetSchema,
    priceJpy: z.number().int().positive(),
    title: z.string().min(1),
  }),
  resumePositionSeconds: z.number().int().nonnegative().nullable(),
});

const mainPlaybackResponseSchema = z.object({
  data: mainPlaybackPayloadSchema,
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

export type FetchMainPlaybackOptions = {
  baseUrl?: string | undefined;
  fetcher?: typeof fetch | undefined;
  fromShortId: string;
  grant: string;
  mainId: string;
  sessionToken?: string | undefined;
};

function buildMainPlaybackPath(mainId: string, fromShortId: string, grant: string): `/${string}` {
  const searchParams = new URLSearchParams({
    fromShortId,
    grant,
  });

  return `/api/fan/mains/${encodeURIComponent(mainId)}/playback?${searchParams.toString()}`;
}

/**
 * main playback API から contract-backed payload を取得する。
 */
export async function fetchMainPlayback({
  baseUrl,
  fetcher,
  fromShortId,
  grant,
  mainId,
  sessionToken,
}: FetchMainPlaybackOptions): Promise<MainPlaybackPayload> {
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
    },
    path: buildMainPlaybackPath(mainId, fromShortId, grant),
    schema: mainPlaybackResponseSchema,
  });

  return response.data;
}

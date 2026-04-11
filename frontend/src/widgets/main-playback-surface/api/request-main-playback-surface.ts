import { creatorSummarySchema } from "@/entities/creator";
import { shortVideoDisplayAssetSchema } from "@/entities/short";
import { viewerSessionCookieName } from "@/entities/viewer";
import {
  mainAccessStateSchema,
  unlockShortSummarySchema,
} from "@/features/unlock-entry";
import { requestJson } from "@/shared/api";
import { z } from "zod";

import type { MainPlaybackSurface } from "../model/main-playback-surface";

const mainPlaybackSurfaceResponseSchema = z.object({
  data: z.object({
    access: mainAccessStateSchema,
    creator: creatorSummarySchema,
    entryShort: unlockShortSummarySchema,
    main: z.object({
      durationSeconds: z.number().int().positive(),
      id: z.string().min(1),
      media: shortVideoDisplayAssetSchema.extend({
        durationSeconds: z.number().int().positive(),
      }),
    }),
    resumePositionSeconds: z.number().int().nonnegative().nullable(),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

type RequestMainPlaybackSurfaceOptions = {
  baseUrl?: string | undefined;
  fetcher?: typeof fetch | undefined;
  fromShortId: string;
  grant: string;
  mainId: string;
  sessionToken?: string | undefined;
  signal?: AbortSignal | undefined;
};

function buildMainPlaybackPath(mainId: string, fromShortId: string, grant: string): `/${string}` {
  const searchParams = new URLSearchParams({
    fromShortId,
    grant,
  });

  return `/api/fan/mains/${encodeURIComponent(mainId)}/playback?${searchParams.toString()}` as `/${string}`;
}

/**
 * main playback surface を API から取得する。
 */
export async function requestMainPlaybackSurface({
  baseUrl,
  fetcher,
  fromShortId,
  grant,
  mainId,
  sessionToken,
  signal,
}: RequestMainPlaybackSurfaceOptions): Promise<MainPlaybackSurface> {
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
    path: buildMainPlaybackPath(mainId, fromShortId, grant),
    schema: mainPlaybackSurfaceResponseSchema,
  });

  return {
    access: response.data.access,
    creator: response.data.creator,
    entryShort: response.data.entryShort,
    main: {
      ...response.data.main,
      priceJpy: 0,
    },
    resumePositionSeconds: response.data.resumePositionSeconds,
    themeShort: response.data.entryShort,
    viewer: {
      isPinned: false,
    },
  };
}

import { z } from "zod";

import { requestJson } from "@/shared/api";
import type { FeedShortSurface } from "@/widgets/immersive-short-surface";

import {
  createReadyFeedShellState,
  getEmptyFeedShellState,
  type FeedShellPageInfo,
  type FeedShellState,
} from "../model/mock-feed-shell";

const mediaAssetSchema = z.object({
  durationSeconds: z.number().int().nullable(),
  id: z.string().min(1),
  kind: z.enum(["image", "video"]),
  posterUrl: z.string().nullable(),
  url: z.string().min(1),
});

const creatorSchema = z.object({
  avatar: mediaAssetSchema.extend({
    durationSeconds: z.null(),
    kind: z.literal("image"),
    posterUrl: z.null(),
  }),
  bio: z.string(),
  displayName: z.string().min(1),
  handle: z.string().min(1),
  id: z.string().uuid(),
});

const shortSchema = z.object({
  caption: z.string(),
  canonicalMainId: z.string().uuid(),
  creatorId: z.string().uuid(),
  id: z.string().uuid(),
  media: mediaAssetSchema.extend({
    kind: z.literal("video"),
  }),
  previewDurationSeconds: z.number().int().nonnegative(),
  title: z.string(),
});

const unlockCtaSchema = z.object({
  mainDurationSeconds: z.number().int().nullable(),
  priceJpy: z.number().int().nullable(),
  resumePositionSeconds: z.number().int().nullable(),
  state: z.enum([
    "continue_main",
    "owner_preview",
    "setup_required",
    "unavailable",
    "unlock_available",
  ]),
});

const recommendedFeedResponseSchema = z.object({
  data: z.object({
    items: z.array(
      z.object({
        creator: creatorSchema,
        short: shortSchema,
        unlockCta: unlockCtaSchema,
        viewer: z.object({
          isPinned: z.boolean(),
        }),
      }),
    ),
    tab: z.literal("recommended"),
  }),
  error: z.null(),
  meta: z.object({
    page: z.object({
      hasNext: z.boolean(),
      nextCursor: z.string().nullable(),
    }),
    requestId: z.string().min(1),
  }),
});

type RecommendedFeedResponse = z.output<typeof recommendedFeedResponseSchema>;

/**
 * recommended feed を取得して shell state に変換する。
 */
export async function fetchRecommendedFeedShellState({
  baseUrl,
  fetcher,
}: {
  baseUrl: string;
  fetcher?: typeof fetch;
}): Promise<FeedShellState> {
  const response = await requestJson({
    baseUrl,
    init: {
      cache: "no-store",
    },
    path: "/api/fan/feed?tab=recommended",
    schema: recommendedFeedResponseSchema,
    ...(fetcher ? { fetcher } : {}),
  });

  const firstItem = response.data.items[0];

  if (!firstItem) {
    return getEmptyFeedShellState("recommended");
  }

  return createReadyFeedShellState({
    page: mapPageInfo(response),
    surface: mapFeedItemToSurface(firstItem),
    tab: "recommended",
  });
}

function mapPageInfo(response: RecommendedFeedResponse): FeedShellPageInfo {
  return {
    hasNext: response.meta.page.hasNext,
    nextCursor: response.meta.page.nextCursor,
  };
}

function mapFeedItemToSurface(item: RecommendedFeedResponse["data"]["items"][number]): FeedShortSurface {
  return {
    creator: {
      avatar: item.creator.avatar,
      bio: item.creator.bio,
      displayName: item.creator.displayName,
      handle: normalizeCreatorHandle(item.creator.handle),
      id: item.creator.id,
    },
    short: {
      caption: item.short.caption,
      canonicalMainId: item.short.canonicalMainId,
      creatorId: item.short.creatorId,
      id: item.short.id,
      media: item.short.media,
      previewDurationSeconds: item.short.previewDurationSeconds,
      title: item.short.title,
    },
    unlockCta: item.unlockCta,
    viewer: item.viewer,
  };
}

function normalizeCreatorHandle(handle: string): `@${string}` {
  const normalized = handle.startsWith("@") ? handle : `@${handle}`;
  return normalized as `@${string}`;
}

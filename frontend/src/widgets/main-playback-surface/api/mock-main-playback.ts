import { z } from "zod";

import { getCreatorById } from "@/entities/creator";
import { getMainById } from "@/entities/main";
import { getPinnedShorts, getShortById } from "@/entities/short";
import { getUnlockSurfaceByShortId, type MainAccessState } from "@/features/unlock-entry";

import type { MainPlaybackSurface } from "../model/main-playback-surface";

const mainMediaSchema = z.object({
  durationSeconds: z.number().int().positive(),
  id: z.string().min(1),
  kind: z.literal("video"),
  posterUrl: z.string().nullable(),
  url: z.string().url(),
});

const playbackMainSchema = z.object({
  durationSeconds: z.number().int().positive(),
  id: z.string().min(1),
  media: mainMediaSchema,
  priceJpy: z.number().int().positive(),
  title: z.string().min(1),
});

const mediaSchema = z.object({
  durationSeconds: z.number().int().nonnegative().nullable(),
  id: z.string().min(1),
  kind: z.literal("video"),
  posterUrl: z.string().nullable(),
  url: z.string().url(),
});

const entryShortSchema = z
  .object({
    canonicalMainId: z.string().min(1),
    caption: z.string().min(1),
    creatorId: z.string().min(1),
    id: z.string().min(1),
    media: mediaSchema,
    previewDurationSeconds: z.number().int().nonnegative(),
    title: z.string().min(1),
  })
  .nullable();

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

const accessSchema = z.object({
  mainId: z.string().min(1),
  reason: z.enum(["owner_preview", "purchase_required", "purchased_access"]),
  status: z.enum(["locked", "owner", "purchased"]),
});

const playbackSurfaceSchema = z.object({
  access: accessSchema,
  creator: creatorSchema,
  entryShort: entryShortSchema,
  main: playbackMainSchema,
  resumePositionSeconds: z.number().int().nonnegative().nullable(),
  themeShort: entryShortSchema.unwrap(),
  viewer: z.object({
    isPinned: z.boolean(),
  }),
});

function buildPlaybackAccess(shortId: string | undefined, mainId: string): {
  access: MainAccessState;
  resumePositionSeconds: number | null;
} | undefined {
  if (!shortId) {
    return undefined;
  }

  const unlock = getUnlockSurfaceByShortId(shortId);

  if (!unlock) {
    return undefined;
  }

  if (unlock.access.status === "owner") {
    return {
      access: {
        mainId,
        reason: "owner_preview",
        status: "owner",
      },
      resumePositionSeconds: null,
    };
  }

  return {
    access: {
      mainId,
      reason: "purchased_access",
      status: "purchased",
    },
    resumePositionSeconds: unlock.unlockCta.resumePositionSeconds,
  };
}

/**
 * main playback surface 用の mock を取得する。
 */
export function getMainPlaybackSurfaceById(
  mainId: string,
  fromShortId?: string,
): MainPlaybackSurface | undefined {
  if (!fromShortId) {
    return undefined;
  }

  const main = getMainById(mainId);

  if (!main) {
    return undefined;
  }

  const candidateEntryShort = getShortById(fromShortId);

  if (!candidateEntryShort || candidateEntryShort.canonicalMainId !== mainId) {
    return undefined;
  }

  const entryShort = candidateEntryShort;
  const themeShort = entryShort ?? getShortById(main.leadShortId);

  if (!themeShort) {
    throw new Error(`Unknown theme short for main playback: ${main.id}`);
  }

  const creator = getCreatorById(themeShort.creatorId);

  if (!creator) {
    throw new Error(`Unknown creator for main playback: ${themeShort.creatorId}`);
  }

  const playbackState = buildPlaybackAccess(entryShort?.id, main.id);

  if (!playbackState) {
    return undefined;
  }

  const pinnedShortIds = new Set(getPinnedShorts().map((short) => short.id));

  return playbackSurfaceSchema.parse({
    access: playbackState.access,
    creator,
    entryShort,
    main: {
      durationSeconds: main.durationSeconds,
      id: main.id,
      media: main.media,
      priceJpy: main.priceJpy,
      title: main.title,
    },
    resumePositionSeconds: playbackState.resumePositionSeconds,
    themeShort,
    viewer: {
      isPinned: entryShort ? pinnedShortIds.has(entryShort.id) : false,
    },
  });
}

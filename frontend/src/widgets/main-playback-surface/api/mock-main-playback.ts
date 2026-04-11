import { z } from "zod";

import { getCreatorById } from "@/entities/creator";
import { getMainById } from "@/entities/main";
import { getShortById } from "@/entities/short";
import {
  getUnlockSurfaceByShortId,
  type MainAccessState,
  type MainPlaybackGrantKind,
} from "@/features/unlock-entry";

import {
  buildMainPlaybackSurface,
  type MainPlaybackPayload,
  type MainPlaybackSurface,
} from "../model/main-playback-surface";

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
  reason: z.enum(["owner_preview", "session_unlocked", "unlock_required"]),
  status: z.enum(["locked", "owner", "unlocked"]),
});

const playbackPayloadSchema = z.object({
  access: accessSchema,
  creator: creatorSchema,
  entryShort: entryShortSchema,
  main: playbackMainSchema,
  resumePositionSeconds: z.number().int().nonnegative().nullable(),
});

function buildPlaybackAccess(
  shortId: string | undefined,
  mainId: string,
  grantKind: MainPlaybackGrantKind,
): {
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

  if (grantKind === "owner") {
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
      reason: "session_unlocked",
      status: "unlocked",
    },
    resumePositionSeconds: unlock.unlockCta.resumePositionSeconds,
  };
}

/**
 * main playback transport payload 用の mock を取得する。
 */
export function getMainPlaybackPayloadById(
  mainId: string,
  fromShortId?: string,
  grantKind?: MainPlaybackGrantKind,
): MainPlaybackPayload | undefined {
  if (!fromShortId || !grantKind) {
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

  const playbackState = buildPlaybackAccess(entryShort?.id, main.id, grantKind);

  if (!playbackState) {
    return undefined;
  }

  return playbackPayloadSchema.parse({
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
  });
}

/**
 * main playback surface 用の mock を取得する。
 */
export function getMainPlaybackSurfaceById(
  mainId: string,
  fromShortId?: string,
  grantKind?: MainPlaybackGrantKind,
): MainPlaybackSurface | undefined {
  if (!fromShortId) {
    return undefined;
  }

  const payload = getMainPlaybackPayloadById(mainId, fromShortId, grantKind);

  if (!payload) {
    return undefined;
  }

  return buildMainPlaybackSurface(payload, fromShortId);
}

import type { FanLibraryItem } from "@/entities/fan-profile";
import {
  requestMainAccessEntry,
  requestUnlockSurfaceByShortId,
} from "@/features/unlock-entry";
import { ApiError } from "@/shared/api";

import { getMainPlaybackSurfaceById } from "./mock-main-playback";
import { requestMainPlaybackSurface } from "./request-main-playback-surface";

export type LibraryMainPlaybackResolution =
  | {
      kind: "locked";
    }
  | {
      kind: "ready";
      surface: Awaited<ReturnType<typeof requestMainPlaybackSurface>>;
    };

function extractGrantFromHref(href: string): string {
  const playbackUrl = new URL(href, "http://localhost");
  const grant = playbackUrl.searchParams.get("grant");

  if (!grant) {
    throw new Error("Main playback href does not contain grant");
  }

  return grant;
}

/**
 * library item から main playback surface を解決する。
 */
export async function resolveLibraryMainPlaybackSurface(
  item: FanLibraryItem,
): Promise<LibraryMainPlaybackResolution> {
  if (!item.entryShort.id.startsWith("short_")) {
    const legacySurface = getMainPlaybackSurfaceById(
      item.main.id,
      item.entryShort.id,
      item.access.status === "owner" ? "owner" : "unlocked",
    );

    if (!legacySurface) {
      throw new Error("Library legacy item could not resolve playback surface");
    }

    return {
      kind: "ready",
      surface: legacySurface,
    };
  }

  let unlock: Awaited<ReturnType<typeof requestUnlockSurfaceByShortId>>;

  try {
    unlock = await requestUnlockSurfaceByShortId({
      shortId: item.entryShort.id,
    });
  } catch (error) {
    if (error instanceof ApiError && error.code === "http" && error.status === 403) {
      return {
        kind: "locked",
      };
    }

    throw error;
  }

  if (unlock.main.id !== item.main.id || unlock.short.id !== item.entryShort.id) {
    throw new Error("Library entry context does not match unlock surface");
  }

  const expectedUnlockState = item.access.status === "owner" ? "owner_preview" : "continue_main";

  if (unlock.unlockCta.state !== expectedUnlockState) {
    return {
      kind: "locked",
    };
  }

  let accessEntry: Awaited<ReturnType<typeof requestMainAccessEntry>>;

  try {
    accessEntry = await requestMainAccessEntry({
      acceptedAge: false,
      acceptedTerms: false,
      entryToken: unlock.mainAccessEntry.token,
      fromShortId: item.entryShort.id,
      mainId: item.main.id,
      routePath: unlock.mainAccessEntry.routePath as `/${string}`,
    });
  } catch (error) {
    if (error instanceof ApiError && error.code === "http" && error.status === 403) {
      return {
        kind: "locked",
      };
    }

    throw error;
  }

  try {
    const surface = await requestMainPlaybackSurface({
      fromShortId: item.entryShort.id,
      grant: extractGrantFromHref(accessEntry.href),
      mainId: item.main.id,
    });

    return {
      kind: "ready",
      surface,
    };
  } catch (error) {
    if (error instanceof ApiError && error.code === "http" && error.status === 403) {
      return {
        kind: "locked",
      };
    }

    throw error;
  }
}

import type { FanLibraryItem } from "@/entities/fan-profile";
import {
  requestMainAccessEntry,
  requestUnlockSurfaceByShortId,
} from "@/features/unlock-entry";

import { getMainPlaybackSurfaceById } from "./mock-main-playback";
import { requestMainPlaybackSurface } from "./request-main-playback-surface";

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
) {
  if (!item.entryShort.id.startsWith("short_")) {
    const legacySurface = getMainPlaybackSurfaceById(
      item.main.id,
      item.entryShort.id,
      item.access.status === "owner" ? "owner" : "unlocked",
    );

    if (!legacySurface) {
      throw new Error("Library legacy item could not resolve playback surface");
    }

    return legacySurface;
  }

  const unlock = await requestUnlockSurfaceByShortId({
    shortId: item.entryShort.id,
  });

  if (unlock.main.id !== item.main.id || unlock.short.id !== item.entryShort.id) {
    throw new Error("Library entry context does not match unlock surface");
  }

  const accessEntry = await requestMainAccessEntry({
    acceptedAge: true,
    acceptedTerms: true,
    entryToken: unlock.mainAccessEntry.token,
    fromShortId: item.entryShort.id,
    mainId: item.main.id,
    routePath: unlock.mainAccessEntry.routePath as `/${string}`,
  });

  return requestMainPlaybackSurface({
    fromShortId: item.entryShort.id,
    grant: extractGrantFromHref(accessEntry.href),
    mainId: item.main.id,
  });
}

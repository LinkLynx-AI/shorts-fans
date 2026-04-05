import type { CreatorSummary } from "@/entities/creator";
import type { MainMediaAsset, MainSummary } from "@/entities/main";
import type { ShortPreviewMeta } from "@/entities/short";
import type { MainAccessState } from "@/features/unlock-entry";

export type MainPlaybackSurfaceMain = MainSummary & {
  media: MainMediaAsset;
};

export type MainPlaybackSurface = {
  access: MainAccessState;
  creator: CreatorSummary;
  entryShort: ShortPreviewMeta | null;
  main: MainPlaybackSurfaceMain;
  resumePositionSeconds: number | null;
  themeShort: ShortPreviewMeta;
  viewer: {
    isPinned: boolean;
  };
};

function formatSecondsAsTimestamp(seconds: number): string {
  const normalized = Math.max(0, Math.floor(seconds));
  const minutes = Math.floor(normalized / 60);
  const remainingSeconds = normalized % 60;

  return `${minutes}:${remainingSeconds.toString().padStart(2, "0")}`;
}

function formatDurationAsMinutes(seconds: number): string {
  const roundedMinutes = Math.max(1, Math.round(seconds / 60));
  return `${roundedMinutes}分`;
}

/**
 * main player の status title を返す。
 */
export function getMainPlaybackStatusTitle(surface: MainPlaybackSurface): string {
  return surface.access.status === "owner" ? "Owner preview" : "Playing main";
}

/**
 * main player の補助コピーを返す。
 */
export function getMainPlaybackStatusCopy(surface: MainPlaybackSurface): string {
  if (surface.access.status === "owner") {
    return "purchase confirmation is skipped for your own main";
  }

  if (surface.resumePositionSeconds !== null) {
    return "resume without another confirmation";
  }

  return "continue from this short without another confirmation";
}

/**
 * main player の右側 pill 表示を返す。
 */
export function getMainPlaybackStatusMeta(surface: MainPlaybackSurface): string {
  if (surface.access.status === "owner") {
    return "Owner preview";
  }

  if (surface.resumePositionSeconds !== null) {
    return formatSecondsAsTimestamp(surface.resumePositionSeconds);
  }

  return formatDurationAsMinutes(surface.main.durationSeconds);
}

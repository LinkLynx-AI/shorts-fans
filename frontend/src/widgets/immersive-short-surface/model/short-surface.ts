import type { CreatorSummary } from "@/entities/creator";
import type { ShortPreviewMeta } from "@/entities/short";
import type { UnlockSurfaceModel } from "@/features/unlock-entry";

export type FeedSurfaceViewerState = {
  isPinned: boolean;
};

export type DetailSurfaceViewerState = FeedSurfaceViewerState & {
  isFollowingCreator: boolean;
};

export type ShortSurfaceBase = {
  creator: CreatorSummary;
  mainEntryEnabled: boolean;
  short: ShortPreviewMeta;
  unlock: UnlockSurfaceModel;
};

export type FeedShortSurface = ShortSurfaceBase & {
  viewer: FeedSurfaceViewerState;
};

export type DetailShortSurface = ShortSurfaceBase & {
  viewer: DetailSurfaceViewerState;
};

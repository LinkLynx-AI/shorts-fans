import type { CreatorSummary } from "@/entities/creator";
import type { UnlockSurfaceModel } from "@/features/unlock-entry";

export type FeedSurfaceViewerState = {
  isFollowingCreator: boolean;
  isPinned: boolean;
};

export type DetailSurfaceViewerState = FeedSurfaceViewerState;

export type ShortSurfaceBase = {
  creator: CreatorSummary;
  mainEntryEnabled: boolean;
  short: UnlockSurfaceModel["short"];
  unlock: UnlockSurfaceModel;
};

export type FeedShortSurface = ShortSurfaceBase & {
  viewer: FeedSurfaceViewerState;
};

export type DetailShortSurface = ShortSurfaceBase & {
  viewer: DetailSurfaceViewerState;
};

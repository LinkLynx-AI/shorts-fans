"use client";

import { ImmersiveShortSurface } from "@/widgets/immersive-short-surface";
import { VerticalSnapReel } from "@/shared/ui";

import type { FeedShortSurface } from "@/widgets/immersive-short-surface";
import type { FeedTab } from "@/entities/short";

import { useFeedLikeState } from "../model/use-feed-like-state";
import { useFeedPinState } from "../model/use-feed-pin-state";

type FeedReelProps = {
  activeTab: FeedTab;
  surfaces: readonly FeedShortSurface[];
};

/**
 * full-screen short surfaces を縦方向の snap scroll で連続視聴させる。
 */
export function FeedReel({ activeTab, surfaces }: FeedReelProps) {
  const { resolveLikeState } = useFeedLikeState({ surfaces });
  const { resolvePinState } = useFeedPinState({ surfaces });

  return (
    <VerticalSnapReel
      getKey={(surface) => surface.short.id}
      items={surfaces}
      renderItem={(surface, { isActive }) => (
        <ImmersiveShortSurface
          activeTab={activeTab}
          isActive={isActive}
          like={resolveLikeState(surface)}
          mode="feed"
          pin={resolvePinState(surface)}
          surface={surface}
        />
      )}
    />
  );
}

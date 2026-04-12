"use client";

import type { FanHubTab } from "@/entities/fan-profile";
import { VerticalSnapReel } from "@/shared/ui";

import type { DetailShortSurface } from "../model/short-surface";
import { ImmersiveShortSurface } from "./immersive-short-surface";

type ShortDetailReelProps = {
  backHref: string;
  fanTab?: FanHubTab | undefined;
  initialIndex: number;
  source: "creator" | "fan";
  surfaces: readonly DetailShortSurface[];
};

/**
 * profile/list 起点の short detail reel を表示する。
 */
export function ShortDetailReel({
  backHref,
  fanTab,
  initialIndex,
  source,
  surfaces,
}: ShortDetailReelProps) {
  return (
    <VerticalSnapReel
      getKey={(surface) => surface.short.id}
      initialIndex={initialIndex}
      items={surfaces}
      renderItem={(surface, { isActive }) => (
        <ImmersiveShortSurface
          backHref={backHref}
          creatorProfileOrigin={{
            from: "short",
            ...(source === "fan" ? { shortFanTab: fanTab } : {}),
            shortId: surface.short.id,
          }}
          isActive={isActive}
          mode="detail"
          surface={surface}
        />
      )}
    />
  );
}

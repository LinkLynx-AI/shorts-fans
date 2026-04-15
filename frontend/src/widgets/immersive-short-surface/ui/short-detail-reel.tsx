"use client";

import Link from "next/link";
import { useEffect, useRef, useState, type ReactNode } from "react";

import type { FanHubTab } from "@/entities/fan-profile";
import { getPublicShortDetail } from "@/entities/short";
import { VerticalSnapReel } from "@/shared/ui";
import { Button } from "@/shared/ui";

import { buildDetailSurfaceFromApi } from "../model/api-short-surface";
import type { DetailShortSurface } from "../model/short-surface";
import {
  FeedLikeShortBackdrop,
  FeedLikeShortBackHeader,
  ImmersiveShortSurface,
} from "./immersive-short-surface";

type ShortDetailReelProps = {
  backHref: string;
  fanTab?: FanHubTab | undefined;
  initialIndex: number;
  initialSurface: DetailShortSurface;
  shortIds: readonly string[];
  source: "creator" | "fan";
};

type ShortDetailItemState =
  | {
      kind: "error";
    }
  | {
      kind: "loading";
    }
  | {
      kind: "ready";
      surface: DetailShortSurface;
    };

const reelStateBaseInsetPx = 76;

function ShortDetailStateShell({
  backHref,
  children,
}: {
  backHref: string;
  children: ReactNode;
}) {
  return (
    <FeedLikeShortBackdrop header={<FeedLikeShortBackHeader backHref={backHref} />}>
      <div
        className="relative flex h-full items-end px-4"
        style={{ paddingBottom: `calc(${reelStateBaseInsetPx + 24}px + env(safe-area-inset-bottom, 0px))` }}
      >
        <div className="w-full bg-gradient-to-t from-black/90 via-black/40 to-transparent px-1 pb-5 pt-16">
          {children}
        </div>
      </div>
    </FeedLikeShortBackdrop>
  );
}

function ShortDetailLoadingState({ backHref }: { backHref: string }) {
  return (
    <ShortDetailStateShell backHref={backHref}>
      <div className="w-full rounded-[28px] border border-white/12 bg-[rgba(7,19,29,0.52)] px-5 py-5 text-white shadow-[0_26px_60px_rgba(5,13,24,0.34)] backdrop-blur-[18px]">
        <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-white/58">short loading</p>
        <p className="mt-3 font-display text-[24px] font-semibold tracking-[-0.05em]">short を準備しています。</p>
        <p className="mt-2 text-[14px] leading-6 text-white/78">次の short surface を読み込んでいます。</p>
      </div>
    </ShortDetailStateShell>
  );
}

function ShortDetailErrorState({ backHref }: { backHref: string }) {
  return (
    <ShortDetailStateShell backHref={backHref}>
      <div className="w-full rounded-[28px] border border-white/12 bg-[rgba(7,19,29,0.52)] px-5 py-5 text-white shadow-[0_26px_60px_rgba(5,13,24,0.34)] backdrop-blur-[18px]">
        <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-white/58">short unavailable</p>
        <h1 className="mt-3 font-display text-[24px] font-semibold tracking-[-0.05em]">
          この short を開けませんでした。
        </h1>
        <p className="mt-2 text-[14px] leading-6 text-white/78">一覧に戻って別の short を選び直してください。</p>
        <div className="mt-5 flex flex-wrap gap-3">
          <Button asChild>
            <Link href={backHref}>一覧に戻る</Link>
          </Button>
        </div>
      </div>
    </ShortDetailStateShell>
  );
}

/**
 * profile/list 起点の short detail reel を表示する。
 */
export function ShortDetailReel({
  backHref,
  fanTab,
  initialIndex,
  initialSurface,
  shortIds,
  source,
}: ShortDetailReelProps) {
  const [activeIndex, setActiveIndex] = useState(initialIndex);
  const [itemStates, setItemStates] = useState<Record<string, ShortDetailItemState>>({
    [initialSurface.short.id]: {
      kind: "ready",
      surface: initialSurface,
    },
  });
  const itemStatesRef = useRef(itemStates);
  const inFlightRef = useRef<Map<string, Promise<void>>>(new Map());
  const isMountedRef = useRef(true);

  useEffect(() => {
    const inFlightTasks = inFlightRef.current;

    // Dev StrictMode re-runs mount effects, so the guard must be restored on setup.
    isMountedRef.current = true;

    return () => {
      isMountedRef.current = false;
      inFlightTasks.clear();
    };
  }, []);

  const setItemState = (shortId: string, nextState: ShortDetailItemState) => {
    itemStatesRef.current = {
      ...itemStatesRef.current,
      [shortId]: nextState,
    };
    setItemStates(itemStatesRef.current);
  };

  useEffect(() => {
    const candidateIndices = [activeIndex - 1, activeIndex, activeIndex + 1].filter(
      (index, position, values) =>
        index >= 0 && index < shortIds.length && values.indexOf(index) === position,
    );

    for (const index of candidateIndices) {
      const shortId = shortIds[index];
      const currentState = shortId ? itemStatesRef.current[shortId] : undefined;

      if (
        !shortId
        || currentState?.kind === "ready"
        || currentState?.kind === "error"
        || inFlightRef.current.has(shortId)
      ) {
        continue;
      }

      setItemState(shortId, {
        kind: "loading",
      });

      const task = (async () => {
        try {
          const detail = await getPublicShortDetail({
            shortId,
          });

          if (!isMountedRef.current) {
            return;
          }

          setItemState(shortId, {
            kind: "ready",
            surface: buildDetailSurfaceFromApi(detail),
          });
        } catch {
          if (!isMountedRef.current) {
            return;
          }

          setItemState(shortId, {
            kind: "error",
          });
        } finally {
          inFlightRef.current.delete(shortId);
        }
      })();

      inFlightRef.current.set(shortId, task);
    }
  }, [activeIndex, shortIds]);

  return (
    <VerticalSnapReel
      getKey={(shortId) => shortId}
      initialIndex={initialIndex}
      items={shortIds}
      onActiveIndexChange={setActiveIndex}
      renderItem={(shortId, { isActive }) => {
        const itemState = itemStates[shortId];

        if (itemState?.kind === "ready") {
          return (
            <ImmersiveShortSurface
              backHref={backHref}
              creatorProfileOrigin={{
                from: "short",
                ...(source === "fan" ? { shortFanTab: fanTab } : {}),
                shortId: itemState.surface.short.id,
              }}
              isActive={isActive}
              mode="detail"
              presentation="feedLike"
              surface={itemState.surface}
            />
          );
        }

        if (itemState?.kind === "error") {
          return <ShortDetailErrorState backHref={backHref} />;
        }

        return <ShortDetailLoadingState backHref={backHref} />;
      }}
    />
  );
}

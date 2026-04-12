"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";

import type { FanHubTab } from "@/entities/fan-profile";
import { getPublicShortDetail } from "@/entities/short";
import {
  Button,
  SurfacePanel,
  VerticalSnapReel,
} from "@/shared/ui";

import { buildDetailSurfaceFromApi } from "../model/api-short-surface";
import type { DetailShortSurface } from "../model/short-surface";
import { ImmersiveShortSurface } from "./immersive-short-surface";

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

function ShortDetailLoadingState() {
  return (
    <section className="absolute inset-0 overflow-hidden bg-[linear-gradient(180deg,#b7e8ff_0%,#68c0eb_22%,#2a648f_56%,#07131d_100%)] text-white">
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.22),transparent_34%)]" />
      <div className="relative flex h-full items-center justify-center px-6 text-center">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.24em] text-white/72">short loading</p>
          <p className="mt-4 text-[18px] font-semibold tracking-[-0.04em]">
            short を準備しています。
          </p>
        </div>
      </div>
    </section>
  );
}

function ShortDetailErrorState({ backHref }: { backHref: string }) {
  return (
    <main className="flex min-h-full items-center justify-center px-6 py-12">
      <SurfacePanel className="w-full max-w-xl px-8 py-9">
        <p className="font-display text-[11px] font-semibold uppercase tracking-[0.24em] text-accent">
          short unavailable
        </p>
        <h1 className="mt-4 font-display text-3xl font-semibold tracking-[-0.05em] text-foreground">
          この short を開けませんでした。
        </h1>
        <p className="mt-3 text-sm leading-7 text-muted">
          一覧に戻って別の short を選び直してください。
        </p>
        <div className="mt-8 flex flex-wrap gap-3">
          <Button asChild>
            <Link href={backHref}>一覧に戻る</Link>
          </Button>
        </div>
      </SurfacePanel>
    </main>
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
              surface={itemState.surface}
            />
          );
        }

        if (itemState?.kind === "error") {
          return <ShortDetailErrorState backHref={backHref} />;
        }

        return <ShortDetailLoadingState />;
      }}
    />
  );
}

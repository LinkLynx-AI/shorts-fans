"use client";

import { useEffect, useRef, useState } from "react";

import type { FanLibraryItem } from "@/entities/fan-profile";
import { getShortThemeStyle } from "@/entities/short";
import { buildFanProfileShortDetailHref } from "@/features/creator-navigation";
import { VerticalSnapReel } from "@/shared/ui";

import { resolveLibraryMainPlaybackSurface } from "../api/resolve-library-main-playback-surface";
import type { MainPlaybackSurface as MainPlaybackSurfaceModel } from "../model/main-playback-surface";
import { MainPlaybackLockedState } from "./main-playback-locked-state";
import { MainPlaybackSurface } from "./main-playback-surface";

type LibraryMainReelProps = {
  backHref: string;
  initialIndex: number;
  items: readonly FanLibraryItem[];
};

type LibraryMainItemState =
  | {
      kind: "error";
    }
  | {
      kind: "locked";
    }
  | {
      kind: "loading";
    }
  | {
      kind: "ready";
      surface: MainPlaybackSurfaceModel;
    };

function getLibraryMainItemKey(item: FanLibraryItem): string {
  return `${item.main.id}:${item.entryShort.id}`;
}

function LibraryMainLoadingState({ item }: { item: FanLibraryItem }) {
  return (
    <section className="absolute inset-0 overflow-hidden text-white" style={getShortThemeStyle(item.entryShort)}>
      <div className="absolute inset-0 bg-[linear-gradient(180deg,var(--short-bg-start)_0%,var(--short-bg-accent)_22%,var(--short-bg-mid)_56%,var(--short-bg-end)_100%)]" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.22),transparent_34%)]" />
      <div className="relative flex h-full items-center justify-center px-6 text-center">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.24em] text-white/72">main loading</p>
          <p className="mt-4 text-[18px] font-semibold tracking-[-0.04em]">
            {item.creator.displayName} の main を準備しています。
          </p>
        </div>
      </div>
    </section>
  );
}

/**
 * fan profile library 起点の main reel を表示する。
 */
export function LibraryMainReel({
  backHref,
  initialIndex,
  items,
}: LibraryMainReelProps) {
  const [activeIndex, setActiveIndex] = useState(initialIndex);
  const [itemStates, setItemStates] = useState<Record<string, LibraryMainItemState>>({});
  const itemStatesRef = useRef<Record<string, LibraryMainItemState>>({});
  const inFlightRef = useRef<Map<string, Promise<void>>>(new Map());
  const isMountedRef = useRef(true);

  useEffect(() => {
    return () => {
      isMountedRef.current = false;
    };
  }, []);

  const setItemState = (key: string, nextState: LibraryMainItemState) => {
    itemStatesRef.current = {
      ...itemStatesRef.current,
      [key]: nextState,
    };
    setItemStates(itemStatesRef.current);
  };

  useEffect(() => {
    const candidateIndices = [activeIndex - 1, activeIndex, activeIndex + 1].filter(
      (index, position, values) =>
        index >= 0 && index < items.length && values.indexOf(index) === position,
    );

    for (const index of candidateIndices) {
      const item = items[index];

      if (!item) {
        continue;
      }

      const itemKey = getLibraryMainItemKey(item);

      if (itemStatesRef.current[itemKey] || inFlightRef.current.has(itemKey)) {
        continue;
      }

      setItemState(itemKey, {
        kind: "loading",
      });

      const task = (async () => {
        try {
          const resolution = await resolveLibraryMainPlaybackSurface(item);

          if (!isMountedRef.current) {
            return;
          }

          setItemState(
            itemKey,
            resolution.kind === "ready"
              ? {
                  kind: "ready",
                  surface: resolution.surface,
                }
              : {
                  kind: "locked",
                },
          );
        } catch {
          if (!isMountedRef.current) {
            return;
          }

          setItemState(itemKey, {
            kind: "error",
          });
        } finally {
          inFlightRef.current.delete(itemKey);
        }
      })();

      inFlightRef.current.set(itemKey, task);
    }
  }, [activeIndex, items]);

  return (
    <VerticalSnapReel
      getKey={(item) => getLibraryMainItemKey(item)}
      initialIndex={initialIndex}
      items={items}
      onActiveIndexChange={setActiveIndex}
      renderItem={(item, { isActive }) => {
        const itemState = itemStates[getLibraryMainItemKey(item)];

        return (
          <>
            {itemState?.kind === "ready" ? (
              <MainPlaybackSurface
                fallbackHref={backHref}
                isActive={isActive}
                surface={itemState.surface}
              />
            ) : itemState?.kind === "locked" ? (
              <MainPlaybackLockedState
                description="この main は short 側の確認導線を通ってから開き直してください。"
                fallbackHref={buildFanProfileShortDetailHref(item.entryShort.id, "library")}
                fallbackLabel="short から開き直す"
                title="この main を開くには確認が必要です。"
              />
            ) : itemState?.kind === "error" ? (
              <MainPlaybackLockedState
                description="library から入り直すか、別の main を試してください。"
                fallbackHref={backHref}
                fallbackLabel="library に戻る"
                title="この main を準備できませんでした。"
              />
            ) : (
              <LibraryMainLoadingState item={item} />
            )}
          </>
        );
      }}
    />
  );
}

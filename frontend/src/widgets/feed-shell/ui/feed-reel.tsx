"use client";

import { useEffect, useRef, useState } from "react";

import { ImmersiveShortSurface } from "@/widgets/immersive-short-surface";

import type { FeedShortSurface } from "@/widgets/immersive-short-surface";
import type { FeedTab } from "@/entities/short";

import { useFeedPinState } from "../model/use-feed-pin-state";

type FeedReelProps = {
  activeTab: FeedTab;
  surfaces: readonly FeedShortSurface[];
};

/**
 * full-screen short surfaces を縦方向の snap scroll で連続視聴させる。
 */
export function FeedReel({ activeTab, surfaces }: FeedReelProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const activeIndexRef = useRef(0);
  const wheelLockRef = useRef(false);
  const [activeIndex, setActiveIndex] = useState(0);
  const { resolvePinState } = useFeedPinState({ surfaces });

  useEffect(() => {
    const container = containerRef.current;

    if (!container) {
      return;
    }

    const handleScroll = () => {
      const nextIndex = Math.round(container.scrollTop / Math.max(container.clientHeight, 1));
      const boundedIndex = Math.min(Math.max(nextIndex, 0), surfaces.length - 1);

      activeIndexRef.current = boundedIndex;
      setActiveIndex((currentIndex) => (currentIndex === boundedIndex ? currentIndex : boundedIndex));
    };

    handleScroll();
    container.addEventListener("scroll", handleScroll, { passive: true });

    return () => {
      container.removeEventListener("scroll", handleScroll);
    };
  }, [surfaces.length]);

  const scrollToIndex = (nextIndex: number) => {
    const container = containerRef.current;

    if (!container) {
      return;
    }

    container.scrollTo({
      behavior: "smooth",
      top: nextIndex * container.clientHeight,
    });
  };

  const handleWheel: React.WheelEventHandler<HTMLDivElement> = (event) => {
    if (surfaces.length <= 1) {
      return;
    }

    if (Math.abs(event.deltaY) < 8) {
      return;
    }

    event.preventDefault();

    if (wheelLockRef.current) {
      return;
    }

    const direction = event.deltaY > 0 ? 1 : -1;
    const nextIndex = Math.min(Math.max(activeIndexRef.current + direction, 0), surfaces.length - 1);

    if (nextIndex === activeIndexRef.current) {
      return;
    }

    wheelLockRef.current = true;
    activeIndexRef.current = nextIndex;
    setActiveIndex(nextIndex);
    scrollToIndex(nextIndex);

    window.setTimeout(() => {
      wheelLockRef.current = false;
    }, 420);
  };

  return (
    <div
      ref={containerRef}
      className="absolute inset-0 snap-y snap-mandatory overflow-y-auto overscroll-y-contain [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden"
      onWheel={handleWheel}
    >
      {surfaces.map((surface, index) => (
        <div key={surface.short.id} className="relative h-dvh min-h-dvh snap-start [scroll-snap-stop:always]">
          <ImmersiveShortSurface
            activeTab={activeTab}
            isActive={index === activeIndex}
            mode="feed"
            pin={resolvePinState(surface)}
            surface={surface}
          />
        </div>
      ))}
    </div>
  );
}

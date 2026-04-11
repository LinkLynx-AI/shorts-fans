"use client";

import { useEffect, useRef, useState } from "react";

import { ImmersiveShortSurface } from "@/widgets/immersive-short-surface";

import type { FeedShortSurface } from "@/widgets/immersive-short-surface";
import type { FeedTab } from "@/entities/short";

type FeedReelProps = {
  activeTab: FeedTab;
  surfaces: readonly FeedShortSurface[];
};

/**
 * full-screen short surfaces を縦方向の snap scroll で連続視聴させる。
 */
export function FeedReel({ activeTab, surfaces }: FeedReelProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const [activeIndex, setActiveIndex] = useState(0);

  useEffect(() => {
    const container = containerRef.current;

    if (!container) {
      return;
    }

    const handleScroll = () => {
      const nextIndex = Math.round(container.scrollTop / Math.max(container.clientHeight, 1));
      const boundedIndex = Math.min(Math.max(nextIndex, 0), surfaces.length - 1);

      setActiveIndex((currentIndex) => (currentIndex === boundedIndex ? currentIndex : boundedIndex));
    };

    handleScroll();
    container.addEventListener("scroll", handleScroll, { passive: true });

    return () => {
      container.removeEventListener("scroll", handleScroll);
    };
  }, [surfaces.length]);

  return (
    <div
      ref={containerRef}
      className="absolute inset-0 snap-y snap-mandatory overflow-y-auto overscroll-y-contain"
    >
      {surfaces.map((surface, index) => (
        <div key={surface.short.id} className="relative h-full snap-start">
          <ImmersiveShortSurface
            activeTab={activeTab}
            isActive={index === activeIndex}
            mode="feed"
            surface={surface}
          />
        </div>
      ))}
    </div>
  );
}

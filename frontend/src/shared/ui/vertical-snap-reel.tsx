"use client";

import type { Key, ReactNode } from "react";
import { useEffect, useLayoutEffect, useRef, useState } from "react";

type VerticalSnapReelRenderArgs = {
  index: number;
  isActive: boolean;
};

type VerticalSnapReelProps<TItem> = {
  className?: string;
  getKey: (item: TItem, index: number) => Key;
  initialIndex?: number;
  items: readonly TItem[];
  onActiveIndexChange?: ((index: number) => void) | undefined;
  renderItem: (item: TItem, args: VerticalSnapReelRenderArgs) => ReactNode;
};

function clampIndex(index: number, length: number): number {
  if (length <= 0) {
    return 0;
  }

  return Math.min(Math.max(index, 0), length - 1);
}

/**
 * 全画面 panel を縦方向の snap scroll で連続表示する。
 */
export function VerticalSnapReel<TItem>({
  className = "absolute inset-0 snap-y snap-mandatory overflow-y-auto overscroll-y-contain [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden",
  getKey,
  initialIndex = 0,
  items,
  onActiveIndexChange,
  renderItem,
}: VerticalSnapReelProps<TItem>) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const boundedInitialIndex = clampIndex(initialIndex, items.length);
  const activeIndexRef = useRef(boundedInitialIndex);
  const wheelLockRef = useRef(false);
  const [activeIndex, setActiveIndex] = useState(boundedInitialIndex);

  useEffect(() => {
    onActiveIndexChange?.(activeIndex);
  }, [activeIndex, onActiveIndexChange]);

  useLayoutEffect(() => {
    const container = containerRef.current;

    if (!container) {
      return;
    }

    activeIndexRef.current = boundedInitialIndex;
    container.scrollTop = boundedInitialIndex * container.clientHeight;
  }, [boundedInitialIndex, items.length]);

  useEffect(() => {
    const container = containerRef.current;

    if (!container) {
      return;
    }

    const handleScroll = () => {
      if (container.clientHeight <= 0) {
        return;
      }

      const nextIndex = Math.round(container.scrollTop / container.clientHeight);
      const boundedIndex = clampIndex(nextIndex, items.length);

      activeIndexRef.current = boundedIndex;
      setActiveIndex((currentIndex) => (currentIndex === boundedIndex ? currentIndex : boundedIndex));
    };

    handleScroll();
    container.addEventListener("scroll", handleScroll, { passive: true });

    return () => {
      container.removeEventListener("scroll", handleScroll);
    };
  }, [items.length]);

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
    if (items.length <= 1) {
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
    const nextIndex = clampIndex(activeIndexRef.current + direction, items.length);

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
      className={className}
      onWheel={handleWheel}
    >
      {items.map((item, index) => (
        <div
          className="relative h-dvh min-h-dvh snap-start [scroll-snap-stop:always]"
          key={getKey(item, index)}
        >
          {renderItem(item, {
            index,
            isActive: index === activeIndex,
          })}
        </div>
      ))}
    </div>
  );
}

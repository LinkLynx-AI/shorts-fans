"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { cn } from "@/shared/lib";

import { getFanNavigationItems, resolveActiveFanNavigation } from "../model/fan-navigation";

/**
 * fan mode 共通の bottom navigation を表示する。
 */
export function FanBottomNavigation() {
  const pathname = usePathname();
  const activeKey = resolveActiveFanNavigation(pathname);

  return (
    <nav
      aria-label="Primary"
      className="grid grid-cols-3 border-t border-border bg-tabbar-surface px-4 pb-[calc(8px+env(safe-area-inset-bottom,0px))] pt-2.5 shadow-[0_-10px_28px_rgba(15,23,42,0.05)]"
    >
      {getFanNavigationItems().map((item) => {
        const isActive = item.key === activeKey;

        return (
          <Link
            key={item.key}
            aria-label={item.ariaLabel}
            aria-current={isActive ? "page" : undefined}
            className={cn(
              "mx-auto inline-flex size-11 items-center justify-center rounded-full transition",
              isActive
                ? "bg-accent-soft text-accent-ink shadow-[inset_0_0_0_1px_rgba(113,180,234,0.18)]"
                : "text-muted hover:bg-surface-subtle hover:text-foreground",
            )}
            href={item.href}
          >
            <item.icon aria-hidden="true" className="size-[19px]" strokeWidth={1.85} />
            <span className="sr-only">{item.ariaLabel}</span>
          </Link>
        );
      })}
    </nav>
  );
}

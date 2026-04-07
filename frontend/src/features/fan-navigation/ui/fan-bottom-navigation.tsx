"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { useHasViewerSession } from "@/entities/viewer";
import { buildFanLoginHref } from "@/features/fan-auth";
import { cn } from "@/shared/lib";

import { getFanNavigationItems, resolveActiveFanNavigation } from "../model/fan-navigation";

/**
 * fan mode 共通の bottom navigation を表示する。
 */
export function FanBottomNavigation() {
  const hasViewerSession = useHasViewerSession();
  const pathname = usePathname();
  const activeKey = resolveActiveFanNavigation(pathname);

  return (
    <nav
      aria-label="Primary"
      className="grid grid-cols-3 border-t border-border/90 bg-tabbar-surface px-2.5 pb-[calc(3px+env(safe-area-inset-bottom,0px))] pt-2 shadow-[0_-12px_28px_rgba(36,94,132,0.08)] backdrop-blur-xl"
    >
      {getFanNavigationItems().map((item) => {
        const isActive = item.key === activeKey;
        // TODO: backend 側で fan profile open 時に auth が要求されるようになったら、
        // `/fan` への一時 login workaround を削除して item.href をそのまま使う。
        const href = item.key === "fan" && !hasViewerSession ? buildFanLoginHref() : item.href;

        return (
          <Link
            key={item.key}
            aria-label={item.ariaLabel}
            aria-current={isActive ? "page" : undefined}
            className={cn(
              "grid min-h-10 place-items-center px-2 py-2 transition",
              isActive ? "text-accent-strong" : "text-accent-strong/72 hover:text-accent-strong/84",
            )}
            href={href}
          >
            <item.icon aria-hidden="true" className="size-[18px]" strokeWidth={1.9} />
            <span className="sr-only">{item.ariaLabel}</span>
          </Link>
        );
      })}
    </nav>
  );
}

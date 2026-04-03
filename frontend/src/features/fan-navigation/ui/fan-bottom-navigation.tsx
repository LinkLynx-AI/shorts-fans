"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { getFanNavigationItems, resolveActiveFanNavigation } from "../model/fan-navigation";
import { cn } from "@/shared/lib";

/**
 * fan mode 共通の bottom navigation を表示する。
 */
export function FanBottomNavigation() {
  const pathname = usePathname();
  const activeKey = resolveActiveFanNavigation(pathname);

  return (
    <nav
      aria-label="Primary"
      className="grid grid-cols-3 border-t border-border/90 bg-tabbar-surface px-2.5 pb-[calc(3px+env(safe-area-inset-bottom,0px))] pt-2 shadow-[0_-12px_28px_rgba(36,94,132,0.08)] backdrop-blur-xl"
    >
      {getFanNavigationItems().map((item) => {
        const isActive = item.key === activeKey;

        return (
          <Link
            key={item.key}
            aria-current={isActive ? "page" : undefined}
            className={cn(
              "grid min-h-[52px] justify-items-center gap-1.5 rounded-[18px] px-2 py-2 text-[11px] font-semibold uppercase tracking-[0.14em] transition",
              isActive ? "text-accent-strong" : "text-accent-strong/72 hover:bg-white/72",
            )}
            href={item.href}
          >
            <span className="grid justify-items-center gap-1.5">
              <span
                className={cn(
                  "inline-flex size-9 items-center justify-center rounded-full border transition",
                  isActive
                    ? "border-accent-strong/18 bg-accent-strong/10"
                    : "border-transparent bg-transparent",
                )}
              >
                <item.icon className="size-[18px]" strokeWidth={1.9} />
              </span>
              <span>{item.label}</span>
            </span>
          </Link>
        );
      })}
    </nav>
  );
}

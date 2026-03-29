"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { cn } from "@/shared/lib";

import { isNavigationItemActive } from "../lib/is-navigation-item-active";
import { consumerNavigationItems } from "../model/navigation-items";

/**
 * consumer shell 配下の下部タブナビゲーションを表示する。
 */
export function BottomTabBar() {
  const pathname = usePathname();

  return (
    <nav
      aria-label="consumer navigation"
      className="pointer-events-none fixed inset-x-0 bottom-0 z-30 flex justify-center px-4 pb-4"
    >
      <div className="pointer-events-auto flex w-full max-w-xl items-center justify-between rounded-[1.75rem] border border-white/75 bg-white/82 px-3 py-2 shadow-[0_28px_80px_rgba(56,24,8,0.22)] backdrop-blur-xl">
        {consumerNavigationItems.map(({ description, href, icon: Icon, label }) => {
          const active = isNavigationItemActive(pathname, href);

          return (
            <Link
              key={href}
              aria-current={active ? "page" : undefined}
              className={cn(
                "group flex min-w-0 flex-1 flex-col items-center gap-1 rounded-[1.2rem] px-2 py-2 text-[0.7rem] font-semibold uppercase tracking-[0.16em] text-muted transition hover:bg-accent/6 hover:text-foreground",
                active && "bg-[#221511] text-white shadow-[0_14px_32px_rgba(34,21,17,0.26)]",
              )}
              href={href}
            >
              <span
                className={cn(
                  "flex size-10 items-center justify-center rounded-2xl border border-transparent bg-white/70 text-foreground transition",
                  active && "bg-white/12 text-white",
                )}
              >
                <Icon className="size-4.5" />
              </span>
              <span>{label}</span>
              <span className={cn("text-[0.6rem] normal-case tracking-normal text-muted", active && "text-stone-300")}>
                {description}
              </span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}

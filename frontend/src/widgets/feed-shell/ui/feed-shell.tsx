import type { ReactNode } from "react";
import Link from "next/link";

import { buildFanLoginHref } from "@/features/fan-auth";
import { cn } from "@/shared/lib";
import { Button, SurfacePanel } from "@/shared/ui";

import type { FeedShellState } from "../model/mock-feed-shell";
import { FeedReel } from "./feed-reel";

type FeedShellProps = {
  state: FeedShellState;
};

const sharedFanNavigationBaseInsetPx = 76;
const sharedFanNavigationInset = `calc(${sharedFanNavigationBaseInsetPx}px + env(safe-area-inset-bottom, 0px))`;

function FeedTabsNavigation({ activeTab }: { activeTab: "following" | "recommended" }) {
  return (
    <nav aria-label="Feed sections" className="flex items-center space-x-6 text-lg font-bold drop-shadow-md">
      {[
        { active: activeTab === "recommended", href: "/?tab=recommended", key: "recommended", label: "For You" },
        { active: activeTab === "following", href: "/?tab=following", key: "following", label: "Following" },
      ].map((item) => (
        <Link
          key={item.key}
          aria-current={item.active ? "page" : undefined}
          className={cn(
            "border-b-2 border-transparent pb-1 text-white/60 transition hover:text-white",
            item.active && "border-white text-white",
          )}
          href={item.href}
        >
          {item.label}
        </Link>
      ))}
    </nav>
  );
}

function FeedTopBar({ activeTab }: { activeTab: "following" | "recommended" }) {
  return (
    <div className="absolute top-0 z-20 flex w-full items-center justify-between px-4 pb-4 pt-14">
      <div className="w-6" />
      <FeedTabsNavigation activeTab={activeTab} />
      <div aria-hidden="true" className="size-11" />
    </div>
  );
}

function FeedShellViewport({ children }: { children: ReactNode }) {
  return (
    <div className="absolute inset-0 overflow-hidden">
      <div className="relative h-full overflow-hidden">{children}</div>
    </div>
  );
}

function FeedFallbackState({
  activeTab,
  ctaHref,
  ctaLabel,
  description,
  title,
}: {
  activeTab: "following" | "recommended";
  ctaHref?: string;
  ctaLabel?: string;
  description: string;
  title: string;
}) {
  return (
    <section className="absolute inset-0 overflow-hidden bg-[linear-gradient(180deg,#8cccf4_0%,#66a8d4_28%,#24405f_62%,#09131e_100%)] text-white">
      <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(255,255,255,0.08)_0%,rgba(255,255,255,0)_24%)]" />
      <div
        className="absolute inset-x-0 h-[46%] bg-[linear-gradient(180deg,rgba(7,19,29,0)_0%,rgba(7,19,29,0.18)_16%,rgba(7,19,29,0.9)_100%)]"
        style={{ bottom: sharedFanNavigationInset }}
      />
      <FeedTopBar activeTab={activeTab} />
      <div className="relative flex h-full items-end px-4" style={{ paddingBottom: `calc(${sharedFanNavigationBaseInsetPx + 24}px + env(safe-area-inset-bottom, 0px))` }}>
        <div className="w-full bg-gradient-to-t from-black/90 via-black/40 to-transparent px-1 pb-5 pt-16">
          <SurfacePanel className="w-full rounded-[28px] border-white/12 bg-[rgba(7,19,29,0.52)] px-5 py-5 text-white shadow-[0_26px_60px_rgba(5,13,24,0.34)] backdrop-blur-[18px]">
            <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-white/58">
              {activeTab === "following" ? "Following feed" : "Recommended feed"}
            </p>
            <h2 className="mt-3 font-display text-[24px] font-semibold tracking-[-0.05em]">{title}</h2>
            <p className="mt-2 text-[14px] leading-6 text-white/78">{description}</p>
            {ctaHref && ctaLabel ? (
              <Button
                asChild
                className="mt-5 h-11 border border-white/18 bg-white text-foreground shadow-[0_16px_32px_rgba(255,255,255,0.18)] hover:bg-white/94"
                size="sm"
              >
                <Link href={ctaHref}>{ctaLabel}</Link>
              </Button>
            ) : null}
          </SurfacePanel>
        </div>
      </div>
    </section>
  );
}

/**
 * fan feed の route shell を表示する。
 */
export function FeedShell({ state }: FeedShellProps) {
  if (state.kind === "ready") {
    return (
      <FeedShellViewport>
        <FeedReel activeTab={state.tab} surfaces={state.surfaces} />
      </FeedShellViewport>
    );
  }

  if (state.kind === "empty") {
    return (
      <FeedShellViewport>
        <FeedFallbackState
          activeTab={state.tab}
          ctaHref="/search"
          ctaLabel="creatorを探す"
          description="気になる creator をフォローすると、新しい short がこの feed に流れてきます。"
          title="フォロー中の creator はまだいません"
        />
      </FeedShellViewport>
    );
  }

  return (
    <FeedShellViewport>
      <FeedFallbackState
        activeTab={state.tab}
        ctaHref={buildFanLoginHref()}
        ctaLabel="ログインへ進む"
        description="ログインすると、フォロー中の short や unlock の続きからそのまま戻れます。"
        title="フォロー中を見るにはログインが必要です"
      />
    </FeedShellViewport>
  );
}

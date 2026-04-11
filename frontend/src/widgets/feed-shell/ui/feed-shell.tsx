import Link from "next/link";

import { buildFanLoginHref } from "@/features/fan-auth";
import { cn } from "@/shared/lib";
import { Button, SurfacePanel } from "@/shared/ui";

import type { FeedShellState } from "../model/mock-feed-shell";
import { FeedReel } from "./feed-reel";

type FeedShellProps = {
  state: FeedShellState;
};

function FeedTabsNavigation({ activeTab }: { activeTab: "following" | "recommended" }) {
  return (
    <nav aria-label="Feed sections" className="inline-flex gap-[18px]">
      {[
        { active: activeTab === "recommended", href: "/?tab=recommended", key: "recommended", label: "おすすめ" },
        { active: activeTab === "following", href: "/?tab=following", key: "following", label: "フォロー中" },
      ].map((item) => (
        <Link
          key={item.key}
          aria-current={item.active ? "page" : undefined}
          className={cn(
            "relative pb-1.5 text-[15px] font-bold tracking-[0] text-white/62 transition hover:text-white/84",
            item.active && "text-white after:absolute after:inset-x-0 after:bottom-0 after:h-0.5 after:rounded-full after:bg-white",
          )}
          href={item.href}
        >
          {item.label}
        </Link>
      ))}
    </nav>
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
    <section className="absolute inset-0 overflow-hidden bg-[linear-gradient(180deg,#94e0ff_0%,#68c0eb_22%,#2a648f_56%,#07131d_100%)] text-white">
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.22),transparent_34%)]" />
      <div className="relative flex h-full flex-col">
        <div className="flex justify-center px-4 pt-6">
          <FeedTabsNavigation activeTab={activeTab} />
        </div>
        <div className="flex flex-1 items-center px-4 pb-24">
          <SurfacePanel className="w-full px-5 py-5 text-foreground">
            <h2 className="font-display text-xl font-semibold tracking-[-0.04em]">{title}</h2>
            <p className="mt-2 text-sm leading-6 text-muted">{description}</p>
            {ctaHref && ctaLabel ? (
              <Button asChild className="mt-4" size="sm">
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
    return <FeedReel activeTab={state.tab} surfaces={state.surfaces} />;
  }

  if (state.kind === "empty") {
    return (
      <FeedFallbackState
        activeTab={state.tab}
        ctaHref="/search"
        ctaLabel="creatorを探す"
        description="following feed は 200 empty を返せる前提なので、ここで空状態を受けられるようにしています。実際の copy と CTA は後続 task で詰めます。"
        title="フォロー中の creator はまだいません"
      />
    );
  }

  return (
    <FeedFallbackState
      activeTab={state.tab}
      ctaHref={buildFanLoginHref()}
      ctaLabel="ログインへ進む"
      description="following feed が auth_required を返したときは、この entry から fan login へ進めるようにしています。"
      title="フォロー中を見るにはログインが必要です"
    />
  );
}

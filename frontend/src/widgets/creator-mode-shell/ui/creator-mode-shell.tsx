import Link from "next/link";
import {
  Plus,
  Settings2,
} from "lucide-react";

import { CreatorAvatar } from "@/entities/creator";
import { CreatorModePrimaryNavigation } from "@/features/creator-mode-navigation";
import { Button, SurfacePanel } from "@/shared/ui";
import {
  RouteStructurePanel,
} from "@/widgets/route-structure-panel";

import type { CreatorModeShellState } from "../model/creator-mode-shell";

function CreatorModeFrame({ children }: { children: React.ReactNode }) {
  return (
    <main className="min-h-svh bg-[radial-gradient(circle_at_top,rgba(214,242,247,0.82),transparent_34%),linear-gradient(180deg,#f7fcfd_0%,#eef7f8_42%,#e8eff6_100%)] text-foreground">
      {children}
    </main>
  );
}

function CreatorUtilityButton({
  ariaLabel,
  icon: Icon,
}: {
  ariaLabel: string;
  icon: typeof Plus;
}) {
  return (
    <button
      aria-label={ariaLabel}
      className="inline-flex size-10 items-center justify-center rounded-full border border-white/74 bg-white/78 text-[#0f566a] shadow-[0_12px_24px_rgba(36,94,132,0.1)] backdrop-blur-md disabled:cursor-not-allowed disabled:opacity-72"
      disabled
      type="button"
    >
      <Icon aria-hidden="true" className="size-4.5" strokeWidth={1.9} />
    </button>
  );
}

function CreatorShellBlockedState({ state }: { state: Exclude<CreatorModeShellState, { kind: "ready" }> }) {
  return (
    <CreatorModeFrame>
      <div className="mx-auto flex min-h-svh max-w-3xl items-center px-4 py-12 sm:px-6">
        <SurfacePanel className="w-full px-6 py-7 sm:px-8 sm:py-9">
          <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-[#0f6172]">{state.eyebrow}</p>
          <h1 className="mt-4 font-display text-[30px] font-semibold tracking-[-0.05em] text-foreground">
            {state.title}
          </h1>
          <p className="mt-3 text-sm leading-7 text-muted">{state.description}</p>
          <div className="mt-8 flex flex-wrap gap-3">
            <Button asChild>
              <Link href={state.ctaHref}>{state.ctaLabel}</Link>
            </Button>
          </div>
        </SurfacePanel>
      </div>
    </CreatorModeFrame>
  );
}

/**
 * `/creator` の route shell を表示する。
 */
export function CreatorModeShell({ state }: { state: CreatorModeShellState }) {
  if (state.kind !== "ready") {
    return <CreatorShellBlockedState state={state} />;
  }

  return (
    <CreatorModeFrame>
      <div className="mx-auto max-w-[1120px] px-4 py-5 sm:px-6 lg:px-8">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="min-w-0">
            <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-[#0f6172]">{state.eyebrow}</p>
            <h1 className="mt-3 font-display text-[34px] font-semibold tracking-[-0.06em] text-foreground">
              {state.title}
            </h1>
            <p className="mt-3 max-w-[68ch] text-sm leading-7 text-muted">{state.description}</p>
          </div>
          <div className="flex items-center gap-2">
            <CreatorUtilityButton ariaLabel="動画を追加" icon={Plus} />
            <CreatorUtilityButton ariaLabel="Account menu" icon={Settings2} />
          </div>
        </div>

        <SurfacePanel className="mt-6 overflow-hidden px-5 py-5 sm:px-6 sm:py-6">
          <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-start">
            <div className="flex items-start gap-4">
              <CreatorAvatar
                className="size-[80px] rounded-full border-white/78 shadow-[0_14px_28px_rgba(36,92,129,0.16)] sm:size-[88px]"
                creator={state.creator}
              />
              <div className="min-w-0">
                <p className="text-[12px] font-semibold uppercase tracking-[0.14em] text-[#0f6172]">
                  {state.creator.handle}
                </p>
                <h2 className="mt-1 text-[24px] font-semibold tracking-[-0.04em] text-foreground">
                  {state.creator.displayName}
                </h2>
                <p className="mt-3 max-w-[56ch] text-sm leading-7 text-muted">{state.creator.bio}</p>
              </div>
            </div>

            <div className="grid gap-2 sm:grid-cols-3 lg:grid-cols-1 xl:grid-cols-3">
              {state.contextBadges.map((badge) => (
                <div
                  className="rounded-[18px] border border-[#dbeff2] bg-[#f7fcfc] px-4 py-3"
                  key={badge.key}
                >
                  <p className="text-[10px] font-semibold uppercase tracking-[0.18em] text-[#5b7b88]">
                    {badge.label}
                  </p>
                  <p className="mt-2 text-sm font-semibold text-foreground">{badge.value}</p>
                </div>
              ))}
            </div>
          </div>
        </SurfacePanel>

        <CreatorModePrimaryNavigation activeKey={state.activeNavigation} className="mt-4" />

        <div className="mt-4 grid gap-4 xl:grid-cols-[320px_minmax(0,1fr)]">
          <SurfacePanel className="px-5 py-5">
            <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-[#0f6172]">Shell role</p>
            <h2 className="mt-3 font-display text-[24px] font-semibold tracking-[-0.05em] text-foreground">
              creator private surface の受け皿
            </h2>
            <p className="mt-3 text-sm leading-7 text-muted">
              この route では fan mode と切り分けた shell と navigation だけを確定し、後続 issue が各 workspace
              surface を載せられるようにしています。
            </p>

            <div className="mt-5 grid gap-3">
              {state.slots.map((slot) => (
                <div className="rounded-[20px] border border-[#dbeff2] bg-[#fbfdfd] px-4 py-4" key={slot.key}>
                  <p className="text-sm font-semibold tracking-[-0.02em] text-foreground">{slot.label}</p>
                  <p className="mt-2 text-sm leading-6 text-muted">{slot.description}</p>
                </div>
              ))}
            </div>
          </SurfacePanel>

          <RouteStructurePanel
            description="この PR では `/creator` 自体だけを実装し、approved workspace content や mode switch entry は別 issue / 別 PR に残します。"
            eyebrow="Route structure"
            items={state.structureItems}
            title="creator route shell"
          />
        </div>
      </div>
    </CreatorModeFrame>
  );
}

"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";

import { CreatorAvatar } from "@/entities/creator";
import { getShortThemeStyle } from "@/entities/short";
import { cn } from "@/shared/lib";
import { Button } from "@/shared/ui";

import {
  getMainPlaybackStatusCopy,
  getMainPlaybackStatusMeta,
  getMainPlaybackStatusTitle,
  type MainPlaybackSurface,
} from "../model/main-playback-surface";

export type MainPlaybackSurfaceProps = {
  fallbackHref: string;
  surface: MainPlaybackSurface;
};

function PinRail({ pinned }: { pinned: boolean }) {
  const label = pinned ? "Pinned short" : "Pin short";

  return (
    <div className="absolute right-4 z-20 flex flex-col items-center gap-2.5" style={{ bottom: "204px" }}>
      <button
        aria-label={label}
        aria-pressed={pinned}
        className={cn(
          "inline-flex size-11 items-center justify-center rounded-full bg-transparent p-0 text-accent-strong/72 transition hover:text-accent",
          pinned && "text-accent",
        )}
        type="button"
      >
        <svg
          aria-hidden="true"
          className="size-[22px]"
          fill={pinned ? "currentColor" : "none"}
          viewBox="0 0 14 18"
        >
          <path
            d="M3 1.75h8a1 1 0 0 1 1 1V16L7 12.9 2 16V2.75a1 1 0 0 1 1-1Z"
            stroke="currentColor"
            strokeLinejoin="round"
            strokeWidth="1.7"
          />
        </svg>
        <span className="sr-only">{label}</span>
      </button>
    </div>
  );
}

/**
 * unlock 後の main 継続視聴 surface を表示する。
 */
export function MainPlaybackSurface({ fallbackHref, surface }: MainPlaybackSurfaceProps) {
  const router = useRouter();
  const statusTitle = getMainPlaybackStatusTitle(surface);
  const statusCopy = getMainPlaybackStatusCopy(surface);
  const statusMeta = getMainPlaybackStatusMeta(surface);
  const continuationCopy = surface.entryShort ? `${surface.entryShort.title} の続き。` : "short の続きから再生中。";

  const handleBack = () => {
    if (window.history.length > 1) {
      router.back();
      return;
    }

    router.push(fallbackHref);
  };

  return (
    <section className="absolute inset-0 overflow-hidden text-white" style={getShortThemeStyle(surface.themeShort)}>
      <div className="absolute inset-0 bg-[linear-gradient(180deg,var(--short-bg-start)_0%,var(--short-bg-accent)_22%,var(--short-bg-mid)_56%,var(--short-bg-end)_100%)]" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.22),transparent_34%)]" />

      <div className="relative h-full">
        <h1 className="sr-only">{surface.main.title}</h1>
        <div className="relative z-10 px-4 pt-4">
          <Button
            aria-label="Back"
            className="text-white hover:bg-white/16 hover:text-white"
            onClick={handleBack}
            size="icon"
            type="button"
            variant="ghost"
          >
            <ArrowLeft className="size-5" strokeWidth={2.1} />
          </Button>
        </div>

        <PinRail pinned={surface.viewer.isPinned} />

        <div className="absolute inset-x-4 z-20" style={{ bottom: "152px" }}>
          <div className="flex min-h-12 items-center justify-between gap-3 rounded-full border border-[#85cdf1]/92 bg-[linear-gradient(90deg,rgba(225,244,255,0.98),rgba(204,235,252,0.96))] px-3 py-1.5 text-left text-foreground shadow-[0_18px_44px_rgba(36,94,132,0.14)] backdrop-blur-xl">
            <span className="flex min-w-0 flex-1 items-center gap-3">
              <span className="inline-flex min-h-[34px] shrink-0 items-center rounded-full bg-accent-strong px-3.5 text-xs font-semibold tracking-[-0.01em] text-white">
                Play
              </span>
              <span className="min-w-0">
                <span className="block truncate text-[15px] font-semibold tracking-[-0.01em]">{statusTitle}</span>
                <span className="mt-0.5 block truncate text-[12px] text-muted">{statusCopy}</span>
              </span>
            </span>
            <span className="inline-flex min-h-[34px] shrink-0 items-center rounded-full bg-accent-strong/12 px-3.5 text-xs font-semibold tracking-[-0.01em] text-accent-strong">
              {statusMeta}
            </span>
          </div>
        </div>

        <div className="absolute inset-x-0 bottom-0 z-10 px-4" style={{ paddingBottom: "68px" }}>
          <div className="w-[min(88%,344px)] max-w-[344px]">
            <div className="flex w-fit max-w-full items-center gap-2.5">
              <Link
                className="inline-flex min-w-0 items-center gap-2 text-left text-white transition hover:opacity-90"
                href={`/creators/${surface.creator.id}`}
              >
                <CreatorAvatar
                  className="size-[38px] rounded-full border-white/68 shadow-[0_8px_20px_rgba(7,19,29,0.2)]"
                  creator={surface.creator}
                />
                <span className="truncate text-[15px] font-bold text-white">{surface.creator.displayName}</span>
              </Link>
            </div>
            <p className="mt-0.5 text-[14px] leading-[1.45] text-white/92">{continuationCopy}</p>
          </div>
        </div>
      </div>
    </section>
  );
}

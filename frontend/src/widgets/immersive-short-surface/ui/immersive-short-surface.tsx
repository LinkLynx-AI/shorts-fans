import Link from "next/link";
import { ArrowLeft } from "lucide-react";

import { getShortThemeStyle, type FeedTab, type ShortPreviewMeta } from "@/entities/short";
import { UnlockCta } from "@/features/unlock-entry";
import { cn } from "@/shared/lib";
import { Button } from "@/shared/ui";

import type { DetailShortSurface, FeedShortSurface } from "../model/mock-short-surface";

export type ImmersiveShortSurfaceProps =
  | {
      activeTab: FeedTab;
      mode: "feed";
      surface: FeedShortSurface;
    }
  | {
      backHref: string;
      mode: "detail";
      surface: DetailShortSurface;
    };

function ShortSurfaceHeader(props: ImmersiveShortSurfaceProps) {
  if (props.mode === "feed") {
    return (
      <div className="relative z-10 flex justify-center px-4 pt-6">
        <nav aria-label="Feed sections" className="inline-flex gap-[18px]">
          {[
            { active: props.activeTab === "recommended", href: "/?tab=recommended", key: "recommended", label: "おすすめ" },
            { active: props.activeTab === "following", href: "/?tab=following", key: "following", label: "フォロー中" },
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
      </div>
    );
  }

  return (
    <div className="relative z-10 px-4 pt-4">
      <Button asChild className="text-white hover:bg-white/16 hover:text-white" size="icon" variant="ghost">
        <Link aria-label="Back" href={props.backHref}>
          <ArrowLeft className="size-5" strokeWidth={2.1} />
        </Link>
      </Button>
    </div>
  );
}

type PinRailProps = {
  pinned: boolean;
};

function PinRail({ pinned }: PinRailProps) {
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

type CreatorBlockProps = {
  creator: FeedShortSurface["creator"];
  followed?: boolean;
  short: ShortPreviewMeta;
};

function FeedCreatorAvatar() {
  return (
    <span
      aria-hidden="true"
      className="h-[38px] w-[38px] shrink-0 rounded-full bg-[linear-gradient(180deg,#a7e8ff_0%,#5ba9d4_56%,#17374f_100%)] shadow-[0_8px_20px_rgba(7,19,29,0.2)]"
    />
  );
}

function CreatorBlock({ creator, followed = false, short }: CreatorBlockProps) {
  return (
    <div className="absolute inset-x-0 bottom-0 z-10 px-4" style={{ paddingBottom: "68px" }}>
      <div className="w-[min(88%,344px)] max-w-[344px]">
        <div className="flex w-fit max-w-full items-center gap-2.5">
          <Link
            className="inline-flex min-w-0 items-center gap-2 text-left text-white transition hover:opacity-90"
            href={`/creators/${creator.id}`}
          >
            <FeedCreatorAvatar />
            <span className="truncate text-[15px] font-bold text-white">{creator.name}</span>
          </Link>
          <button
            className={cn(
              "min-h-7 shrink-0 rounded-full border border-white/62 bg-transparent px-3 text-[11px] font-semibold text-white/92 transition",
              followed && "border-[#b6eaff]/78 text-[#d7f5ff]",
            )}
            type="button"
          >
            {followed ? "Following" : "Follow"}
          </button>
        </div>
        <p className="mt-0.5 text-[14px] leading-[1.45] text-white/92">{short.caption}</p>
      </div>
    </div>
  );
}

/**
 * `feed` と `short detail` で共有する immersive short surface を表示する。
 */
export function ImmersiveShortSurface(props: ImmersiveShortSurfaceProps) {
  const { mode, surface } = props;
  const { creator, short, unlockCta, viewer } = surface;
  const followed = mode === "detail" ? viewer.isFollowingCreator : undefined;
  const pinned = viewer.isPinned;

  return (
    <section className="absolute inset-0 overflow-hidden text-white" style={getShortThemeStyle(short)}>
      <div className="absolute inset-0 bg-[linear-gradient(180deg,var(--short-bg-start)_0%,var(--short-bg-accent)_22%,var(--short-bg-mid)_56%,var(--short-bg-end)_100%)]" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.22),transparent_34%)]" />

      <div className="relative h-full">
        <h1 className="sr-only">{mode === "feed" ? "Feed" : "Short detail"}</h1>
        <ShortSurfaceHeader {...props} />
        <PinRail pinned={pinned} />
        <div className="absolute inset-x-4 z-20" style={{ bottom: "152px" }}>
          <UnlockCta
            cta={unlockCta}
            className="w-full"
            {...(mode === "feed" ? { href: `/shorts/${short.id}` } : {})}
          />
        </div>
        <CreatorBlock creator={creator} followed={followed} short={short} />
      </div>
    </section>
  );
}

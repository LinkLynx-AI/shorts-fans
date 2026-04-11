"use client";

import Link from "next/link";
import type { CSSProperties } from "react";
import { useEffect, useRef, useState } from "react";
import { ArrowLeft } from "lucide-react";
import { useRouter } from "next/navigation";

import {
  CreatorAvatar,
  updateCreatorFollow,
  useCreatorFollowToggle,
  getCreatorInitials,
} from "@/entities/creator";
import { getShortThemeStyle, type FeedTab, type ShortPreviewMeta } from "@/entities/short";
import { useHasViewerSession } from "@/entities/viewer";
import { buildCreatorProfileHref } from "@/features/creator-navigation";
import {
  buildFanLoginHref,
  isAuthRequiredResponse,
  useFanAuthDialog,
} from "@/features/fan-auth";
import { getUnlockEntryAction, UnlockCta, UnlockPaywallDialog } from "@/features/unlock-entry";
import { cn } from "@/shared/lib";
import { Button } from "@/shared/ui";

import type { DetailShortSurface, FeedShortSurface } from "../model/short-surface";

const feedSurfaceStyle = {
  "--short-bg-accent": "#68c0eb",
  "--short-bg-end": "#07131d",
  "--short-bg-mid": "#2a648f",
  "--short-bg-start": "#94e0ff",
  "--short-tile-bottom": "#0f2234",
  "--short-tile-mid": "#4cc0eb",
  "--short-tile-top": "#d8f3ff",
} as CSSProperties;

export type ImmersiveShortSurfaceProps =
  | {
      activeTab: FeedTab;
      isActive?: boolean;
      mode: "feed";
      surface: FeedShortSurface;
    }
  | {
      backHref: string;
      isActive?: boolean;
      mode: "detail";
      surface: DetailShortSurface;
    };

/**
 * feed/detail 共通の header 領域を表示する。
 */
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

/**
 * short の pin 状態を表す操作レールを表示する。
 */
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
  followState?:
    | {
        errorMessage: string | null;
        isFollowing: boolean;
        isPending: boolean;
        onToggle: () => void;
      }
    | undefined;
  followed?: boolean | undefined;
  profileHref: string;
  short: ShortPreviewMeta;
};

/**
 * feed surface 用の creator avatar を表示し、avatar 不在時は initials fallback を描画する。
 */
function FeedCreatorAvatar({ creator }: Pick<CreatorBlockProps, "creator">) {
  if (!creator.avatar) {
    return (
      <span
        aria-hidden="true"
        className="flex h-[38px] w-[38px] shrink-0 items-center justify-center rounded-full bg-[linear-gradient(180deg,#b2ecff_0%,#65bae0_56%,#1b4362_100%)] text-[11px] font-semibold uppercase tracking-[0.08em] text-white shadow-[0_8px_20px_rgba(7,19,29,0.2)]"
      >
        {getCreatorInitials(creator.displayName)}
      </span>
    );
  }

  return (
    <CreatorAvatar
      className="size-[38px] rounded-full border-white/68 shadow-[0_8px_20px_rgba(7,19,29,0.2)]"
      creator={creator}
    />
  );
}

/**
 * creator 名、follow 状態、caption をまとめた下部 creator block を表示する。
 */
function CreatorBlock({ creator, followState, followed = false, profileHref, short }: CreatorBlockProps) {
  const caption = short.caption.trim();
  const interactiveFollowState = followState ?? null;
  const resolvedIsFollowing = followState?.isFollowing ?? followed;
  const followLabel = followState
    ? followState.isPending
      ? followState.isFollowing
        ? "Unfollowing..."
        : "Following..."
      : followState.isFollowing
        ? "Following"
        : "Follow"
    : followed
      ? "Following"
      : "Follow";
  const followCtaClassName = cn(
    "min-h-7 shrink-0 rounded-full border border-white/62 bg-transparent px-3 text-[11px] font-semibold text-white/92 transition",
    resolvedIsFollowing && "border-[#b6eaff]/78 text-[#d7f5ff]",
  );

  return (
    <div className="absolute inset-x-0 bottom-0 z-10 px-4" style={{ paddingBottom: "68px" }}>
      <div className="w-[min(88%,344px)] max-w-[344px]">
        <div className="flex w-fit max-w-full items-center gap-2.5">
          <Link
            className="inline-flex min-w-0 items-center gap-2 text-left text-white transition hover:opacity-90"
            href={profileHref}
          >
            <FeedCreatorAvatar creator={creator} />
            <span className="truncate text-[15px] font-bold text-white">{creator.displayName}</span>
          </Link>
          {interactiveFollowState ? (
            <button
              aria-busy={interactiveFollowState.isPending || undefined}
              aria-pressed={resolvedIsFollowing}
              className={followCtaClassName}
              disabled={interactiveFollowState.isPending}
              onClick={interactiveFollowState.onToggle}
              type="button"
            >
              {followLabel}
            </button>
          ) : (
            <span className={followCtaClassName}>{followLabel}</span>
          )}
        </div>
        {caption ? <p className="mt-0.5 text-[14px] leading-[1.45] text-white/92">{caption}</p> : null}
        {followState?.errorMessage ? (
          <p
            className="mt-3 rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
            role="alert"
          >
            {followState.errorMessage}
          </p>
        ) : null}
      </div>
    </div>
  );
}

type FeedCreatorBlockProps = Omit<CreatorBlockProps, "followState" | "followed"> & {
  hasViewerSession: boolean;
  initialIsFollowing: boolean;
};

function FeedCreatorBlock({
  creator,
  hasViewerSession,
  initialIsFollowing,
  profileHref,
  short,
}: FeedCreatorBlockProps) {
  const { openFanAuthDialog } = useFanAuthDialog();
  const { errorMessage, isFollowing, isPending, toggleFollow } = useCreatorFollowToggle({
    creatorId: creator.id,
    hasViewerSession,
    initialIsFollowing,
    onAuthRequired: () => {
      openFanAuthDialog();
    },
    onUnauthenticated: () => {
      openFanAuthDialog();
    },
    updateFollow: updateCreatorFollow,
  });

  return (
    <CreatorBlock
      creator={creator}
      followState={{
        errorMessage,
        isFollowing,
        isPending,
        onToggle: () => {
          void toggleFollow();
        },
      }}
      profileHref={profileHref}
      short={short}
    />
  );
}

/**
 * `feed` と `short detail` で共有する immersive short surface を表示する。
 */
export function ImmersiveShortSurface(props: ImmersiveShortSurfaceProps) {
  const [isHydrated, setIsHydrated] = useState(false);
  const [acceptAge, setAcceptAge] = useState(false);
  const [acceptTerms, setAcceptTerms] = useState(false);
  const [isPaywallOpen, setIsPaywallOpen] = useState(false);
  const [isSubmittingMainAccess, setIsSubmittingMainAccess] = useState(false);
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const hasViewerSession = useHasViewerSession();
  const router = useRouter();
  const { mode, surface } = props;
  const { creator, short, unlock, viewer } = surface;
  const isActive = props.isActive ?? true;
  const pinned = viewer.isPinned;
  const unlockAction = getUnlockEntryAction(unlock);
  const surfaceStyle = mode === "feed" ? feedSurfaceStyle : getShortThemeStyle(short);
  const profileHref =
    mode === "feed"
      ? buildCreatorProfileHref(creator.id, {
          from: "feed",
          tab: props.activeTab,
        })
      : buildCreatorProfileHref(creator.id, {
          from: "short",
          shortId: short.id,
        });

  const resetPaywallState = () => {
    setAcceptAge(false);
    setAcceptTerms(false);
    setIsPaywallOpen(false);
  };

  useEffect(() => {
    setIsHydrated(true);
  }, []);

  useEffect(() => {
    const video = videoRef.current;

    if (!video) {
      return;
    }

    if (!isActive) {
      video.pause();
      return;
    }

    const playPromise = video.play();

    if (playPromise) {
      void playPromise.catch(() => {
        video.muted = true;
      });
    }
  }, [isActive, short.media.url]);

  /**
   * setup-required main の setup dialog を開く前に確認状態を初期化する。
   */
  const handleOpenPaywall = () => {
    if (!hasViewerSession) {
      router.push(buildFanLoginHref());
      return;
    }

    resetPaywallState();
    setIsPaywallOpen(true);
  };

  /**
   * setup dialog を閉じる。送信中は多重操作を防ぐ。
   */
  const handleClosePaywall = () => {
    if (isSubmittingMainAccess) {
      return;
    }

    resetPaywallState();
  };

  /**
   * main access entry を叩いて main playback へ遷移する。
   */
  const handleOpenMain = async () => {
    if (!hasViewerSession) {
      router.push(buildFanLoginHref());
      return;
    }

    if (isSubmittingMainAccess) {
      return;
    }

    setIsSubmittingMainAccess(true);

    try {
      const response = await fetch(unlock.mainAccessEntry.routePath, {
        body: JSON.stringify({
          acceptedAge: acceptAge,
          acceptedTerms: acceptTerms,
          entryToken: unlock.mainAccessEntry.token,
          fromShortId: short.id,
        }),
        headers: {
          "Content-Type": "application/json",
        },
        method: "POST",
      });
      const payload = (await response.json().catch(() => null)) as
        | {
            data?: {
              href?: string;
            } | null;
            error?: {
              code: string;
              message: string;
            };
          }
        | null;

      if (!response.ok && isAuthRequiredResponse(payload)) {
        resetPaywallState();
        router.push(buildFanLoginHref());
        return;
      }

      if (response.ok && payload?.data?.href) {
        resetPaywallState();
        router.push(payload.data.href);
        return;
      }

      router.push(`/shorts/${short.id}`);
    } finally {
      setIsSubmittingMainAccess(false);
    }
  };

  return (
    <section className="absolute inset-0 overflow-hidden text-white" style={surfaceStyle}>
      <div className="absolute inset-0 bg-[linear-gradient(180deg,var(--short-bg-start)_0%,var(--short-bg-accent)_22%,var(--short-bg-mid)_56%,var(--short-bg-end)_100%)]" />
      <video
        ref={videoRef}
        aria-hidden="true"
        autoPlay={isActive}
        className="absolute inset-0 size-full object-cover"
        loop
        muted
        playsInline
        poster={short.media.posterUrl ?? undefined}
        preload={isActive ? "auto" : "metadata"}
        src={short.media.url}
      />
      <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(6,21,33,0.08)_0%,rgba(6,21,33,0.18)_20%,rgba(6,21,33,0.36)_58%,rgba(6,21,33,0.74)_100%)]" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.18),transparent_34%)]" />

      <div className="relative h-full">
        <h1 className="sr-only">{mode === "feed" ? "Feed" : "Short detail"}</h1>
        <ShortSurfaceHeader {...props} />
        <PinRail pinned={pinned} />
        <div className="absolute inset-x-4 z-20" style={{ bottom: "152px" }}>
          <UnlockCta
            className="w-full"
            cta={unlock.unlockCta}
            disabled={!isHydrated || isSubmittingMainAccess}
            {...(surface.mainEntryEnabled && unlockAction === "open_main"
              ? { onClick: handleOpenMain }
              : surface.mainEntryEnabled && unlockAction === "open_paywall"
                ? { onClick: handleOpenPaywall }
                : {})}
          />
        </div>
        {mode === "feed" ? (
          <FeedCreatorBlock
            creator={creator}
            hasViewerSession={hasViewerSession}
            initialIsFollowing={viewer.isFollowingCreator}
            profileHref={profileHref}
            short={short}
          />
        ) : (
          <CreatorBlock creator={creator} followed={viewer.isFollowingCreator} profileHref={profileHref} short={short} />
        )}
        <UnlockPaywallDialog
          acceptAge={acceptAge}
          acceptTerms={acceptTerms}
          isSubmitting={isSubmittingMainAccess}
          onAcceptAgeChange={setAcceptAge}
          onAcceptTermsChange={setAcceptTerms}
          onClose={handleClosePaywall}
          onConfirm={handleOpenMain}
          open={isPaywallOpen}
          unlock={unlock}
        />
      </div>
    </section>
  );
}

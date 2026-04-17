"use client";

import Link from "next/link";
import type { CSSProperties, KeyboardEvent as ReactKeyboardEvent, MouseEvent as ReactMouseEvent, ReactNode } from "react";
import { useEffect, useRef, useState } from "react";
import { ArrowLeft } from "lucide-react";
import { useRouter } from "next/navigation";

import {
  CreatorAvatar,
  updateCreatorFollow,
  useCreatorFollowToggle,
  getCreatorInitials,
} from "@/entities/creator";
import { getShortThemeStyle, type FeedTab } from "@/entities/short";
import { useCurrentViewer, useHasViewerSession } from "@/entities/viewer";
import {
  buildCreatorProfileHref,
  type CreatorProfileRouteOrigin,
} from "@/features/creator-navigation";
import {
  isAuthRequiredApiError,
  isAuthRequiredResponse,
  isFreshAuthRequiredApiError,
  isFreshAuthRequiredResponse,
  useFanAuthDialogControls,
} from "@/features/fan-auth";
import {
  getUnlockEntryAction,
  requestMainAccessEntry,
  requestUnlockSurfaceByShortId,
  type UnlockSurfaceModel,
  UnlockCta,
  UnlockPaywallDialog,
} from "@/features/unlock-entry";
import { cn } from "@/shared/lib";
import { Button } from "@/shared/ui";

import type { DetailShortSurface, FeedShortSurface } from "../model/short-surface";
import { useShortPinState } from "../model/use-short-pin-state";

const feedSurfaceStyle = {
  "--short-bg-accent": "#68c0eb",
  "--short-bg-end": "#07131d",
  "--short-bg-mid": "#2a648f",
  "--short-bg-start": "#94e0ff",
  "--short-tile-bottom": "#0f2234",
  "--short-tile-mid": "#4cc0eb",
  "--short-tile-top": "#d8f3ff",
} as CSSProperties;

const sharedFanNavigationBaseInsetPx = 76;
const feedActionRailOffsetPx = 152;
const feedPinErrorOffsetPx = feedActionRailOffsetPx + 116;
const sharedFanNavigationInset = `calc(${sharedFanNavigationBaseInsetPx}px + env(safe-area-inset-bottom, 0px))`;
const feedActionRailBottom = `calc(${sharedFanNavigationBaseInsetPx + feedActionRailOffsetPx}px + env(safe-area-inset-bottom, 0px))`;
const feedPinErrorBottom = `calc(${sharedFanNavigationBaseInsetPx + feedPinErrorOffsetPx}px + env(safe-area-inset-bottom, 0px))`;

export function FeedLikeShortBackHeader({ backHref }: { backHref: string }) {
  return (
    <div className="absolute top-0 z-20 flex w-full items-center justify-between px-4 pb-4 pt-14">
      <Button asChild className="text-white hover:bg-white/16 hover:text-white" size="icon" variant="ghost">
        <Link aria-label="Back" href={backHref}>
          <ArrowLeft className="size-5" strokeWidth={2.1} />
        </Link>
      </Button>
      <div className="flex-1" />
      <div aria-hidden="true" className="size-11" />
    </div>
  );
}

export function FeedLikeShortBackdrop({
  children,
  header,
  mediaLayer,
}: {
  children: ReactNode;
  header: ReactNode;
  mediaLayer?: ReactNode;
}) {
  return (
    <section className="absolute inset-0 overflow-hidden text-white" style={feedSurfaceStyle}>
      <div className="absolute inset-0 bg-[linear-gradient(180deg,var(--short-bg-start)_0%,var(--short-bg-accent)_22%,var(--short-bg-mid)_56%,var(--short-bg-end)_100%)]" />
      {mediaLayer}
      <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(3,10,18,0.08)_0%,rgba(3,10,18,0.02)_28%,rgba(3,10,18,0.34)_62%,rgba(3,10,18,0.86)_100%)]" />
      <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(255,255,255,0.08)_0%,rgba(255,255,255,0)_22%)]" />
      <div
        className="absolute inset-x-0 h-[46%] bg-[linear-gradient(180deg,rgba(7,19,29,0)_0%,rgba(7,19,29,0.18)_16%,rgba(7,19,29,0.9)_100%)]"
        style={{ bottom: sharedFanNavigationInset }}
      />

      <div className="relative h-full">
        {header}
        {children}
      </div>
    </section>
  );
}

function clampPlaybackProgress(progress: number) {
  if (!Number.isFinite(progress)) {
    return 0;
  }

  if (progress <= 0) {
    return 0;
  }

  if (progress >= 1) {
    return 1;
  }

  return progress;
}

function calculatePlaybackProgress(currentTime: number, duration: number) {
  if (!Number.isFinite(currentTime) || !Number.isFinite(duration) || duration <= 0) {
    return 0;
  }

  return clampPlaybackProgress(currentTime / duration);
}

function buildUnlockStateKey(unlock: UnlockSurfaceModel): string {
  return [
    unlock.short.id,
    unlock.main.id,
    unlock.access.status,
    unlock.access.reason,
    unlock.unlockCta.state,
    unlock.mainAccessEntry.routePath,
    unlock.mainAccessEntry.token,
    unlock.setup.required ? "setup-required" : "setup-optional",
    unlock.setup.requiresAgeConfirmation ? "age-required" : "age-optional",
    unlock.setup.requiresTermsAcceptance ? "terms-required" : "terms-optional",
  ].join("::");
}

export type ImmersiveShortSurfaceProps =
  | {
      activeTab: FeedTab;
      isActive?: boolean;
      mode: "feed";
      pin?: {
        errorMessage: string | null;
        isPending: boolean;
        isPinned: boolean;
        onToggle: () => void;
      };
      surface: FeedShortSurface;
    }
  | {
      backHref: string;
      creatorProfileOrigin: CreatorProfileRouteOrigin;
      isActive?: boolean;
      mode: "detail";
      presentation?: "default" | "feedLike";
      surface: DetailShortSurface;
    };

/**
 * feed/detail 共通の header 領域を表示する。
 */
function ShortSurfaceHeader(props: ImmersiveShortSurfaceProps) {
  if (props.mode === "feed") {
    return (
      <div className="absolute top-0 z-20 flex w-full items-center justify-between px-4 pb-4 pt-14">
        <div className="w-6" />
        <nav aria-label="Feed sections" className="flex items-center space-x-6 text-lg font-bold drop-shadow-md">
          {[
            { active: props.activeTab === "recommended", href: "/?tab=recommended", key: "recommended", label: "For You" },
            { active: props.activeTab === "following", href: "/?tab=following", key: "following", label: "Following" },
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
        <div aria-hidden="true" className="size-11" />
      </div>
    );
  }

  if (props.presentation === "feedLike") {
    return <FeedLikeShortBackHeader backHref={props.backHref} />;
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
  disabled?: boolean;
  onToggle?: (() => void) | undefined;
  pinned: boolean;
  variant?: "default" | "feed";
};

/**
 * short の pin 状態を表す操作レールを表示する。
 */
function PinRail({ disabled = false, onToggle, pinned, variant = "default" }: PinRailProps) {
  const label = pinned ? "Pinned short" : "Pin short";
  const isFeedVariant = variant === "feed";

  return (
    <div className="flex flex-col items-center gap-2.5">
      <button
        aria-label={label}
        aria-busy={disabled || undefined}
        aria-pressed={pinned}
        className={cn(
          isFeedVariant
            ? "inline-flex size-11 items-center justify-center bg-transparent p-0 text-white drop-shadow-lg transition-transform hover:scale-110 disabled:cursor-wait disabled:hover:scale-100"
            : "inline-flex size-11 items-center justify-center rounded-full bg-transparent p-0 text-accent-strong/72 transition hover:text-accent disabled:cursor-wait disabled:hover:text-accent-strong/72",
          pinned && (isFeedVariant ? "text-[#8fd1ff]" : "text-accent"),
        )}
        disabled={disabled}
        onClick={onToggle}
        type="button"
      >
        <svg
          aria-hidden="true"
          className={cn("size-[22px]", isFeedVariant && "h-7 w-7")}
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
  short: FeedShortSurface["short"];
  variant?: "default" | "feed" | undefined;
};

/**
 * feed surface 用の creator avatar を表示し、avatar 不在時は initials fallback を描画する。
 */
function FeedCreatorAvatar({
  className,
  creator,
}: Pick<CreatorBlockProps, "creator"> & {
  className?: string;
}) {
  if (!creator.avatar) {
    return (
      <span
        aria-hidden="true"
        className={cn(
          "flex h-[38px] w-[38px] shrink-0 items-center justify-center rounded-full bg-[linear-gradient(180deg,#b2ecff_0%,#65bae0_56%,#1b4362_100%)] text-[11px] font-semibold uppercase tracking-[0.08em] text-white shadow-[0_8px_20px_rgba(7,19,29,0.2)]",
          className,
        )}
      >
        {getCreatorInitials(creator.displayName)}
      </span>
    );
  }

  return (
    <CreatorAvatar
      className={cn("size-[38px] rounded-full border-white/68 shadow-[0_8px_20px_rgba(7,19,29,0.2)]", className)}
      creator={creator}
    />
  );
}

function FeedActionRail({
  disabled = false,
  onToggle,
  pinned,
}: {
  disabled?: boolean;
  onToggle?: () => void;
  pinned: boolean;
}) {
  return (
    <div
      className="absolute right-3 z-20 flex flex-col items-center space-y-6"
      data-testid="feed-action-rail"
      style={{ bottom: feedActionRailBottom }}
    >
      <PinRail disabled={disabled} onToggle={onToggle} pinned={pinned} variant="feed" />
    </div>
  );
}

function FeedPlaybackProgressBar({
  onSeek,
  progress,
}: {
  onSeek: (nextProgress: number) => void;
  progress: number;
}) {
  const handleClick = (event: ReactMouseEvent<HTMLDivElement>) => {
    const bounds = event.currentTarget.getBoundingClientRect();

    if (bounds.width <= 0) {
      return;
    }

    onSeek(clampPlaybackProgress((event.clientX - bounds.left) / bounds.width));
  };

  const handleKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
    let nextProgress: number | null = null;

    switch (event.key) {
      case "ArrowLeft":
      case "ArrowDown":
        nextProgress = progress - 0.05;
        break;
      case "ArrowRight":
      case "ArrowUp":
        nextProgress = progress + 0.05;
        break;
      case "Home":
        nextProgress = 0;
        break;
      case "End":
        nextProgress = 1;
        break;
      default:
        return;
    }

    event.preventDefault();
    onSeek(clampPlaybackProgress(nextProgress));
  };

  return (
    <div
      aria-label="Short playback progress"
      aria-valuemax={100}
      aria-valuemin={0}
      aria-valuenow={Math.round(progress * 100)}
      className="absolute left-0 z-10 h-5 w-full cursor-pointer"
      data-testid="feed-playback-progress-bar"
      onClick={handleClick}
      onKeyDown={handleKeyDown}
      role="slider"
      style={{ bottom: sharedFanNavigationInset, touchAction: "pan-y" }}
      tabIndex={0}
    >
      <div aria-hidden="true" className="absolute bottom-0 left-0 h-[2px] w-full overflow-hidden bg-white/20">
        <div
          className="h-full w-full rounded-r-full bg-white"
          data-testid="feed-playback-progress-fill"
          style={{
            transform: `scaleX(${progress})`,
            transformOrigin: "left center",
          }}
        />
      </div>
    </div>
  );
}

/**
 * creator 名、follow 状態、caption をまとめた下部 creator block を表示する。
 */
function CreatorBlock({
  creator,
  followState,
  followed = false,
  profileHref,
  short,
  variant = "default",
}: CreatorBlockProps) {
  const caption = short.caption.trim();
  const interactiveFollowState = followState ?? null;
  const resolvedIsFollowing = followState?.isFollowing ?? followed;
  const isFeedVariant = variant === "feed";
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
    isFeedVariant
      ? "rounded-full border border-white/80 bg-transparent px-3 py-1 text-xs font-semibold text-white backdrop-blur-sm transition"
      : "min-h-7 shrink-0 rounded-full border border-white/62 bg-transparent px-3 text-[11px] font-semibold text-white/92 transition",
    resolvedIsFollowing && (isFeedVariant ? "border-white/24 bg-white/18 text-white" : "border-[#b6eaff]/78 text-[#d7f5ff]"),
  );

  if (isFeedVariant) {
    return (
      <div className="w-[calc(100%-88px)]">
        <div className="mb-2.5 flex w-max max-w-full items-center gap-2.5">
          <Link
            aria-label={creator.displayName}
            className="inline-flex min-w-0 items-center gap-2.5 text-left text-white transition hover:opacity-90"
            href={profileHref}
          >
            <FeedCreatorAvatar className="size-8 border-white/60 shadow-sm" creator={creator} />
            <span className="truncate text-[15px] font-bold text-white">{creator.handle}</span>
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
        {caption ? (
          <p className="pr-16 text-sm font-medium leading-snug text-white drop-shadow-md line-clamp-2">{caption}</p>
        ) : null}
        {followState?.errorMessage ? (
          <p
            className="mt-3 rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
            role="alert"
          >
            {followState.errorMessage}
          </p>
        ) : null}
      </div>
    );
  }

  return (
    <div className="absolute inset-x-0 bottom-0 z-10 px-4" style={{ paddingBottom: "68px" }}>
      <div className="w-[min(88%,344px)] max-w-[344px]">
        <div className="flex w-fit max-w-full items-center gap-2">
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
  variant = "default",
}: FeedCreatorBlockProps) {
  const { openFanAuthDialog } = useFanAuthDialogControls();
  const { errorMessage, isFollowing, isPending, toggleFollow } = useCreatorFollowToggle({
    creatorId: creator.id,
    hasViewerSession,
    initialIsFollowing,
    onAuthRequired: () => {
      openFanAuthDialog({
        postAuthNavigation: "none",
      });
    },
    onUnauthenticated: () => {
      openFanAuthDialog({
        postAuthNavigation: "none",
      });
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
      variant={variant}
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
  const [feedPlaybackProgress, setFeedPlaybackProgress] = useState(0);
  const [isPaywallOpen, setIsPaywallOpen] = useState(false);
  const [isSubmittingMainAccess, setIsSubmittingMainAccess] = useState(false);
  const [resolvedUnlockState, setResolvedUnlockState] = useState<{
    baseUnlockKey: string;
    shortId: string;
    viewerIdentityKey: string | null;
    unlock: UnlockSurfaceModel;
  } | null>(null);
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const currentViewer = useCurrentViewer();
  const hasViewerSession = useHasViewerSession();
  const router = useRouter();
  const { openFanAuthDialog } = useFanAuthDialogControls();
  const { mode, surface } = props;
  const { creator, short, unlock, viewer } = surface;
  const detailPresentation = mode === "detail" ? props.presentation ?? "default" : "default";
  const usesFeedPresentation = mode === "feed" || detailPresentation === "feedLike";
  const detailPinState = useShortPinState({
    enabled: mode === "detail",
    initialIsPinned: viewer.isPinned,
    shortId: short.id,
  });
  const isActive = props.isActive ?? true;
  const pinErrorMessage = mode === "feed" ? props.pin?.errorMessage ?? null : detailPinState.errorMessage;
  const pinned = mode === "feed" ? props.pin?.isPinned ?? viewer.isPinned : detailPinState.isPinned;
  const propUnlockKey = buildUnlockStateKey(unlock);
  const viewerIdentityKey = currentViewer?.id ?? null;
  const resolvedUnlock =
    resolvedUnlockState?.shortId === short.id &&
    resolvedUnlockState.baseUnlockKey === propUnlockKey &&
    resolvedUnlockState.viewerIdentityKey === viewerIdentityKey
      ? resolvedUnlockState.unlock
      : null;
  const activeUnlock = resolvedUnlock ?? unlock;
  const unlockAction = getUnlockEntryAction(activeUnlock);
  const surfaceStyle = usesFeedPresentation ? feedSurfaceStyle : getShortThemeStyle(short);
  const profileHref =
    mode === "feed"
      ? buildCreatorProfileHref(creator.id, {
          from: "feed",
          tab: props.activeTab,
        })
      : buildCreatorProfileHref(creator.id, props.creatorProfileOrigin);
  const usesApiBackedUnlockFlow = short.id.startsWith("short_");

  const closePaywall = () => {
    setIsPaywallOpen(false);
  };

  const resetPaywallSelections = () => {
    setAcceptAge(false);
    setAcceptTerms(false);
  };

  const resetPaywallState = () => {
    resetPaywallSelections();
    closePaywall();
  };

  const storeResolvedUnlock = (
    nextUnlock: UnlockSurfaceModel,
    resolvedViewerIdentityKey: string | null,
  ) => {
    setResolvedUnlockState({
      baseUnlockKey: propUnlockKey,
      shortId: short.id,
      viewerIdentityKey: resolvedViewerIdentityKey,
      unlock: nextUnlock,
    });
  };

  const resolveUnlockSurfaceAfterAuth = async ({
    resolvedViewerIdentityKey,
  }: {
    resolvedViewerIdentityKey: string | null;
  }) => {
    if (!usesApiBackedUnlockFlow) {
      return null;
    }

    const nextUnlock = await requestUnlockSurfaceByShortId({
      shortId: short.id,
    });

    storeResolvedUnlock(nextUnlock, resolvedViewerIdentityKey);

    return nextUnlock;
  };

  const shouldOpenPaywallForUnlock = (targetUnlock: UnlockSurfaceModel) => {
    const targetAction = getUnlockEntryAction(targetUnlock);

    return targetAction === "open_paywall" || (usesFeedPresentation && targetUnlock.unlockCta.state === "unlock_available");
  };

  const restoreUnlockSurfaceAfterAuth = async ({
    preservePaywallSelections,
    restoredViewer,
    targetUnlock,
  }: {
    preservePaywallSelections: boolean;
    restoredViewer?: {
      id: string;
    } | null | undefined;
    targetUnlock: UnlockSurfaceModel;
  }) => {
    let unlockAfterAuth = targetUnlock;
    const restoredViewerIdentityKey = restoredViewer?.id ?? null;

    if (usesApiBackedUnlockFlow) {
      const nextUnlock = await resolveUnlockSurfaceAfterAuth({
        resolvedViewerIdentityKey: restoredViewerIdentityKey,
      });
      if (nextUnlock) {
        unlockAfterAuth = nextUnlock;
      }
    }

    const shouldRestorePaywall = shouldOpenPaywallForUnlock(unlockAfterAuth);

    if (!preservePaywallSelections) {
      resetPaywallSelections();
    }

    closePaywall();

    if (shouldRestorePaywall) {
      setIsPaywallOpen(true);
    }
  };

  const openReAuthDialog = ({
    preservePaywallSelections = isPaywallOpen,
    targetUnlock = activeUnlock,
  }: {
    preservePaywallSelections?: boolean;
    targetUnlock?: UnlockSurfaceModel;
  } = {}) => {
    openFanAuthDialog({
      allowClose: false,
      initialMode: "re-auth",
      onAfterAuthenticated:
        usesApiBackedUnlockFlow || preservePaywallSelections
          ? async (restoredViewer) => {
              await restoreUnlockSurfaceAfterAuth({
                preservePaywallSelections,
                restoredViewer,
                targetUnlock,
              });
            }
          : undefined,
      postAuthNavigation: "none",
    });
  };

  const openAuthDialogForUnlock = (
    targetUnlock: UnlockSurfaceModel = activeUnlock,
    {
      preservePaywallSelections = isPaywallOpen,
    }: {
      preservePaywallSelections?: boolean;
    } = {},
  ) => {
    const shouldRefreshUnlockAfterAuth = usesApiBackedUnlockFlow;
    const shouldRestoreUnlockAfterAuth =
      shouldRefreshUnlockAfterAuth || preservePaywallSelections || shouldOpenPaywallForUnlock(targetUnlock);

    openFanAuthDialog({
      onAfterAuthenticated: shouldRestoreUnlockAfterAuth
        ? async (restoredViewer) => {
            await restoreUnlockSurfaceAfterAuth({
              preservePaywallSelections,
              restoredViewer,
              targetUnlock,
            });
          }
        : undefined,
      postAuthNavigation: "none",
    });
  };
  const seekFeedPlayback = (nextProgress: number) => {
    if (!usesFeedPresentation) {
      return;
    }

    const video = videoRef.current;

    if (!video) {
      return;
    }

    const fallbackDurationSeconds = short.media.durationSeconds ?? short.previewDurationSeconds;
    const resolvedDuration =
      Number.isFinite(video.duration) && video.duration > 0 ? video.duration : fallbackDurationSeconds;

    if (!Number.isFinite(resolvedDuration) || resolvedDuration <= 0) {
      return;
    }

    const clampedProgress = clampPlaybackProgress(nextProgress);

    video.currentTime = resolvedDuration * clampedProgress;
    setFeedPlaybackProgress(clampedProgress);
  };

  useEffect(() => {
    setIsHydrated(true);
  }, []);

  useEffect(() => {
    if (!usesFeedPresentation) {
      return;
    }

    const video = videoRef.current;

    if (!video) {
      return;
    }

    const syncProgress = () => {
      const nextProgress = calculatePlaybackProgress(video.currentTime, video.duration);

      setFeedPlaybackProgress((currentProgress) =>
        Math.abs(currentProgress - nextProgress) < 0.001 ? currentProgress : nextProgress,
      );
    };

    setFeedPlaybackProgress(0);

    video.addEventListener("durationchange", syncProgress);
    video.addEventListener("emptied", syncProgress);
    video.addEventListener("loadedmetadata", syncProgress);
    video.addEventListener("timeupdate", syncProgress);

    return () => {
      video.removeEventListener("durationchange", syncProgress);
      video.removeEventListener("emptied", syncProgress);
      video.removeEventListener("loadedmetadata", syncProgress);
      video.removeEventListener("timeupdate", syncProgress);
    };
  }, [usesFeedPresentation, short.media.id, short.media.url]);

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
  const handleOpenMain = async (targetUnlock: UnlockSurfaceModel) => {
    if (!hasViewerSession) {
      openAuthDialogForUnlock(targetUnlock);
      return;
    }

    if (isSubmittingMainAccess) {
      return;
    }

    setIsSubmittingMainAccess(true);

    try {
      if (usesApiBackedUnlockFlow) {
        const response = await requestMainAccessEntry({
          acceptedAge: acceptAge,
          acceptedTerms: acceptTerms,
          entryToken: targetUnlock.mainAccessEntry.token,
          fromShortId: short.id,
          mainId: targetUnlock.main.id,
          routePath: targetUnlock.mainAccessEntry.routePath as `/${string}`,
        });

        resetPaywallState();
        router.push(response.href);
        return;
      }

      const response = await fetch(targetUnlock.mainAccessEntry.routePath, {
        body: JSON.stringify({
          acceptedAge: acceptAge,
          acceptedTerms: acceptTerms,
          entryToken: targetUnlock.mainAccessEntry.token,
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
            } | null;
          }
        | null;

      if (!response.ok && isAuthRequiredResponse(payload)) {
        openAuthDialogForUnlock(targetUnlock, {
          preservePaywallSelections: isPaywallOpen,
        });
        return;
      }

      if (!response.ok && isFreshAuthRequiredResponse(payload)) {
        openReAuthDialog({
          preservePaywallSelections: isPaywallOpen,
          targetUnlock,
        });
        return;
      }

      if (response.ok && payload?.data?.href) {
        resetPaywallState();
        router.push(payload.data.href);
        return;
      }
    } catch (error) {
      if (isAuthRequiredApiError(error)) {
        openAuthDialogForUnlock(targetUnlock, {
          preservePaywallSelections: isPaywallOpen,
        });
        return;
      }

      if (isFreshAuthRequiredApiError(error)) {
        openReAuthDialog({
          preservePaywallSelections: isPaywallOpen,
          targetUnlock,
        });
        return;
      }

      router.push(`/shorts/${short.id}`);
    } finally {
      setIsSubmittingMainAccess(false);
    }
  };

  /**
   * CTA 押下時に必要な unlock surface を解決して既存 flow へ接続する。
   */
  const handleActivateUnlock = async () => {
    if (!hasViewerSession) {
      openAuthDialogForUnlock(activeUnlock);
      return;
    }

    if (isSubmittingMainAccess) {
      return;
    }

    const targetUnlock = activeUnlock;
    const nextAction = getUnlockEntryAction(targetUnlock);
    const shouldOpenPaywall = shouldOpenPaywallForUnlock(targetUnlock);

    if (shouldOpenPaywall) {
      resetPaywallSelections();
      setIsPaywallOpen(true);
      return;
    }

    if (nextAction === "open_main") {
      await handleOpenMain(targetUnlock);
    }
  };

  const pinProps =
    mode === "feed" && props.pin
      ? {
          disabled: props.pin.isPending,
          onToggle: props.pin.onToggle,
        }
      : {
          disabled: detailPinState.isPending,
          onToggle: detailPinState.onToggle,
        };

  if (usesFeedPresentation) {
    return (
      <FeedLikeShortBackdrop
        header={<ShortSurfaceHeader {...props} />}
        mediaLayer={
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
        }
      >
        <h1 className="sr-only">{mode === "feed" ? "Feed" : "Short detail"}</h1>
        <FeedActionRail pinned={pinned} {...pinProps} />
        {pinErrorMessage ? (
          <p
            aria-live="polite"
            className="absolute right-4 z-20 max-w-[220px] rounded-[20px] border border-white/16 bg-[rgba(7,19,29,0.72)] px-3 py-2 text-[11px] leading-[1.45] text-white/92 shadow-[0_16px_28px_rgba(7,19,29,0.28)] backdrop-blur-[10px]"
            role="alert"
            style={{ bottom: feedPinErrorBottom }}
          >
            {pinErrorMessage}
          </p>
        ) : null}
        <div
          className="absolute left-0 z-10 w-full bg-gradient-to-t from-black/90 via-black/40 to-transparent px-4 pb-5 pt-16"
          style={{ bottom: sharedFanNavigationInset }}
        >
          <UnlockCta
            className="mb-4 w-full"
            cta={activeUnlock.unlockCta}
            disabled={!isHydrated || isSubmittingMainAccess}
            variant="feed"
            {...(surface.mainEntryEnabled && unlockAction !== "unavailable"
              ? {
                  onClick: () => {
                    void handleActivateUnlock();
                  },
                }
              : {})}
          />
          <FeedCreatorBlock
            creator={creator}
            hasViewerSession={hasViewerSession}
            initialIsFollowing={viewer.isFollowingCreator}
            profileHref={profileHref}
            short={short}
            variant="feed"
          />
        </div>
        <FeedPlaybackProgressBar onSeek={seekFeedPlayback} progress={feedPlaybackProgress} />
        <UnlockPaywallDialog
          acceptAge={acceptAge}
          acceptTerms={acceptTerms}
          isSubmitting={isSubmittingMainAccess}
          onAcceptAgeChange={setAcceptAge}
          onAcceptTermsChange={setAcceptTerms}
          onClose={handleClosePaywall}
          onConfirm={() => {
            void handleOpenMain(activeUnlock);
          }}
          open={isPaywallOpen}
          unlock={activeUnlock}
        />
      </FeedLikeShortBackdrop>
    );
  }

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
      <div
        className={cn(
          "absolute inset-0",
          "bg-[linear-gradient(180deg,rgba(6,21,33,0.08)_0%,rgba(6,21,33,0.18)_20%,rgba(6,21,33,0.36)_58%,rgba(6,21,33,0.74)_100%)]",
        )}
      />
      <div
        className={cn(
          "absolute inset-0",
          "bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.18),transparent_34%)]",
        )}
      />

      <div className="relative h-full">
        <h1 className="sr-only">Short detail</h1>
        <ShortSurfaceHeader {...props} />
        <div className="absolute right-4 z-20" style={{ bottom: "204px" }}>
          <PinRail pinned={pinned} {...pinProps} />
        </div>
        {pinErrorMessage ? (
          <p
            aria-live="polite"
            className="absolute right-4 z-20 max-w-[220px] rounded-[20px] border border-white/16 bg-[rgba(7,19,29,0.72)] px-3 py-2 text-[11px] leading-[1.45] text-white/92 shadow-[0_16px_28px_rgba(7,19,29,0.28)] backdrop-blur-[10px]"
            role="alert"
            style={{ bottom: "258px" }}
          >
            {pinErrorMessage}
          </p>
        ) : null}
        <div className="absolute inset-x-4 z-20" style={{ bottom: "152px" }}>
          <UnlockCta
            className="w-full"
            cta={activeUnlock.unlockCta}
            disabled={!isHydrated || isSubmittingMainAccess}
            {...(surface.mainEntryEnabled && unlockAction !== "unavailable"
              ? {
                  onClick: () => {
                    void handleActivateUnlock();
                  },
                }
              : {})}
          />
        </div>
        <FeedCreatorBlock
          creator={creator}
          hasViewerSession={hasViewerSession}
          initialIsFollowing={viewer.isFollowingCreator}
          profileHref={profileHref}
          short={short}
        />
        <UnlockPaywallDialog
          acceptAge={acceptAge}
          acceptTerms={acceptTerms}
          isSubmitting={isSubmittingMainAccess}
          onAcceptAgeChange={setAcceptAge}
          onAcceptTermsChange={setAcceptTerms}
          onClose={handleClosePaywall}
          onConfirm={() => {
            void handleOpenMain(activeUnlock);
          }}
          open={isPaywallOpen}
          unlock={activeUnlock}
        />
      </div>
    </section>
  );
}

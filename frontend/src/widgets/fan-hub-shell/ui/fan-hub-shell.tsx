"use client";

import Link from "next/link";
import {
  ArrowLeft,
  Bookmark,
  ChevronRight,
  Menu,
  SquarePlay,
} from "lucide-react";

import type { FanHubState } from "@/entities/fan-profile";
import { useCurrentViewer } from "@/entities/viewer";
import { useCreatorModeEntry } from "@/features/creator-entry";
import { useFanLogoutEntry } from "@/features/fan-auth";
import {
  buildFanProfileLibraryMainHref,
  buildFanProfileShortDetailHref,
} from "@/features/creator-navigation";
import { getShortThemeStyle } from "@/entities/short";
import {
  BottomSheetMenu,
  BottomSheetMenuAction,
  BottomSheetMenuClose,
  BottomSheetMenuGroup,
  Button,
} from "@/shared/ui";

type FanHubShellProps = {
  state: FanHubState;
};

function buildLibraryTileLabel(item: FanHubState["libraryItems"][number]): string {
  const caption = item.entryShort.caption.trim();

  if (caption) {
    return `${item.creator.displayName} ${caption}`;
  }

  const fallbackLabel = item.access.status === "owner" ? "owner preview main" : "unlocked main";

  return `${item.creator.displayName} ${fallbackLabel}`;
}

function FanProfileAvatar() {
  return (
    <span
      aria-hidden="true"
      className="block size-[82px] shrink-0 rounded-full bg-[linear-gradient(180deg,#dff5ff_0%,#86d0f0_44%,#22557a_100%)] shadow-[0_10px_24px_rgba(40,95,135,0.14)]"
    />
  );
}

function FanStat({
  count,
  href,
  label,
}: {
  count: number;
  href?: string;
  label: string;
}) {
  const content = (
    <>
      <strong className="block font-display text-[19px] font-semibold tracking-[-0.04em] text-foreground">
        {count}
      </strong>
      <span className="mt-1 block text-[11px] text-muted">
        {label}
        {href ? (
          <b aria-hidden="true" className="ml-1 text-[10px] font-bold">
            {">"}
          </b>
        ) : null}
      </span>
    </>
  );

  if (!href) {
    return <div className="min-w-0 text-center">{content}</div>;
  }

  return (
    <Link aria-label={label} className="min-w-0 text-center transition hover:opacity-90" href={href}>
      {content}
    </Link>
  );
}

function FanTabLink({
  active,
  href,
  icon,
  label,
}: {
  active: boolean;
  href: string;
  icon: "library" | "pinned";
  label: string;
}) {
  const Icon = icon === "library" ? SquarePlay : Bookmark;

  return (
    <Link
      aria-current={active ? "page" : undefined}
      aria-label={label}
      className={[
        "inline-flex min-h-[42px] items-center justify-center border-t-2 bg-transparent px-0 pb-2 pt-2.5 text-accent-strong/56 transition hover:text-accent-strong/80",
        active ? "border-foreground text-foreground" : "border-transparent",
      ].join(" ")}
      href={href}
    >
      <Icon aria-hidden="true" className="size-[18px]" strokeWidth={1.9} />
      <span className="sr-only">{label}</span>
    </Link>
  );
}

function FanMediaTile({
  href,
  label,
  short,
}: {
  href?: string;
  label: string;
  short: {
    id: string;
    media: {
      posterUrl: string | null;
      url: string;
    };
  };
}) {
  const frame = (
    <span className="relative block aspect-[3/4] overflow-hidden rounded-[4px] bg-[linear-gradient(180deg,var(--short-tile-top)_0%,var(--short-tile-mid)_42%,var(--short-tile-bottom)_100%)] shadow-[0_14px_28px_rgba(36,94,132,0.12)] transition-transform hover:scale-[1.01]">
      {short.media.posterUrl ? (
        <span
          aria-hidden="true"
          className="absolute inset-0 block bg-cover bg-center"
          style={{ backgroundImage: `url("${short.media.posterUrl}")` }}
        />
      ) : (
        <video
          aria-hidden="true"
          className="absolute inset-0 size-full object-cover"
          muted
          playsInline
          preload="none"
          src={short.media.url}
        />
      )}
      <span className="absolute inset-0 bg-[linear-gradient(180deg,rgba(6,21,33,0.04)_0%,rgba(6,21,33,0.2)_58%,rgba(6,21,33,0.4)_100%)]" />
    </span>
  );

  if (!href) {
    return (
      <button
        aria-label={label}
        className="block cursor-default text-left"
        style={getShortThemeStyle(short)}
        type="button"
      >
        {frame}
      </button>
    );
  }

  return (
    <Link aria-label={label} className="block text-left" href={href} style={getShortThemeStyle(short)}>
      {frame}
    </Link>
  );
}

/**
 * private consumer hub の UI を表示する。
 */
export function FanHubShell({ state }: FanHubShellProps) {
  const { activeTab, libraryItems, overview, pinnedItems } = state;
  const hasItems = activeTab === "library" ? libraryItems.length > 0 : pinnedItems.length > 0;
  const currentViewer = useCurrentViewer();
  const {
    clearError: clearCreatorModeError,
    enterCreatorMode,
    errorMessage: creatorModeErrorMessage,
    isSubmitting: isCreatorModeSubmitting,
  } = useCreatorModeEntry();
  const {
    clearError: clearLogoutError,
    errorMessage: logoutErrorMessage,
    isSubmitting: isLogoutSubmitting,
    logout,
  } = useFanLogoutEntry();
  const isAccountActionPending = isCreatorModeSubmitting || isLogoutSubmitting;
  const accountMenuErrorMessage = creatorModeErrorMessage ?? logoutErrorMessage;

  const clearAccountMenuErrors = () => {
    clearCreatorModeError();
    clearLogoutError();
  };

  return (
    <section className="min-h-full overflow-y-auto px-4 pb-28 pt-4 text-foreground">
      <div className="flex items-center justify-between gap-3">
        <Button asChild size="icon" variant="ghost">
          <Link aria-label="Back" href="/">
            <ArrowLeft className="size-5" strokeWidth={2.1} />
          </Link>
        </Button>
        <BottomSheetMenu
          description="fan profile から creator registration、creator mode、logout を操作するメニュー"
          title="アカウントメニュー"
          trigger={
            <button
              aria-label="Account menu"
              className="inline-flex size-[34px] items-center justify-center rounded-full text-accent-strong transition hover:bg-accent/10"
              onClick={clearAccountMenuErrors}
              type="button"
            >
              <Menu aria-hidden="true" className="size-5" strokeWidth={1.9} />
            </button>
          }
        >
          <BottomSheetMenuGroup>
            <BottomSheetMenuClose asChild>
              <BottomSheetMenuAction asChild>
                <Link
                  aria-disabled={isAccountActionPending}
                  aria-label="プロフィールを編集"
                  className={isAccountActionPending ? "pointer-events-none opacity-55" : ""}
                  href="/fan/settings/profile"
                  onClick={clearAccountMenuErrors}
                  tabIndex={isAccountActionPending ? -1 : undefined}
                >
                  <span>プロフィールを編集</span>
                  <ChevronRight aria-hidden="true" className="size-4 text-muted" strokeWidth={2.2} />
                </Link>
              </BottomSheetMenuAction>
            </BottomSheetMenuClose>
            {currentViewer?.canAccessCreatorMode ? (
              <BottomSheetMenuAction
                disabled={isAccountActionPending}
                onClick={() => {
                  clearLogoutError();
                  void enterCreatorMode();
                }}
              >
                <span>{isCreatorModeSubmitting ? "Creator mode を開いています..." : "Creator mode に切り替え"}</span>
                <ChevronRight aria-hidden="true" className="size-4 text-muted" strokeWidth={2.2} />
              </BottomSheetMenuAction>
            ) : (
              <BottomSheetMenuClose asChild>
                <BottomSheetMenuAction asChild>
                  <Link
                    aria-disabled={isAccountActionPending}
                    aria-label="Creator登録を始める"
                    className={isAccountActionPending ? "pointer-events-none opacity-55" : ""}
                    href="/fan/creator/register"
                    onClick={clearLogoutError}
                    tabIndex={isAccountActionPending ? -1 : undefined}
                  >
                    <span>Creator登録を始める</span>
                    <ChevronRight aria-hidden="true" className="size-4 text-muted" strokeWidth={2.2} />
                  </Link>
                </BottomSheetMenuAction>
              </BottomSheetMenuClose>
            )}

            <BottomSheetMenuAction
              disabled={isAccountActionPending}
              onClick={() => {
                clearCreatorModeError();
                void logout();
              }}
              tone="danger"
              withDivider
            >
              <span>{isLogoutSubmitting ? "ログアウトしています..." : "ログアウト"}</span>
              <ChevronRight aria-hidden="true" className="size-4 text-[#d76a7f]" strokeWidth={2.2} />
            </BottomSheetMenuAction>
          </BottomSheetMenuGroup>

          {accountMenuErrorMessage ? (
            <p
              aria-live="polite"
              className="mt-3 rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
              role="alert"
            >
              {accountMenuErrorMessage}
            </p>
          ) : null}
        </BottomSheetMenu>
      </div>

      <section className="mt-3">
        <div className="flex items-center gap-4">
          <FanProfileAvatar />
          <div className="min-w-0 flex-1">
            <div className="grid grid-cols-3 gap-2 text-center">
              <FanStat count={overview.counts.following} href="/fan/following" label="Following" />
              <FanStat count={overview.counts.pinnedShorts} label="Pinned" />
              <FanStat count={overview.counts.library} label="Library" />
            </div>
          </div>
        </div>

        <div className="mt-3.5">
          <h1 className="text-[17px] font-bold text-foreground">{overview.title}</h1>
        </div>

        <nav aria-label="Profile sections" className="mt-[18px] grid grid-cols-2 border-t border-border/70">
          <FanTabLink active={activeTab === "pinned"} href="/fan?tab=pinned" icon="pinned" label="Pinned" />
          <FanTabLink active={activeTab === "library"} href="/fan?tab=library" icon="library" label="Library" />
        </nav>

        <div className="pt-2">
          {hasItems ? (
            <div className="grid grid-cols-3 gap-[3px]">
              {activeTab === "library"
                ? libraryItems.map((item) => (
                    <FanMediaTile
                      href={buildFanProfileLibraryMainHref(item.main.id, item.entryShort.id)}
                      key={item.main.id}
                      label={buildLibraryTileLabel(item)}
                      short={item.entryShort}
                    />
                  ))
                : pinnedItems.map((item) => (
                    <FanMediaTile
                      key={item.short.id}
                      href={buildFanProfileShortDetailHref(item.short.id, "pinned")}
                      label={`${item.creator.displayName} ${item.short.caption}`}
                      short={item.short}
                    />
                  ))}
            </div>
          ) : (
            <p className="mt-4 text-[13px] leading-6 text-muted">
              {activeTab === "library" ? "unlock した main はまだありません。" : "pin した short はまだありません。"}
            </p>
          )}
        </div>
      </section>
    </section>
  );
}

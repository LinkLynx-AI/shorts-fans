"use client";

import Link from "next/link";
import {
  ChevronRight,
  Pin,
  Play,
  Settings,
} from "lucide-react";

import type { FanHubState } from "@/entities/fan-profile";
import { getShortThemeStyle } from "@/entities/short";
import { useCurrentViewer } from "@/entities/viewer";
import { useCreatorModeEntry } from "@/features/creator-entry";
import {
  useFanAuthDialog,
  useFanLogoutEntry,
} from "@/features/fan-auth";
import {
  buildFanProfileLibraryMainHref,
  buildFanProfileShortDetailHref,
} from "@/features/creator-navigation";
import { FollowingCreatorList } from "@/features/following-creator-list";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  BottomSheetMenu,
  BottomSheetMenuAction,
  BottomSheetMenuClose,
  BottomSheetMenuGroup,
  SurfacePanel,
} from "@/shared/ui";

const FAN_HUB_BRAND_COLOR = "#4DA8DA";
const FAN_HUB_FONT_FAMILY =
  'ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", "Helvetica Neue", Arial, sans-serif';

type FanHubShellProps = {
  headerProfile: {
    avatarUrl: string | null;
    displayName: string;
    handle: string;
  };
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

function buildProfileInitials(displayName: string): string {
  const words = displayName
    .trim()
    .split(/[\s_]+/)
    .filter(Boolean);

  if (words.length === 0) {
    return "ME";
  }

  return words
    .slice(0, 2)
    .map((word) => word[0]?.toUpperCase() ?? "")
    .join("");
}

function FanProfileAvatar({
  avatarUrl,
  displayName,
}: {
  avatarUrl: string | null;
  displayName: string;
}) {
  return (
    <Avatar className="size-16 border border-[#eef1f5] bg-[linear-gradient(180deg,#f0f3f8_0%,#dce3eb_100%)] text-[18px] font-bold text-[#2b2d33] shadow-none">
      {avatarUrl ? <AvatarImage alt={`${displayName} avatar`} src={avatarUrl} /> : null}
      <AvatarFallback className="bg-transparent text-inherit">{buildProfileInitials(displayName)}</AvatarFallback>
    </Avatar>
  );
}

function FanTabLink({
  active,
  href,
  label,
}: {
  active: boolean;
  href: string;
  label: string;
}) {
  return (
    <Link
      aria-current={active ? "page" : undefined}
      aria-label={label}
      className={[
        "inline-flex min-h-[52px] items-center justify-center border-b-[3px] px-2 text-[14px] font-bold transition-colors",
        active
          ? "border-[#4DA8DA] text-[#4DA8DA]"
          : "border-transparent text-[#9ca3af] hover:text-[#4b5563]",
      ].join(" ")}
      href={href}
    >
      <span>{label}</span>
    </Link>
  );
}

function FanMediaTile({
  badgeKind,
  href,
  label,
  short,
}: {
  badgeKind: "pin" | "play";
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
  const BadgeIcon = badgeKind === "pin" ? Pin : Play;
  const frame = (
    <span className="relative block aspect-[9/16] overflow-hidden bg-[linear-gradient(180deg,var(--short-tile-top)_0%,var(--short-tile-mid)_42%,var(--short-tile-bottom)_100%)] transition-transform hover:scale-[1.01]">
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
      <span className="absolute inset-0 bg-[linear-gradient(180deg,rgba(6,21,33,0.03)_0%,rgba(6,21,33,0.12)_52%,rgba(6,21,33,0.22)_100%)]" />
      <span className="absolute right-1.5 top-1.5 inline-flex size-7 items-center justify-center rounded-full bg-[rgba(79,82,91,0.72)] text-white shadow-[0_4px_10px_rgba(15,23,42,0.14)] backdrop-blur-md">
        <BadgeIcon
          aria-hidden="true"
          className="size-3.5"
          fill={badgeKind === "pin" ? "currentColor" : undefined}
          strokeWidth={2.2}
        />
      </span>
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
export function FanHubShell({ headerProfile, state }: FanHubShellProps) {
  const { activeTab, followingItems, libraryItems, overview, pinnedItems } = state;
  const hasItems =
    activeTab === "following"
      ? followingItems.length > 0
      : activeTab === "library"
        ? libraryItems.length > 0
        : pinnedItems.length > 0;
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
  const { openFanAuthDialog } = useFanAuthDialog();
  const isAccountActionPending = isCreatorModeSubmitting || isLogoutSubmitting;
  const accountMenuErrorMessage = creatorModeErrorMessage ?? logoutErrorMessage;

  const clearAccountMenuErrors = () => {
    clearCreatorModeError();
    clearLogoutError();
  };

  return (
    <section
      className="min-h-full overflow-y-auto bg-white pb-28 text-[#1f2430]"
      style={{ fontFamily: FAN_HUB_FONT_FAMILY }}
    >
      <div className="sticky top-0 z-20 flex items-center justify-between bg-white px-4 pb-4 pt-5 shadow-sm">
        <span aria-hidden="true" className="block size-6" />
        <h1 className="text-[17px] font-bold text-[#1f2430]">Profile</h1>
        <BottomSheetMenu
          description="fan profile から creator registration、creator mode、logout を操作するメニュー"
          title="アカウントメニュー"
          trigger={
            <button
              aria-label="Account menu"
              className="inline-flex size-6 items-center justify-center rounded-full text-[#4b5563] transition hover:text-[#111827]"
              onClick={clearAccountMenuErrors}
              type="button"
            >
              <Settings aria-hidden="true" className="size-6" strokeWidth={2.1} />
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

      <section>
        <div className="mb-2 bg-white px-6 py-5 shadow-sm">
          <div className="flex items-center gap-4">
            <FanProfileAvatar avatarUrl={headerProfile.avatarUrl} displayName={headerProfile.displayName} />
            <div className="min-w-0 flex-1">
              <p className="truncate text-[19px] font-extrabold text-[#20232d]">
                {headerProfile.displayName}
              </p>
              <p className="mt-1 truncate text-[13px] font-medium text-[#9097a5]">{headerProfile.handle}</p>
            </div>
          </div>
          <span className="sr-only">
            {`Following ${overview.counts.following}, Pinned Shorts ${overview.counts.pinnedShorts}, Library ${overview.counts.library}, ${overview.title}`}
          </span>
        </div>

        <nav
          aria-label="Profile sections"
          className="sticky top-[72px] z-10 grid grid-cols-3 border-b border-[#e5e7eb] bg-white px-4 text-[14px] font-bold"
        >
          <FanTabLink active={activeTab === "following"} href="/fan?tab=following" label="Following" />
          <FanTabLink active={activeTab === "pinned"} href="/fan?tab=pinned" label="Pinned Shorts" />
          <FanTabLink active={activeTab === "library"} href="/fan?tab=library" label="Library" />
        </nav>

        <div className="bg-white pb-6">
          {activeTab === "following" ? (
            <FollowingCreatorList items={followingItems} layout="embedded" onAuthRequired={openFanAuthDialog} />
          ) : hasItems ? (
            <div className="grid grid-cols-3 gap-0.5 p-0.5">
              {activeTab === "library"
                ? libraryItems.map((item) => (
                    <FanMediaTile
                      badgeKind="play"
                      href={buildFanProfileLibraryMainHref(item.main.id, item.entryShort.id)}
                      key={item.main.id}
                      label={buildLibraryTileLabel(item)}
                      short={item.entryShort}
                    />
                  ))
                : pinnedItems.map((item) => (
                    <FanMediaTile
                      badgeKind="pin"
                      key={item.short.id}
                      href={buildFanProfileShortDetailHref(item.short.id, "pinned")}
                      label={`${item.creator.displayName} ${item.short.caption}`}
                      short={item.short}
                    />
                  ))}
            </div>
          ) : (
            <SurfacePanel className="mx-4 mt-5 rounded-[28px] border-[#edf1f5] bg-white px-5 py-6 shadow-[0_10px_28px_rgba(15,23,42,0.05)]">
              <p className="text-[15px] font-bold text-[#222632]">
                {activeTab === "library" ? "Library はまだ空です" : "Pinned Shorts はまだありません"}
              </p>
              <p className="mt-2 text-[13px] leading-6 text-[#6f7787]">
                {activeTab === "library"
                  ? "unlock した main はここにまとまります。気になる short から続きを開いて library を育ててください。"
                  : "pin した short はここに保存されます。あとで見返したい short を feed から追加できます。"}
              </p>
              <Link
                aria-label={activeTab === "library" ? "feed で続きを探す" : "feed で short を探す"}
                className="mt-4 inline-flex min-h-10 items-center justify-center rounded-full px-4 text-sm font-semibold text-white transition"
                href="/"
                style={{ backgroundColor: FAN_HUB_BRAND_COLOR }}
              >
                {activeTab === "library" ? "feed で続きを探す" : "feed で short を探す"}
              </Link>
            </SurfacePanel>
          )}
        </div>
      </section>
    </section>
  );
}

"use client";

import { Button } from "@/shared/ui";

import {
  deriveCreatorWorkspaceReviewNotifications,
  type CreatorWorkspaceReviewNotification,
  type CreatorWorkspaceReviewSurfaceState,
} from "../model/creator-workspace-review-surface";

function getCreatorWorkspaceReviewNoticeClassName(tone: CreatorWorkspaceReviewNotification["tone"]): string {
  switch (tone) {
    case "revision":
      return "border-[rgba(244,152,45,0.18)] bg-[linear-gradient(180deg,rgba(255,248,238,0.96),rgba(252,242,224,0.92))]";
    case "removed":
      return "border-[rgba(255,184,189,0.84)] bg-[linear-gradient(180deg,rgba(255,247,248,0.98),rgba(255,241,243,0.96))]";
    case "pending":
      return "border-[rgba(126,190,228,0.42)] bg-[linear-gradient(180deg,rgba(245,251,255,0.98),rgba(236,247,252,0.94))]";
    case "paused":
      return "border-[rgba(167,220,249,0.4)] bg-[linear-gradient(180deg,rgba(251,253,255,0.98),rgba(244,250,253,0.94))]";
    default:
      return "border-[rgba(167,220,249,0.4)] bg-[linear-gradient(180deg,rgba(251,253,255,0.98),rgba(244,250,253,0.94))]";
  }
}

function getCreatorWorkspaceReviewNoticeBadgeClassName(tone: CreatorWorkspaceReviewNotification["tone"]): string {
  switch (tone) {
    case "revision":
      return "bg-[rgba(244,152,45,0.14)] text-[#8e4e0a]";
    case "removed":
      return "bg-[rgba(217,77,77,0.12)] text-[#9f2437]";
    case "pending":
      return "bg-[rgba(16,130,200,0.12)] text-[#0a5b8c]";
    case "paused":
      return "bg-[rgba(16,130,200,0.12)] text-[#0a5b8c]";
    default:
      return "bg-[rgba(52,168,83,0.12)] text-[#1d6f3a]";
  }
}

function CreatorWorkspaceReviewNoticeCard({
  notification,
}: {
  notification: CreatorWorkspaceReviewNotification;
}) {
  return (
    <div className={`rounded-[18px] border px-[14px] py-3 text-foreground ${getCreatorWorkspaceReviewNoticeClassName(notification.tone)}`}>
      <div className="flex items-start gap-3">
        <span className={`inline-flex min-h-7 items-center justify-center rounded-full px-3 text-[10px] font-bold uppercase tracking-[0.14em] ${getCreatorWorkspaceReviewNoticeBadgeClassName(notification.tone)}`}>
          {notification.label}
        </span>
        <div className="grid gap-1">
          <b className="text-[13px] leading-[1.35] text-foreground">{notification.headline}</b>
          <span className="text-[11px] leading-[1.55] text-muted">{notification.detail}</span>
        </div>
      </div>
    </div>
  );
}

function CreatorWorkspaceReviewNoticesLoading() {
  return (
    <section className="mt-[14px] grid gap-2">
      {Array.from({ length: 2 }).map((_, index) => (
        <div
          aria-hidden="true"
          className="h-[78px] animate-pulse rounded-[18px] border border-[rgba(167,220,249,0.32)] bg-[rgba(167,220,249,0.18)]"
          key={index}
        />
      ))}
    </section>
  );
}

function CreatorWorkspaceReviewNoticesError({
  message,
  onRetry,
}: {
  message: string;
  onRetry: () => void;
}) {
  return (
    <section className="mt-[14px] rounded-[18px] border border-[rgba(255,184,189,0.84)] bg-[linear-gradient(180deg,rgba(255,247,248,0.98),rgba(255,241,243,0.96))] px-4 py-4 text-foreground">
      <p className="m-0 text-[13px] leading-6 text-muted" role="alert">
        {message}
      </p>
      <div className="mt-3">
        <Button onClick={onRetry} size="sm" type="button" variant="secondary">
          再読み込み
        </Button>
      </div>
    </section>
  );
}

export function CreatorWorkspaceReviewNotices({
  onRetry,
  state,
}: {
  onRetry: () => void;
  state: CreatorWorkspaceReviewSurfaceState;
}) {
  if (state.kind === "loading") {
    return <CreatorWorkspaceReviewNoticesLoading />;
  }

  if (state.kind === "error") {
    return <CreatorWorkspaceReviewNoticesError message={state.message} onRetry={onRetry} />;
  }

  const notifications = deriveCreatorWorkspaceReviewNotifications(state.surface.packages);

  if (notifications.length === 0) {
    return null;
  }

  return (
    <section className="mt-[14px] grid gap-2">
      {notifications.map((notification) => (
        <CreatorWorkspaceReviewNoticeCard key={notification.key} notification={notification} />
      ))}
    </section>
  );
}

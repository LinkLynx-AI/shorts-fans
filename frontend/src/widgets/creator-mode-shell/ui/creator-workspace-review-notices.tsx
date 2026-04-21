"use client";

import {
  CircleAlert,
  CircleX,
} from "lucide-react";

import { Button } from "@/shared/ui";

import {
  deriveCreatorWorkspaceReviewNotifications,
  type CreatorWorkspaceReviewNotification,
  type CreatorWorkspaceReviewSurfaceState,
} from "../model/creator-workspace-review-surface";

function getCreatorWorkspaceReviewNoticeIconClassName(tone: CreatorWorkspaceReviewNotification["tone"]): string {
  switch (tone) {
    case "revision":
      return "bg-[rgba(244,152,45,0.09)] text-[#a65f11]";
    case "removed":
      return "bg-[rgba(159,36,55,0.08)] text-[#9f2437]";
    case "pending":
      return "bg-[rgba(16,130,200,0.08)] text-[#0a5b8c]";
    case "paused":
      return "bg-[rgba(16,130,200,0.08)] text-[#0a5b8c]";
    default:
      return "bg-[rgba(52,168,83,0.08)] text-[#1d6f3a]";
  }
}

function CreatorWorkspaceReviewNoticeIcon({
  tone,
}: {
  tone: CreatorWorkspaceReviewNotification["tone"];
}) {
  if (tone === "removed") {
    return <CircleX className="size-4" strokeWidth={2.1} />;
  }

  return <CircleAlert className="size-4" strokeWidth={2.1} />;
}

function CreatorWorkspaceReviewNoticeCard({
  notification,
}: {
  notification: CreatorWorkspaceReviewNotification;
}) {
  return (
    <div className="py-3 text-foreground">
      <div className="flex items-center gap-3">
        <span
          aria-hidden="true"
          className={`inline-flex size-8 shrink-0 items-center justify-center rounded-full ${getCreatorWorkspaceReviewNoticeIconClassName(notification.tone)}`}
        >
          <CreatorWorkspaceReviewNoticeIcon tone={notification.tone} />
        </span>
        <div className="min-w-0 flex-1">
          <p className="m-0 text-[13px] font-bold leading-[1.3] text-foreground">{notification.headline}</p>
          <p className="m-0 mt-0.5 text-[11px] leading-[1.45] text-muted">{notification.detail}</p>
        </div>
      </div>
    </div>
  );
}

function CreatorWorkspaceReviewNoticesLoading() {
  return (
    <section className="mt-[18px] divide-y divide-[rgba(7,19,29,0.08)] border-y border-[rgba(7,19,29,0.08)]">
      {Array.from({ length: 2 }).map((_, index) => (
        <div
          aria-hidden="true"
          className="h-[57px] animate-pulse bg-[linear-gradient(90deg,rgba(167,220,249,0.10),rgba(167,220,249,0.18),rgba(167,220,249,0.10))]"
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
    <section className="mt-[18px] divide-y divide-[rgba(7,19,29,0.08)] border-y border-[rgba(7,19,29,0.08)]">
      {notifications.map((notification) => (
        <CreatorWorkspaceReviewNoticeCard key={notification.key} notification={notification} />
      ))}
    </section>
  );
}

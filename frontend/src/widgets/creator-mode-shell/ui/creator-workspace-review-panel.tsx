"use client";

import { startTransition, useEffect, useState } from "react";

import {
  createCreatorWorkspaceSubmissionReview,
  CreatorWorkspaceSubmissionReviewApiError,
} from "@/features/creator-workspace-submission-review";
import { ApiError } from "@/shared/api";
import { Button } from "@/shared/ui";

import {
  buildCreatorWorkspaceReviewActionLabel,
  buildCreatorWorkspaceReviewPackageHeadline,
  resolveCreatorWorkspaceObjectReviewBadge,
  resolveCreatorWorkspacePackageReviewBadge,
  resolveCreatorWorkspaceReviewBlockerLabel,
  type CreatorWorkspaceItemReviewSurfaceState,
  type CreatorWorkspaceReviewBadge,
} from "../model/creator-workspace-review-surface";

function getCreatorWorkspaceReviewBadgeClassName(tone: CreatorWorkspaceReviewBadge["tone"]): string {
  switch (tone) {
    case "revision":
      return "bg-[rgba(244,152,45,0.14)] text-[#8e4e0a]";
    case "removed":
      return "bg-[rgba(217,77,77,0.12)] text-[#9f2437]";
    case "pending":
      return "bg-[rgba(16,130,200,0.12)] text-[#0a5b8c]";
    case "paused":
      return "bg-[rgba(16,130,200,0.12)] text-[#0a5b8c]";
    case "approved":
      return "bg-[rgba(52,168,83,0.12)] text-[#1d6f3a]";
    case "hidden":
      return "bg-[rgba(7,19,29,0.12)] text-[#1b3f5a]";
  }
}

function buildCreatorWorkspaceSubmissionReviewErrorMessage(error: unknown): string {
  if (error instanceof CreatorWorkspaceSubmissionReviewApiError) {
    switch (error.code) {
      case "auth_required":
      case "creator_mode_unavailable":
        return "creator mode が利用できないため、申請を続けられません。";
      case "not_found":
        return "対象 package が見つからないため、最新状態を読み込み直してください。";
      case "review_state_conflict":
        return "審査状態が更新されたため、いまは申請できません。最新状態を読み込み直します。";
      case "submission_not_ready":
        return "申請条件を満たしていないため送信できません。必要項目を確認してください。";
      case "internal_error":
        return "申請を完了できませんでした。少し時間を置いてからやり直してください。";
    }
  }

  if (error instanceof ApiError && error.code === "network") {
    return "申請を完了できませんでした。通信環境を確認してからやり直してください。";
  }

  return "申請を完了できませんでした。少し時間を置いてからやり直してください。";
}

function shouldSyncCreatorWorkspaceReviewState(error: unknown): boolean {
  if (!(error instanceof CreatorWorkspaceSubmissionReviewApiError)) {
    return false;
  }

  switch (error.code) {
    case "not_found":
    case "review_state_conflict":
    case "submission_not_ready":
      return true;
    default:
      return false;
  }
}

function CreatorWorkspaceReviewPanelLoading() {
  return (
    <section className="grid gap-3 rounded-[24px] border border-[rgba(167,220,249,0.36)] bg-[rgba(248,251,253,0.9)] px-4 py-4">
      <div aria-hidden="true" className="h-4 w-24 animate-pulse rounded-full bg-[rgba(167,220,249,0.32)]" />
      <div aria-hidden="true" className="h-4 w-40 animate-pulse rounded-full bg-[rgba(167,220,249,0.22)]" />
      <div aria-hidden="true" className="h-10 w-28 animate-pulse rounded-[999px] bg-[rgba(167,220,249,0.22)]" />
    </section>
  );
}

function CreatorWorkspaceReviewPanelError({
  message,
  onRetry,
}: {
  message: string;
  onRetry: () => void;
}) {
  return (
    <section className="grid gap-3 rounded-[24px] border border-[rgba(255,184,189,0.84)] bg-[linear-gradient(180deg,rgba(255,247,248,0.98),rgba(255,241,243,0.96))] px-4 py-4">
      <p className="m-0 text-sm leading-6 text-muted" role="alert">
        {message}
      </p>
      <div>
        <Button onClick={onRetry} size="sm" type="button" variant="secondary">
          再読み込み
        </Button>
      </div>
    </section>
  );
}

export function CreatorWorkspaceReviewPanel({
  onRetry,
  onSync,
  state,
}: {
  onRetry: () => void;
  onSync: () => void;
  state: CreatorWorkspaceItemReviewSurfaceState;
}) {
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const resetKey = state.kind === "ready"
    ? `${state.surface.target.kind}:${state.surface.target.id}:${state.surface.package.reviewStatus}:${state.surface.review.state}`
    : state.kind;

  useEffect(() => {
    setErrorMessage(null);
    setIsSubmitting(false);
  }, [resetKey]);

  if (state.kind === "idle") {
    return null;
  }

  if (state.kind === "loading") {
    return <CreatorWorkspaceReviewPanelLoading />;
  }

  if (state.kind === "error") {
    return <CreatorWorkspaceReviewPanelError message={state.message} onRetry={onRetry} />;
  }

  const surface = state.surface;
  const packageBadge = resolveCreatorWorkspacePackageReviewBadge(surface.package.reviewStatus);
  const targetBadge = resolveCreatorWorkspaceObjectReviewBadge(surface.review.state);
  const submitActionLabel = buildCreatorWorkspaceReviewActionLabel(surface.package.submitAction);
  const isShortTarget = surface.target.kind === "short";

  async function handleSubmitReview() {
    if (isSubmitting) {
      return;
    }

    setIsSubmitting(true);
    setErrorMessage(null);

    try {
      await createCreatorWorkspaceSubmissionReview({
        mainId: surface.target.canonicalMainId,
      });

      startTransition(() => {
        onSync();
      });
    } catch (error) {
      setErrorMessage(buildCreatorWorkspaceSubmissionReviewErrorMessage(error));

      if (shouldSyncCreatorWorkspaceReviewState(error)) {
        startTransition(() => {
          onSync();
        });
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <section className="grid gap-3 rounded-[24px] border border-[rgba(167,220,249,0.36)] bg-[rgba(248,251,253,0.9)] px-4 py-4 text-foreground">
      <div className="flex flex-wrap items-center gap-2">
        <span className={`inline-flex min-h-7 items-center justify-center rounded-full px-3 text-[10px] font-bold uppercase tracking-[0.14em] ${getCreatorWorkspaceReviewBadgeClassName(packageBadge.tone)}`}>
          package {packageBadge.label}
        </span>
        {targetBadge ? (
          <span className={`inline-flex min-h-7 items-center justify-center rounded-full px-3 text-[10px] font-bold uppercase tracking-[0.14em] ${getCreatorWorkspaceReviewBadgeClassName(targetBadge.tone)}`}>
            {surface.target.kind === "main" ? "本編" : "ショート"} {targetBadge.label}
          </span>
        ) : null}
      </div>

      <div className="grid gap-1">
        <h3 className="m-0 text-sm font-bold text-foreground">審査状況</h3>
        <p className="m-0 text-[13px] leading-[1.6] text-muted">
          {buildCreatorWorkspaceReviewPackageHeadline(surface.package)}
        </p>
      </div>

      {surface.review.reasonCode ? (
        <div className="grid gap-1 rounded-[18px] border border-[rgba(167,220,249,0.32)] bg-white/72 px-3 py-3">
          <span className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-strong">reason code</span>
          <code className="text-[12px] text-foreground">{surface.review.reasonCode}</code>
        </div>
      ) : null}

      {surface.package.blockers.length > 0 ? (
        <div className="grid gap-2">
          <span className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-strong">申請前の確認</span>
          <ul className="m-0 grid gap-2 pl-5 text-[13px] leading-[1.6] text-muted">
            {surface.package.blockers.map((blockerCode) => (
              <li key={blockerCode}>{resolveCreatorWorkspaceReviewBlockerLabel(blockerCode)}</li>
            ))}
          </ul>
        </div>
      ) : null}

      {isShortTarget ? (
        <p className="m-0 text-[12px] leading-[1.6] text-muted">
          このショートからの submit / resubmit は、linked main package 単位で反映されます。
        </p>
      ) : null}

      {errorMessage ? (
        <p className="m-0 rounded-[18px] bg-[#fff0f1] px-4 py-3 text-[13px] leading-[1.6] text-[#b2394f]" role="alert">
          {errorMessage}
        </p>
      ) : null}

      {submitActionLabel ? (
        <div>
          <Button disabled={isSubmitting} onClick={() => {
            void handleSubmitReview();
          }} type="button">
            {isSubmitting ? "送信中..." : submitActionLabel}
          </Button>
        </div>
      ) : null}
    </section>
  );
}

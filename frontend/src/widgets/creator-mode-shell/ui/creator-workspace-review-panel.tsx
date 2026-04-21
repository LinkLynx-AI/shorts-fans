"use client";

import {
  startTransition,
  useEffect,
  useRef,
  useState,
} from "react";

import {
  createCreatorWorkspaceSubmissionReview,
  CreatorWorkspaceSubmissionReviewApiError,
} from "@/features/creator-workspace-submission-review";
import { ApiError } from "@/shared/api";
import { Button } from "@/shared/ui";

import {
  buildCreatorWorkspaceReviewPackageHeadline,
  hasCreatorWorkspaceReviewIssue,
  resolveCreatorWorkspaceObjectReviewBadge,
  resolveCreatorWorkspacePackageReviewBadge,
  resolveCreatorWorkspaceReviewBlockerLabel,
  resolveCreatorWorkspaceReviewReasonCopy,
  type CreatorWorkspaceItemReviewSurfaceState,
  type CreatorWorkspaceReviewBadge,
} from "../model/creator-workspace-review-surface";

function getCreatorWorkspaceReviewPanelClassName(tone: CreatorWorkspaceReviewBadge["tone"]): string {
  switch (tone) {
    case "revision":
      return "border-[rgba(244,152,45,0.24)] bg-[linear-gradient(180deg,rgba(255,251,245,0.98),rgba(253,245,232,0.94))]";
    case "removed":
      return "border-[rgba(217,77,77,0.24)] bg-[linear-gradient(180deg,rgba(255,248,248,0.98),rgba(255,242,243,0.94))]";
    case "pending":
    case "paused":
    case "hidden":
      return "border-[rgba(167,220,249,0.36)] bg-[linear-gradient(180deg,rgba(251,253,255,0.98),rgba(244,250,253,0.94))]";
    case "approved":
      return "border-[rgba(167,220,249,0.24)] bg-[rgba(248,251,253,0.9)]";
  }
}

function getCreatorWorkspaceReviewAccentClassName(tone: CreatorWorkspaceReviewBadge["tone"]): string {
  switch (tone) {
    case "revision":
      return "bg-[#f4982d]";
    case "removed":
      return "bg-[#d94d4d]";
    case "pending":
    case "paused":
    case "hidden":
      return "bg-[#1082c8]";
    case "approved":
      return "bg-[#34a853]";
  }
}

function getCreatorWorkspaceReviewStatusClassName(tone: CreatorWorkspaceReviewBadge["tone"]): string {
  switch (tone) {
    case "revision":
      return "text-[#8e4e0a]";
    case "removed":
      return "text-[#9f2437]";
    case "pending":
    case "paused":
    case "hidden":
      return "text-[#0a5b8c]";
    case "approved":
      return "text-[#1d6f3a]";
  }
}

function getCreatorWorkspaceReviewPanelTone(
  packageBadge: CreatorWorkspaceReviewBadge | null,
  targetBadge: CreatorWorkspaceReviewBadge | null,
): CreatorWorkspaceReviewBadge["tone"] {
  if (packageBadge?.tone === "removed" || targetBadge?.tone === "removed") {
    return "removed";
  }

  if (packageBadge?.tone === "revision" || targetBadge?.tone === "revision") {
    return "revision";
  }

  return packageBadge?.tone ?? targetBadge?.tone ?? "paused";
}

function buildCreatorWorkspaceReviewIssueTitle(tone: CreatorWorkspaceReviewBadge["tone"]): string {
  switch (tone) {
    case "removed":
      return "公開できません";
    case "revision":
      return "修正が必要です";
    default:
      return "確認が必要です";
  }
}

function buildCreatorWorkspaceSubmissionReviewErrorMessage(error: unknown): string {
  if (error instanceof CreatorWorkspaceSubmissionReviewApiError) {
    switch (error.code) {
      case "auth_required":
        return "ログイン状態を確認してから、もう一度再申請してください。";
      case "creator_mode_unavailable":
        return "creator mode の利用状態を確認してから、もう一度再申請してください。";
      case "not_found":
        return "対象の package を確認できませんでした。最新の状態を再読み込みしてください。";
      case "review_state_conflict":
        return "審査状態が更新されています。最新の状態を確認してから再申請してください。";
      case "submission_not_ready":
        return "再申請に必要な項目がまだ揃っていません。内容を確認してから再申請してください。";
      case "internal_error":
        return "再申請できませんでした。少し時間を置いてからやり直してください。";
    }
  }

  if (error instanceof ApiError && error.code === "network") {
    return "再申請できませんでした。通信環境を確認してからやり直してください。";
  }

  return "再申請できませんでした。少し時間を置いてからやり直してください。";
}

function shouldSyncCreatorWorkspaceReviewState(error: unknown): boolean {
  if (!(error instanceof CreatorWorkspaceSubmissionReviewApiError)) {
    return false;
  }

  return error.code === "not_found"
    || error.code === "review_state_conflict"
    || error.code === "submission_not_ready";
}

function buildCreatorWorkspaceReviewScopeLabel({
  packageBadge,
  targetBadge,
  targetKind,
}: {
  packageBadge: CreatorWorkspaceReviewBadge | null;
  targetBadge: CreatorWorkspaceReviewBadge | null;
  targetKind: "main" | "short";
}): string {
  const targetLabel = targetKind === "main" ? "本編" : "ショート";

  if (packageBadge && targetBadge) {
    if (packageBadge.label === targetBadge.label) {
      return `package / ${targetLabel}`;
    }

    return `package: ${packageBadge.label} / ${targetLabel}: ${targetBadge.label}`;
  }

  if (packageBadge) {
    return "package";
  }

  if (targetBadge) {
    return targetLabel;
  }

  return "package";
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
  onSync,
  onRetry,
  state,
}: {
  onSync: () => void;
  onRetry: () => void;
  state: CreatorWorkspaceItemReviewSurfaceState;
}) {
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const resetKey = state.kind === "ready"
    ? [
        state.surface.requestId,
        state.surface.target.kind,
        state.surface.target.id,
        state.surface.package.canonicalMainId,
        state.surface.package.linkedShortCount,
        state.surface.package.readiness,
        state.surface.package.reviewStatus,
        state.surface.package.submitAction,
        [...state.surface.package.blockers].sort().join(","),
        state.surface.review.state,
        state.surface.review.reasonCode ?? "",
      ].join(":")
    : state.kind;
  const latestResetKeyRef = useRef(resetKey);
  latestResetKeyRef.current = resetKey;

  useEffect(() => {
    setErrorMessage(null);
    setIsSubmitting(false);
  }, [resetKey]);

  if (state.kind === "idle") {
    return null;
  }

  if (state.kind === "loading") {
    return null;
  }

  if (state.kind === "error") {
    return <CreatorWorkspaceReviewPanelError message={state.message} onRetry={onRetry} />;
  }

  const surface = state.surface;
  if (!hasCreatorWorkspaceReviewIssue(surface)) {
    return null;
  }

  const packageBadge = surface.package.reviewStatus === "changes_requested" || surface.package.reviewStatus === "rejected"
    ? resolveCreatorWorkspacePackageReviewBadge(surface.package.reviewStatus)
    : null;
  const targetBadge = resolveCreatorWorkspaceObjectReviewBadge(surface.review.state);
  const packageHeadline = buildCreatorWorkspaceReviewPackageHeadline(surface.package);
  const reasonCopy = resolveCreatorWorkspaceReviewReasonCopy(surface.review.reasonCode);
  const panelTone = getCreatorWorkspaceReviewPanelTone(packageBadge, targetBadge);
  const scopeLabel = buildCreatorWorkspaceReviewScopeLabel({
    packageBadge,
    targetBadge,
    targetKind: surface.target.kind,
  });
  const statusLabel = targetBadge?.label ?? packageBadge?.label ?? "要確認";
  const canResubmit = surface.package.submitAction === "resubmit";

  async function handleResubmit() {
    if (!canResubmit) {
      return;
    }

    const requestResetKey = resetKey;
    const isCurrentRequest = () => latestResetKeyRef.current === requestResetKey;

    setErrorMessage(null);
    setIsSubmitting(true);
    try {
      await createCreatorWorkspaceSubmissionReview({
        mainId: surface.package.canonicalMainId,
      });
      if (!isCurrentRequest()) {
        return;
      }
      startTransition(() => {
        onSync();
      });
    } catch (error) {
      if (!isCurrentRequest()) {
        return;
      }
      setErrorMessage(buildCreatorWorkspaceSubmissionReviewErrorMessage(error));
      if (shouldSyncCreatorWorkspaceReviewState(error)) {
        startTransition(() => {
          onSync();
        });
      }
    } finally {
      if (isCurrentRequest()) {
        setIsSubmitting(false);
      }
    }
  }

  return (
    <section className={`rounded-[18px] border px-4 py-4 text-foreground ${getCreatorWorkspaceReviewPanelClassName(panelTone)}`}>
      <div className="flex items-start gap-3">
        <span aria-hidden="true" className={`mt-[7px] size-2.5 shrink-0 rounded-full ${getCreatorWorkspaceReviewAccentClassName(panelTone)}`} />
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
            <h3 className="m-0 text-[15px] font-bold leading-[1.45] text-foreground">
              {buildCreatorWorkspaceReviewIssueTitle(panelTone)}
            </h3>
            <span className={`text-[12px] font-bold leading-[1.45] ${getCreatorWorkspaceReviewStatusClassName(panelTone)}`}>
              {statusLabel}
            </span>
          </div>
          {packageHeadline ? (
            <p className="m-0 mt-1 text-[13px] leading-[1.65] text-muted">
              {packageHeadline}
            </p>
          ) : null}
        </div>
      </div>

      <dl className="m-0 mt-3 grid gap-2 border-t border-[rgba(7,19,29,0.08)] pt-3 text-[12px] leading-[1.55]">
        <div className="flex items-center justify-between gap-3">
          <dt className="text-muted">対象</dt>
          <dd className="m-0 text-right font-bold text-foreground">{scopeLabel}</dd>
        </div>
        {reasonCopy ? (
          <div className="flex items-start justify-between gap-3">
            <dt className="text-muted">理由</dt>
            <dd className="m-0 min-w-0 text-right font-bold text-foreground">{reasonCopy.label}</dd>
          </div>
        ) : null}
      </dl>

      {reasonCopy ? (
        <p className="m-0 mt-3 border-t border-[rgba(7,19,29,0.08)] pt-3 text-[12px] leading-[1.65] text-muted">
          {reasonCopy.description}
        </p>
      ) : null}

      {surface.package.blockers.length > 0 ? (
        <div className="mt-3 border-t border-[rgba(7,19,29,0.08)] pt-3">
          <p className="m-0 text-[12px] font-bold text-foreground">審査投入前の確認</p>
          <ul className="m-0 mt-2 grid gap-1.5 pl-4 text-[12px] leading-[1.6] text-muted">
            {surface.package.blockers.map((blockerCode) => (
              <li key={blockerCode}>{resolveCreatorWorkspaceReviewBlockerLabel(blockerCode)}</li>
            ))}
          </ul>
        </div>
      ) : null}

      {errorMessage ? (
        <p className="m-0 mt-3 border-t border-[rgba(7,19,29,0.08)] pt-3 text-[12px] leading-[1.65] text-[#9f2437]" role="alert">
          {errorMessage}
        </p>
      ) : null}

      {canResubmit ? (
        <div className="mt-4">
          <Button disabled={isSubmitting} onClick={handleResubmit} type="button">
            {isSubmitting ? "再申請中..." : "再申請する"}
          </Button>
        </div>
      ) : null}
    </section>
  );
}

import { ApiError } from "@/shared/api";
import { getSubmissionReviewReasonOption } from "@/entities/submission-review";

import type {
  CreatorWorkspaceItemReviewSurface,
  CreatorWorkspaceReviewPackageSummary,
  CreatorWorkspaceReviewSurface,
} from "../api/get-creator-workspace-review-surface";
import type { ApprovedCreatorWorkspaceManagedItemTone } from "./approved-creator-workspace";

export type CreatorWorkspaceReviewBadge = {
  label: string;
  tone: ApprovedCreatorWorkspaceManagedItemTone;
};

export type CreatorWorkspaceReviewNotification = {
  detail: string;
  headline: string;
  key: string;
  label: string;
  tone: ApprovedCreatorWorkspaceManagedItemTone;
};

export type CreatorWorkspaceReviewReasonCopy = {
  description: string;
  label: string;
};

export type CreatorWorkspaceReviewSurfaceState =
  | {
      kind: "error";
      message: string;
    }
  | {
      kind: "loading";
    }
  | {
      kind: "ready";
      surface: CreatorWorkspaceReviewSurface;
    };

export type CreatorWorkspaceItemReviewSurfaceState =
  | {
      kind: "error";
      message: string;
    }
  | {
      kind: "idle";
    }
  | {
      kind: "loading";
    }
  | {
      kind: "ready";
      surface: CreatorWorkspaceItemReviewSurface;
    };

type CreatorWorkspaceReviewSurfaceErrorState = Extract<
  CreatorWorkspaceReviewSurfaceState,
  { kind: "error" }
>;
type CreatorWorkspaceReviewSurfaceLoadingState = Extract<
  CreatorWorkspaceReviewSurfaceState,
  { kind: "loading" }
>;
type CreatorWorkspaceItemReviewSurfaceErrorState = Extract<
  CreatorWorkspaceItemReviewSurfaceState,
  { kind: "error" }
>;
type CreatorWorkspaceItemReviewSurfaceLoadingState = Extract<
  CreatorWorkspaceItemReviewSurfaceState,
  { kind: "loading" }
>;

const creatorWorkspaceReviewSurfaceErrorMessage =
  "審査状況を読み込めませんでした。少し時間を置いてから再読み込みしてください。";
const creatorWorkspaceReviewSurfaceNetworkErrorMessage =
  "審査状況を読み込めませんでした。通信環境を確認してから再読み込みしてください。";

export function buildLoadingCreatorWorkspaceReviewSurfaceState(): CreatorWorkspaceReviewSurfaceLoadingState {
  return {
    kind: "loading",
  };
}

export function buildErrorCreatorWorkspaceReviewSurfaceState(
  error: unknown,
): CreatorWorkspaceReviewSurfaceErrorState {
  if (error instanceof ApiError && error.code === "network") {
    return {
      kind: "error",
      message: creatorWorkspaceReviewSurfaceNetworkErrorMessage,
    };
  }

  return {
    kind: "error",
    message: creatorWorkspaceReviewSurfaceErrorMessage,
  };
}

export function buildLoadingCreatorWorkspaceItemReviewSurfaceState(): CreatorWorkspaceItemReviewSurfaceLoadingState {
  return {
    kind: "loading",
  };
}

export function buildErrorCreatorWorkspaceItemReviewSurfaceState(
  error: unknown,
): CreatorWorkspaceItemReviewSurfaceErrorState {
  if (error instanceof ApiError && error.code === "network") {
    return {
      kind: "error",
      message: creatorWorkspaceReviewSurfaceNetworkErrorMessage,
    };
  }

  return {
    kind: "error",
    message: creatorWorkspaceReviewSurfaceErrorMessage,
  };
}

function resolveCreatorWorkspaceReviewTone(
  state: "approved" | "approved_for_publish" | "approved_for_unlock" | "changes_requested" | "draft" | "pending_review" | "rejected" | "revision_requested",
): ApprovedCreatorWorkspaceManagedItemTone {
  switch (state) {
    case "approved":
    case "approved_for_publish":
    case "approved_for_unlock":
      return "approved";
    case "changes_requested":
    case "revision_requested":
      return "revision";
    case "pending_review":
      return "pending";
    case "rejected":
      return "removed";
    case "draft":
      return "paused";
  }
}

function resolveCreatorWorkspaceReviewLabel(
  state: "approved" | "approved_for_publish" | "approved_for_unlock" | "changes_requested" | "draft" | "pending_review" | "rejected" | "revision_requested",
): string | null {
  switch (state) {
    case "approved":
    case "approved_for_publish":
    case "approved_for_unlock":
      return null;
    case "changes_requested":
    case "revision_requested":
      return "差し戻し";
    case "pending_review":
      return "審査中";
    case "rejected":
      return "却下";
    case "draft":
      return "未申請";
  }
}

export function resolveCreatorWorkspaceObjectReviewBadge(
  state: string,
): CreatorWorkspaceReviewBadge | null {
  switch (state) {
    case "rejected":
    case "revision_requested": {
      const label = resolveCreatorWorkspaceReviewLabel(state);
      if (!label) {
        return null;
      }

      return {
        label,
        tone: resolveCreatorWorkspaceReviewTone(state),
      };
    }
    default:
      return null;
  }
}

export function hasCreatorWorkspaceReviewIssue(surface: CreatorWorkspaceItemReviewSurface): boolean {
  const isNormalPackageStatus = surface.package.reviewStatus === "approved"
    || surface.package.reviewStatus === "pending_review";

  return surface.package.blockers.length > 0
    || surface.package.readiness === "blocked"
    || (surface.package.readiness === "conflict" && !isNormalPackageStatus)
    || surface.package.reviewStatus === "changes_requested"
    || surface.package.reviewStatus === "rejected"
    || surface.review.state === "rejected"
    || surface.review.state === "revision_requested";
}

export function resolveCreatorWorkspacePackageReviewBadge(
  reviewStatus: CreatorWorkspaceReviewPackageSummary["reviewStatus"],
): CreatorWorkspaceReviewBadge | null {
  const label = resolveCreatorWorkspaceReviewLabel(reviewStatus);
  if (!label) {
    return null;
  }

  return {
    label,
    tone: resolveCreatorWorkspaceReviewTone(reviewStatus),
  };
}

export function resolveCreatorWorkspaceReviewBlockerLabel(blockerCode: string): string {
  switch (blockerCode) {
    case "linked_short_missing":
      return "linked short がまだありません。";
    case "main_asset_not_ready":
      return "本編動画の processing が完了していません。";
    case "short_asset_not_ready":
      return "ショート動画の processing が完了していません。";
    case "main_price_missing":
      return "本編価格が未設定です。";
    case "ownership_missing":
      return "ownership の確認が未完了です。";
    case "consent_missing":
      return "consent の確認が未完了です。";
    default:
      return blockerCode;
  }
}

export function resolveCreatorWorkspaceReviewReasonCopy(reasonCode: string | null): CreatorWorkspaceReviewReasonCopy | null {
  if (reasonCode === null) {
    return null;
  }

  const reasonOption = getSubmissionReviewReasonOption(reasonCode);
  if (reasonOption) {
    return {
      description: reasonOption.description,
      label: reasonOption.label,
    };
  }

  return {
    description: "詳細は運営からの案内を確認してください。",
    label: "審査基準の確認が必要です",
  };
}

export function buildCreatorWorkspaceReviewPackageHeadline(
  summary: CreatorWorkspaceReviewPackageSummary,
): string | null {
  switch (summary.reviewStatus) {
    case "changes_requested":
      switch (summary.readiness) {
        case "ready":
          return "修正内容を確認してください。";
        case "blocked":
          return "再審査前に必要項目を満たしてください。";
        case "conflict":
          return "現在の審査状態では再審査できません。";
        case "none":
          return "修正内容を確認してください。";
      }
    case "rejected":
      return "審査で公開不可となったため、この動画は再申請できません。";
    case "pending_review":
      return null;
    case "approved":
      return null;
    case "draft":
      switch (summary.readiness) {
        case "ready":
          return "自動審査投入の反映を待っています。";
        case "blocked":
          return "審査投入前に必要項目を満たしてください。";
        case "conflict":
          return "現在の審査状態では審査投入できません。";
        case "none":
          return "この動画はまだ審査投入待ちです。";
      }
  }
}

function formatCount(value: number): string {
  return value.toLocaleString("ja-JP");
}

export function deriveCreatorWorkspaceReviewNotifications(
  packages: readonly CreatorWorkspaceReviewPackageSummary[],
): readonly CreatorWorkspaceReviewNotification[] {
  let changesRequestedCount = 0;
  let rejectedCount = 0;

  for (const item of packages) {
    switch (item.reviewStatus) {
      case "changes_requested":
        changesRequestedCount += 1;
        break;
      case "rejected":
        rejectedCount += 1;
        break;
    }
  }

  return [
    changesRequestedCount > 0 ? {
      detail: "修正内容を確認",
      headline: `差し戻し ${formatCount(changesRequestedCount)}件`,
      key: "changes_requested",
      label: "差し戻し",
      tone: "revision",
    } : null,
    rejectedCount > 0 ? {
      detail: "該当動画の確認をお願いします",
      headline: `公開不可 ${formatCount(rejectedCount)}件`,
      key: "rejected",
      label: "公開不可",
      tone: "removed",
    } : null,
  ].filter((item): item is CreatorWorkspaceReviewNotification => item !== null);
}

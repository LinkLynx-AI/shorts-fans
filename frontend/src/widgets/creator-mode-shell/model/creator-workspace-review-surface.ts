import { ApiError } from "@/shared/api";

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
): string {
  switch (state) {
    case "approved":
    case "approved_for_publish":
    case "approved_for_unlock":
      return "承認済み";
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
    case "approved_for_publish":
    case "approved_for_unlock":
    case "draft":
    case "pending_review":
    case "rejected":
    case "revision_requested":
      return {
        label: resolveCreatorWorkspaceReviewLabel(state),
        tone: resolveCreatorWorkspaceReviewTone(state),
      };
    default:
      return null;
  }
}

export function resolveCreatorWorkspacePackageReviewBadge(
  reviewStatus: CreatorWorkspaceReviewPackageSummary["reviewStatus"],
): CreatorWorkspaceReviewBadge {
  return {
    label: resolveCreatorWorkspaceReviewLabel(reviewStatus),
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

export function buildCreatorWorkspaceReviewActionLabel(
  action: CreatorWorkspaceReviewPackageSummary["submitAction"],
): string | null {
  switch (action) {
    case "submit":
      return "審査へ申請";
    case "resubmit":
      return "再申請する";
    case "none":
      return null;
  }
}

export function buildCreatorWorkspaceReviewPackageHeadline(
  summary: CreatorWorkspaceReviewPackageSummary,
): string {
  switch (summary.reviewStatus) {
    case "changes_requested":
      switch (summary.readiness) {
        case "ready":
          return summary.submitAction === "resubmit"
            ? "修正後に再申請できます。"
            : "修正内容を確認してください。";
        case "blocked":
          return "再申請前に必要項目を満たしてください。";
        case "conflict":
          return "現在の審査状態では再申請できません。";
        case "none":
          return "修正内容を確認してください。";
      }
    case "rejected":
      return "却下されたため、この package は self-serve で再申請できません。";
    case "pending_review":
      return "審査結果を待っています。";
    case "approved":
      return "現在の package は承認済みです。";
    case "draft":
      switch (summary.readiness) {
        case "ready":
          return "この package は審査へ申請できます。";
        case "blocked":
          return "申請前に必要項目を満たしてください。";
        case "conflict":
          return "現在の審査状態では申請できません。";
        case "none":
          return "この package はまだ申請待ちです。";
      }
  }
}

function formatCount(value: number): string {
  return value.toLocaleString("ja-JP");
}

function buildNotificationDetail(count: number, detail: string): string {
  return `${formatCount(count)}件の package が対象です。${detail}`;
}

export function deriveCreatorWorkspaceReviewNotifications(
  packages: readonly CreatorWorkspaceReviewPackageSummary[],
): readonly CreatorWorkspaceReviewNotification[] {
  const changesRequestedCount = packages.filter((item) => item.reviewStatus === "changes_requested").length;
  const rejectedCount = packages.filter((item) => item.reviewStatus === "rejected").length;
  const readyDraftCount = packages.filter((item) => item.reviewStatus === "draft" && item.readiness === "ready").length;
  const blockedDraftCount = packages.filter((item) => item.reviewStatus === "draft" && item.readiness === "blocked").length;
  const pendingReviewCount = packages.filter((item) => item.reviewStatus === "pending_review").length;

  return [
    changesRequestedCount > 0 ? {
      detail: buildNotificationDetail(changesRequestedCount, "detail から修正後の再申請を進めてください。"),
      headline: `差し戻し対応が${formatCount(changesRequestedCount)}件あります`,
      key: "changes_requested",
      label: "差し戻し",
      tone: "revision",
    } : null,
    rejectedCount > 0 ? {
      detail: buildNotificationDetail(rejectedCount, "self-serve では再申請できないため、内容確認が必要です。"),
      headline: `却下された package が${formatCount(rejectedCount)}件あります`,
      key: "rejected",
      label: "却下",
      tone: "removed",
    } : null,
    readyDraftCount > 0 ? {
      detail: buildNotificationDetail(readyDraftCount, "detail からそのまま審査へ申請できます。"),
      headline: `申請できる package が${formatCount(readyDraftCount)}件あります`,
      key: "ready_draft",
      label: "未申請",
      tone: "paused",
    } : null,
    blockedDraftCount > 0 ? {
      detail: buildNotificationDetail(blockedDraftCount, "申請前に価格や processing 状態を確認してください。"),
      headline: `申請前の確認が必要な package が${formatCount(blockedDraftCount)}件あります`,
      key: "blocked_draft",
      label: "要確認",
      tone: "paused",
    } : null,
    pendingReviewCount > 0 ? {
      detail: buildNotificationDetail(pendingReviewCount, "審査結果が出るまで detail からの再申請はできません。"),
      headline: `審査中の package が${formatCount(pendingReviewCount)}件あります`,
      key: "pending_review",
      label: "審査中",
      tone: "pending",
    } : null,
  ].filter((item): item is CreatorWorkspaceReviewNotification => item !== null);
}

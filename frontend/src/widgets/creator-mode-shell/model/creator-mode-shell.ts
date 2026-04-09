import {
  getCreatorById,
  type CreatorSummary,
} from "@/entities/creator";
import type { CurrentViewer } from "@/entities/viewer";
import type { CreatorModeNavigationKey } from "@/features/creator-mode-navigation";
import {
  getMockApprovedCreatorWorkspaceState,
  type ApprovedCreatorWorkspaceState,
} from "./approved-creator-workspace";

export type CreatorModeShellBlockedState = {
  ctaHref: string;
  ctaLabel: string;
  description: string;
  eyebrow: string;
  kind: "capability_required" | "mode_required" | "unauthenticated";
  title: string;
};

export type CreatorModeShellReadyState = {
  activeNavigation: CreatorModeNavigationKey;
  creator: CreatorSummary;
  kind: "ready";
  workspace: ApprovedCreatorWorkspaceState;
};

export type CreatorModeShellState = CreatorModeShellBlockedState | CreatorModeShellReadyState;

function getMockCreatorModeOwner(): CreatorSummary {
  const creator = getCreatorById("creator_mina_rei");

  if (!creator) {
    throw new Error("creator shell mock owner is missing");
  }

  return creator;
}

/**
 * creator mode route 用の mock shell state を返す。
 */
export function getMockCreatorModeShellState(
  activeNavigation: CreatorModeNavigationKey = "dashboard",
): CreatorModeShellReadyState {
  const creator = getMockCreatorModeOwner();

  return {
    activeNavigation,
    creator,
    kind: "ready",
    workspace: getMockApprovedCreatorWorkspaceState(creator.id),
  };
}

/**
 * current viewer bootstrap から creator mode route の shell state を解決する。
 */
export function resolveCreatorModeShellState(
  currentViewer: CurrentViewer | null,
  activeNavigation: CreatorModeNavigationKey = "dashboard",
): CreatorModeShellState {
  if (currentViewer === null) {
    return {
      ctaHref: "/login",
      ctaLabel: "ログインへ進む",
      description: "creator mode は private workspace なので、まず同じ identity でログインしてください。",
      eyebrow: "Creator access",
      kind: "unauthenticated",
      title: "creator mode を開くにはログインが必要です。",
    };
  }

  if (!currentViewer.canAccessCreatorMode) {
    return {
      ctaHref: "/",
      ctaLabel: "フィードへ戻る",
      description:
        "この viewer には creator capability がまだ付与されていません。creator onboarding 完了後に creator mode が解放されます。",
      eyebrow: "Creator access",
      kind: "capability_required",
      title: "creator mode はまだ利用できません。",
    };
  }

  if (currentViewer.activeMode !== "creator") {
    return {
      ctaHref: "/",
      ctaLabel: "フィードへ戻る",
      description:
        "この route は creator mode 前提です。mode switch 自体は別 PR の担当なので、profile / account menu から creator mode に入ってください。",
      eyebrow: "Mode mismatch",
      kind: "mode_required",
      title: "creator mode に切り替えてから開いてください。",
    };
  }

  return getMockCreatorModeShellState(activeNavigation);
}

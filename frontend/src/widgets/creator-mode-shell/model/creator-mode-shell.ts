import {
  getCreatorById,
  type CreatorSummary,
} from "@/entities/creator";
import type { CurrentViewer } from "@/entities/viewer";
import type { CreatorModeNavigationKey } from "@/features/creator-mode-navigation";
import type { RouteStructureItem } from "@/widgets/route-structure-panel";

type CreatorModeContextBadge = {
  key: string;
  label: string;
  value: string;
};

type CreatorModeShellSlot = {
  description: string;
  key: string;
  label: string;
};

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
  contextBadges: readonly CreatorModeContextBadge[];
  creator: CreatorSummary;
  description: string;
  eyebrow: string;
  kind: "ready";
  slots: readonly CreatorModeShellSlot[];
  structureItems: readonly RouteStructureItem[];
  title: string;
};

export type CreatorModeShellState = CreatorModeShellBlockedState | CreatorModeShellReadyState;

const creatorModeStructureItems = [
  {
    description: "approved creator 用の landing と summary blocks は後続 task がこの shell 上に載せます。",
    key: "dashboard",
    label: "Dashboard home",
  },
  {
    description: "main + short の submission package を扱う upload/import 導線をここからつなぎます。",
    key: "upload",
    label: "Upload flow",
  },
  {
    description: "canonical main と複数 short の関係を整理する linkage 面を creator mode 配下に分離します。",
    key: "linkage",
    label: "Linkage workspace",
  },
  {
    description: "review / moderation 状態の可視化と後続の remediation 導線をここへ載せます。",
    key: "review",
    label: "Review state",
  },
] as const satisfies readonly RouteStructureItem[];

const creatorModeShellSlots = [
  {
    description: "approved creator 向けの summary surface がここへ載る前提です。",
    key: "summary",
    label: "Summary slot",
  },
  {
    description: "投稿管理、一覧、review state などの workspace content をここへ差し込みます。",
    key: "content",
    label: "Content slot",
  },
] as const satisfies readonly CreatorModeShellSlot[];

const creatorModeContextBadges = [
  {
    key: "home",
    label: "Home",
    value: "Dashboard",
  },
  {
    key: "scope",
    label: "Scope",
    value: "Private workspace",
  },
  {
    key: "identity",
    label: "Model",
    value: "1 user + active mode",
  },
] as const satisfies readonly CreatorModeContextBadge[];

function getMockCreatorModeOwner(): CreatorSummary {
  const creator = getCreatorById("creator_mina_rei");

  if (!creator) {
    throw new Error("creator shell mock owner is missing");
  }

  return creator;
}

/**
 * `/creator` 用の mock shell state を返す。
 */
export function getMockCreatorModeShellState(): CreatorModeShellReadyState {
  return {
    activeNavigation: "dashboard",
    contextBadges: creatorModeContextBadges,
    creator: getMockCreatorModeOwner(),
    description:
      "fan mode と分離した creator private surface として、dashboard・upload・linkage・review を同じ shell で受けられるようにします。",
    eyebrow: "Creator mode",
    kind: "ready",
    slots: creatorModeShellSlots,
    structureItems: creatorModeStructureItems,
    title: "Dashboard shell",
  };
}

/**
 * current viewer bootstrap から `/creator` の shell state を解決する。
 */
export function resolveCreatorModeShellState(currentViewer: CurrentViewer | null): CreatorModeShellState {
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

  return getMockCreatorModeShellState();
}

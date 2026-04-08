export type CreatorModeNavigationKey = "dashboard" | "upload" | "linkage" | "review";

export type CreatorModeNavigationItem = {
  description: string;
  href: string;
  key: CreatorModeNavigationKey;
  label: string;
};

const creatorModeNavigationItems = [
  {
    description: "creator mode の home で、workspace 全体の入口を担います。",
    href: "/creator",
    key: "dashboard",
    label: "Dashboard",
  },
  {
    description: "main と short の submission package を追加する upload 入口です。",
    href: "/creator/upload",
    key: "upload",
    label: "Upload",
  },
  {
    description: "short と canonical main の紐付けを整理する linkage 面です。",
    href: "/creator/linkage",
    key: "linkage",
    label: "Linkage",
  },
  {
    description: "review / moderation 状態を確認する private surface です。",
    href: "/creator/review",
    key: "review",
    label: "Review",
  },
] as const satisfies readonly CreatorModeNavigationItem[];

/**
 * creator mode の主要遷移を返す。
 */
export function getCreatorModeNavigationItems(): readonly CreatorModeNavigationItem[] {
  return creatorModeNavigationItems;
}

/**
 * 現在の run で遷移可能な creator navigation かを判定する。
 */
export function isCreatorModeNavigationAvailable(key: CreatorModeNavigationKey): boolean {
  return key === "dashboard";
}

/**
 * pathname から active な creator navigation key を解決する。
 */
export function resolveActiveCreatorModeNavigation(pathname: string): CreatorModeNavigationKey {
  if (pathname.startsWith("/creator/upload")) {
    return "upload";
  }

  if (pathname.startsWith("/creator/linkage")) {
    return "linkage";
  }

  if (pathname.startsWith("/creator/review")) {
    return "review";
  }

  return "dashboard";
}

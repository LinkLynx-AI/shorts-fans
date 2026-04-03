import { CircleUserRound, Search, SquarePlay } from "lucide-react";
import type { LucideIcon } from "lucide-react";

export type FanNavigationKey = "fan" | "feed" | "search";

export type FanNavigationItem = {
  ariaLabel: string;
  href: string;
  icon: LucideIcon;
  key: FanNavigationKey;
};

const navigationItems = [
  { ariaLabel: "フィード", href: "/", icon: SquarePlay, key: "feed" },
  { ariaLabel: "検索", href: "/search", icon: Search, key: "search" },
  { ariaLabel: "マイ", href: "/fan", icon: CircleUserRound, key: "fan" },
] as const satisfies readonly FanNavigationItem[];

/**
 * bottom navigation の item 定義を返す。
 */
export function getFanNavigationItems(): readonly FanNavigationItem[] {
  return navigationItems;
}

/**
 * pathname から active な navigation key を解決する。
 */
export function resolveActiveFanNavigation(pathname: string): FanNavigationKey {
  if (pathname.startsWith("/search")) {
    return "search";
  }

  if (pathname.startsWith("/fan")) {
    return "fan";
  }

  return "feed";
}

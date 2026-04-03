import { House, Search, UserRound } from "lucide-react";
import type { LucideIcon } from "lucide-react";

export type FanNavigationKey = "fan" | "feed" | "search";

export type FanNavigationItem = {
  href: string;
  icon: LucideIcon;
  key: FanNavigationKey;
  label: string;
};

const navigationItems = [
  { href: "/", icon: House, key: "feed", label: "Feed" },
  { href: "/search", icon: Search, key: "search", label: "Search" },
  { href: "/fan", icon: UserRound, key: "fan", label: "My" },
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

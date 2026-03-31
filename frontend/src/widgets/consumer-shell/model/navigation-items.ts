import { Compass, Film, House, UserRound, type LucideIcon } from "lucide-react";

export type ConsumerNavigationItem = {
  description: string;
  href: "/" | "/home" | "/profile" | "/subscriptions";
  icon: LucideIcon;
  label: string;
};

export const consumerNavigationItems: ConsumerNavigationItem[] = [
  {
    description: "discovery hub",
    href: "/home",
    icon: House,
    label: "home",
  },
  {
    description: "primary lane",
    href: "/",
    icon: Film,
    label: "shorts",
  },
  {
    description: "retention feed",
    href: "/subscriptions",
    icon: Compass,
    label: "subscriptions",
  },
  {
    description: "account",
    href: "/profile",
    icon: UserRound,
    label: "profile",
  },
];

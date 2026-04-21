import {
  Clapperboard,
  IdCard,
  LayoutDashboard,
  type LucideIcon,
} from "lucide-react";
import Link from "next/link";

import { cn } from "@/shared/lib";

export type AdminReviewSection = "creator-reviews" | "home" | "submission-reviews";

type AdminReviewNavigationProps = {
  active: AdminReviewSection;
};

type AdminReviewNavigationItem = {
  description: string;
  href: string;
  icon: LucideIcon;
  key: AdminReviewSection;
  label: string;
};

const adminReviewNavigationItems: readonly AdminReviewNavigationItem[] = [
  {
    description: "審査入口",
    href: "/admin",
    icon: LayoutDashboard,
    key: "home",
    label: "Admin",
  },
  {
    description: "登録申請",
    href: "/admin/creator-reviews",
    icon: IdCard,
    key: "creator-reviews",
    label: "Creator 審査",
  },
  {
    description: "main / short",
    href: "/admin/submission-reviews",
    icon: Clapperboard,
    key: "submission-reviews",
    label: "Video 審査",
  },
];

/**
 * admin review surface 間を移動する route-local navigation を表示する。
 */
export function AdminReviewNavigation({ active }: AdminReviewNavigationProps) {
  return (
    <nav aria-label="Admin review navigation" className="w-full">
      <div className="grid gap-2 rounded-[24px] border border-border bg-white p-2 shadow-[0_14px_32px_rgba(15,23,42,0.06)] sm:grid-cols-3">
        {adminReviewNavigationItems.map((item) => {
          const Icon = item.icon;
          const isActive = active === item.key;

          return (
            <Link
              aria-current={isActive ? "page" : undefined}
              className={cn(
                "flex min-h-[64px] items-center gap-3 rounded-[18px] px-4 py-3 transition",
                isActive
                  ? "border border-border bg-surface-subtle text-foreground shadow-[0_8px_18px_rgba(15,23,42,0.06)]"
                  : "text-muted hover:bg-surface-subtle hover:text-foreground",
              )}
              href={item.href}
              key={item.key}
              prefetch={false}
            >
              <span
                className={cn(
                  "flex size-9 shrink-0 items-center justify-center rounded-full border",
                  isActive
                    ? "border-[#d7e6f5] bg-[#f5faff] text-[#1f628f]"
                    : "border-border bg-white text-muted",
                )}
              >
                <Icon aria-hidden="true" size={17} strokeWidth={2.2} />
              </span>
              <span className="min-w-0">
                <span className="block text-sm font-semibold tracking-[-0.02em]">{item.label}</span>
                <span className="mt-0.5 block text-[12px] leading-4 text-muted">{item.description}</span>
              </span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}

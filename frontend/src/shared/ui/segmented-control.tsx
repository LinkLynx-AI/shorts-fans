import Link from "next/link";
import type { ReactNode } from "react";

import { cn } from "@/shared/lib";

export type SegmentedControlItem = {
  active?: boolean;
  href: string;
  icon?: ReactNode;
  key: string;
  label: string;
  prefetch?: boolean;
};

type SegmentedControlProps = {
  ariaLabel: string;
  className?: string;
  items: readonly SegmentedControlItem[];
  variant?: "pill" | "underline";
};

/**
 * route 遷移に使う segmented control を表示する。
 */
export function SegmentedControl({
  ariaLabel,
  className,
  items,
  variant = "pill",
}: SegmentedControlProps) {
  return (
    <nav aria-label={ariaLabel} className={cn(variant === "pill" ? "inline-flex" : "flex justify-center", className)}>
      <div
        className={cn(
          "inline-flex items-center",
          variant === "pill"
            ? "gap-1 rounded-full border border-border bg-surface-subtle p-1 shadow-[0_8px_20px_rgba(15,23,42,0.05)]"
            : "gap-7 border-b border-border",
        )}
      >
        {items.map((item) => (
          <Link
            {...(item.prefetch !== undefined ? { prefetch: item.prefetch } : {})}
            key={item.key}
            aria-current={item.active ? "page" : undefined}
            className={cn(
              "inline-flex items-center justify-center gap-2 transition",
              variant === "pill"
                ? "min-h-10 rounded-full px-4 text-[14px] font-semibold"
                : "relative min-h-10 pb-3 text-[15px] font-semibold",
              item.active
                ? variant === "pill"
                  ? "border border-border bg-white text-foreground shadow-[0_8px_18px_rgba(15,23,42,0.07)]"
                  : "text-accent-ink after:absolute after:right-0 after:bottom-0 after:left-0 after:h-0.5 after:rounded-full after:bg-accent"
                : variant === "pill"
                  ? "text-muted hover:text-foreground"
                  : "text-muted hover:text-foreground",
            )}
            href={item.href}
          >
            <span className="inline-flex items-center justify-center gap-2">
              {item.icon ? <span className="inline-flex shrink-0 items-center justify-center">{item.icon}</span> : null}
              <span>{item.label}</span>
            </span>
          </Link>
        ))}
      </div>
    </nav>
  );
}

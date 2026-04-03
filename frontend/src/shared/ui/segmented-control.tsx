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
            ? "gap-1.5 rounded-full border border-white/72 bg-white/80 p-1.5 shadow-[0_10px_24px_rgba(36,94,132,0.12)] backdrop-blur-md"
            : "gap-6",
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
                ? "min-h-10 rounded-full px-4 text-[11px] font-semibold uppercase tracking-[0.14em]"
                : "relative min-h-8 pb-2 text-[15px] font-semibold",
              item.active
                ? variant === "pill"
                  ? "bg-accent-strong text-white shadow-[0_10px_22px_rgba(16,130,200,0.22)]"
                  : "text-white after:absolute after:right-0 after:bottom-0 after:left-0 after:h-0.5 after:rounded-full after:bg-white"
                : variant === "pill"
                  ? "text-accent-strong hover:bg-white/84"
                  : "text-white/68 hover:text-white/84",
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

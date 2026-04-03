import type { ReactNode } from "react";

import type { ShortPreviewMeta } from "../model/short";
import { getShortThemeStyle } from "../model/short";
import { cn } from "@/shared/lib";

type ShortMetaPillProps = {
  children: ReactNode;
  className?: string;
};

type ShortPosterProps = {
  className?: string;
  meta?: string;
  short: ShortPreviewMeta;
  title?: string;
  variant?: "grid" | "hero";
};

/**
 * short metadata を表示する pill を描画する。
 */
export function ShortMetaPill({ children, className }: ShortMetaPillProps) {
  return (
    <span
      className={cn(
        "inline-flex min-h-8 items-center rounded-full bg-white/18 px-3 text-[11px] font-semibold uppercase tracking-[0.14em] text-white backdrop-blur-sm",
        className,
      )}
    >
      {children}
    </span>
  );
}

/**
 * short surface 用の poster tile を表示する。
 */
export function ShortPoster({
  className,
  meta,
  short,
  title,
  variant = "grid",
}: ShortPosterProps) {
  return (
    <div
      className={cn(
        "relative overflow-hidden border border-white/24 text-white shadow-[0_20px_48px_rgba(7,19,29,0.24)]",
        variant === "hero" ? "aspect-[3/5] rounded-[34px]" : "aspect-[3/4] rounded-[8px]",
        className,
      )}
      style={getShortThemeStyle(short)}
    >
      <div className="absolute inset-0 bg-[linear-gradient(180deg,var(--short-tile-top)_0%,var(--short-tile-mid)_42%,var(--short-tile-bottom)_100%)]" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.24),transparent_34%)]" />
      <div className="absolute inset-x-0 bottom-0 space-y-2 bg-[linear-gradient(180deg,rgba(6,21,33,0)_0%,rgba(6,21,33,0.72)_100%)] px-4 pb-4 pt-10">
        {meta ? <ShortMetaPill>{meta}</ShortMetaPill> : null}
        <p className={cn("font-display leading-tight tracking-[-0.04em]", variant === "hero" ? "text-3xl" : "text-sm")}>
          {title ?? short.title}
        </p>
      </div>
    </div>
  );
}

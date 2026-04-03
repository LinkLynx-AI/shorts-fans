import Link from "next/link";

import type { ShortPreviewMeta } from "@/entities/short";
import { cn } from "@/shared/lib";

export type UnlockCtaProps = {
  className?: string;
  href?: string;
  short: ShortPreviewMeta;
  state?: "locked" | "unlocked";
};

function getUnlockMeta(short: ShortPreviewMeta, state: UnlockCtaProps["state"]) {
  if (state === "unlocked") {
    return short.progress;
  }

  return `${short.price} | ${short.duration.replace("分", "m")}`;
}

/**
 * short から main unlock へつなぐ CTA を表示する。
 */
export function UnlockCta({
  className,
  href,
  short,
  state = "locked",
}: UnlockCtaProps) {
  const content = (
    <span className="flex min-w-0 flex-1 items-center justify-between gap-3">
      <span className="truncate text-[15px] font-semibold tracking-[-0.01em]">
        {state === "unlocked" ? "Continue main" : "Unlock"}
      </span>
      <span
        className={cn(
          "inline-flex min-h-[34px] shrink-0 items-center rounded-full px-3.5 text-xs font-semibold tracking-[-0.01em]",
          state === "unlocked"
            ? "bg-accent-strong/12 text-accent-strong"
            : "bg-[linear-gradient(135deg,var(--accent)_0%,var(--accent-strong)_100%)] text-white shadow-[0_8px_20px_rgba(16,130,200,0.2)]",
        )}
      >
        {getUnlockMeta(short, state)}
      </span>
    </span>
  );

  const classes = cn(
    "flex min-h-12 items-center justify-between gap-3 rounded-full border border-[#bae7ff]/90 bg-[linear-gradient(90deg,rgba(240,251,255,0.97),rgba(213,243,255,0.93))] px-3 py-1.5 text-left text-foreground shadow-[0_18px_44px_rgba(36,94,132,0.14)] backdrop-blur-xl",
    state === "unlocked" &&
      "border-[#85cdf1]/92 bg-[linear-gradient(90deg,rgba(225,244,255,0.98),rgba(204,235,252,0.96))]",
    className,
  );

  if (!href) {
    return <div className={classes}>{content}</div>;
  }

  return (
    <Link className={cn(classes, "transition hover:translate-y-[-1px]")} href={href}>
      {content}
    </Link>
  );
}

import Link from "next/link";
import type { CSSProperties, ReactNode } from "react";
import { ArrowLeft } from "lucide-react";

import { cn } from "@/shared/lib";
import { Button } from "@/shared/ui";

type DetailShellProps = {
  backHref: string;
  backButtonClassName?: string;
  bodyClassName?: string;
  children?: ReactNode;
  headerContent?: ReactNode;
  style?: CSSProperties;
  variant?: "immersive" | "surface";
};

/**
 * short detail / creator profile 共通の page shell を表示する。
 */
export function DetailShell({
  backHref,
  backButtonClassName,
  bodyClassName,
  children,
  headerContent,
  style,
  variant = "surface",
}: DetailShellProps) {
  if (variant === "immersive") {
    return (
      <section className="relative min-h-full overflow-y-auto px-5 pb-28 pt-3 text-white" style={style}>
        <div className="absolute inset-0 bg-[linear-gradient(180deg,var(--short-bg-start)_0%,var(--short-bg-mid)_54%,var(--short-bg-end)_100%)]" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.22),transparent_34%)]" />

        <div className="relative space-y-5">
          <Button
            asChild
            className="border border-white/18 bg-black/12 text-white backdrop-blur-sm hover:bg-black/24 hover:text-white"
            size="icon"
            variant="ghost"
          >
            <Link aria-label="Back" href={backHref}>
              <ArrowLeft className="size-5" strokeWidth={2.1} />
            </Link>
          </Button>

          {children}
        </div>
      </section>
    );
  }

  return (
    <section className="min-h-full overflow-y-auto bg-white px-5 pb-28 pt-3 text-foreground">
      <div className={cn("flex min-h-11 items-center gap-3", headerContent ? "border-b border-border pb-3" : "")}>
        <Button
          asChild
          className={cn("bg-surface-subtle hover:bg-accent-soft", backButtonClassName)}
          size="icon"
          variant="ghost"
        >
          <Link aria-label="Back" href={backHref}>
            <ArrowLeft className="size-5" strokeWidth={2.1} />
          </Link>
        </Button>
        {headerContent ? <div className="min-w-0 flex-1 text-center">{headerContent}</div> : null}
        {headerContent ? <div aria-hidden="true" className="size-10 shrink-0" /> : null}
      </div>
      <div className={cn(headerContent ? "mt-4" : "mt-5", bodyClassName)}>{children}</div>
    </section>
  );
}

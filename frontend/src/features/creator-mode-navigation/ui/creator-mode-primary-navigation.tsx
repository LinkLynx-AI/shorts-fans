import Link from "next/link";
import {
  Clapperboard,
  LayoutDashboard,
  Link2,
  ShieldCheck,
  type LucideIcon,
} from "lucide-react";

import { cn } from "@/shared/lib";

import {
  getCreatorModeNavigationItems,
  isCreatorModeNavigationAvailable,
  type CreatorModeNavigationKey,
} from "../model/creator-mode-navigation";

const navigationIconByKey: Record<CreatorModeNavigationKey, LucideIcon> = {
  dashboard: LayoutDashboard,
  linkage: Link2,
  review: ShieldCheck,
  upload: Clapperboard,
};

type CreatorModePrimaryNavigationProps = {
  activeKey: CreatorModeNavigationKey;
  className?: string;
  variant?: "cards" | "compact";
};

type CreatorModeNavigationCardProps = {
  active: boolean;
  description: string;
  href: string;
  label: string;
  navigationKey: CreatorModeNavigationKey;
};

function CreatorModeNavigationCard({
  active,
  description,
  href,
  label,
  navigationKey,
}: CreatorModeNavigationCardProps) {
  const Icon = navigationIconByKey[navigationKey];
  const available = isCreatorModeNavigationAvailable(navigationKey);
  const cardClassName = cn(
    "group flex min-h-[112px] flex-col justify-between rounded-[24px] border px-4 py-4 text-left transition",
    active
      ? "border-[#d8e8f7] bg-accent-soft text-foreground shadow-[0_16px_36px_rgba(15,23,42,0.08)]"
      : available
        ? "border-border bg-white text-foreground shadow-[0_12px_28px_rgba(15,23,42,0.06)] hover:-translate-y-0.5 hover:bg-surface-subtle"
        : "border-border bg-surface-subtle text-muted shadow-[0_10px_24px_rgba(15,23,42,0.04)]",
  );

  if (available) {
    return (
      <Link
        aria-current={active ? "page" : undefined}
        aria-label={label}
        className={cardClassName}
        href={href}
      >
        <div className="flex items-start justify-between gap-3">
          <span
            className={cn(
              "inline-flex size-10 items-center justify-center rounded-full border",
              active ? "border-[#d4e6f6] bg-white text-accent-ink" : "border-border bg-surface-subtle text-accent-ink",
            )}
          >
            <Icon aria-hidden="true" className="size-4.5" strokeWidth={1.9} />
          </span>
          <span
            className={cn(
              "inline-flex min-h-7 items-center rounded-full px-3 text-[10px] font-semibold uppercase tracking-[0.18em]",
              active ? "bg-white text-accent-ink" : "bg-accent-soft text-accent-ink",
            )}
          >
            live
          </span>
        </div>
        <div className="mt-4">
          <p className="text-[15px] font-semibold tracking-[-0.03em]">{label}</p>
          <p className={cn("mt-2 text-sm leading-6", active ? "text-muted-strong" : "text-muted")}>{description}</p>
        </div>
      </Link>
    );
  }

  return (
    <div aria-disabled="true" className={cardClassName}>
      <div className="flex items-start justify-between gap-3">
        <span className="inline-flex size-10 items-center justify-center rounded-full border border-border bg-white text-muted">
          <Icon aria-hidden="true" className="size-4.5" strokeWidth={1.9} />
        </span>
        <span className="inline-flex min-h-7 items-center rounded-full border border-border bg-white px-3 text-[10px] font-semibold uppercase tracking-[0.18em] text-muted">
          soon
        </span>
      </div>
      <div className="mt-4">
        <p className="text-[15px] font-semibold tracking-[-0.03em] text-foreground">{label}</p>
        <p className="mt-2 text-sm leading-6 text-muted">{description}</p>
      </div>
    </div>
  );
}

/**
 * creator mode の primary navigation を表示する。
 */
export function CreatorModePrimaryNavigation({
  activeKey,
  className,
  variant = "cards",
}: CreatorModePrimaryNavigationProps) {
  if (variant === "compact") {
    return (
      <nav aria-label="Creator primary" className={className}>
        <div className="grid grid-cols-4 gap-2">
          {getCreatorModeNavigationItems().map((item) => {
            const active = item.key === activeKey;
            const available = isCreatorModeNavigationAvailable(item.key);

            if (available) {
              return (
                <Link
                  aria-current={active ? "page" : undefined}
                  aria-label={item.label}
                  className={cn(
                    "inline-flex min-h-9 items-center justify-center rounded-full border px-3 text-[12px] font-semibold tracking-[-0.01em] transition",
                    active
                      ? "border-[#d8e8f7] bg-accent-soft text-accent-ink shadow-[0_10px_24px_rgba(15,23,42,0.06)]"
                      : "border-border bg-white text-foreground shadow-[0_8px_18px_rgba(15,23,42,0.04)] hover:bg-surface-subtle",
                  )}
                  href={item.href}
                  key={item.key}
                >
                  {item.label}
                </Link>
              );
            }

            return (
              <div
                aria-disabled="true"
                className="inline-flex min-h-9 items-center justify-center rounded-full border border-border bg-surface-subtle px-3 text-[12px] font-semibold tracking-[-0.01em] text-muted"
                key={item.key}
              >
                {item.label}
              </div>
            );
          })}
        </div>
      </nav>
    );
  }

  return (
    <nav aria-label="Creator primary" className={className}>
      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        {getCreatorModeNavigationItems().map((item) => (
          <CreatorModeNavigationCard
            active={item.key === activeKey}
            description={item.description}
            href={item.href}
            key={item.key}
            label={item.label}
            navigationKey={item.key}
          />
        ))}
      </div>
    </nav>
  );
}

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
      ? "border-[#143d54]/10 bg-[linear-gradient(135deg,#123d55_0%,#216277_100%)] text-white shadow-[0_20px_50px_rgba(18,61,85,0.22)]"
      : available
        ? "border-white/72 bg-white/82 text-foreground shadow-[0_14px_32px_rgba(36,94,132,0.12)] hover:-translate-y-0.5 hover:bg-white"
        : "border-white/68 bg-white/58 text-muted shadow-[0_10px_24px_rgba(36,94,132,0.08)]",
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
              active ? "border-white/18 bg-white/12" : "border-[#0f566a]/12 bg-[#e9f7fb] text-[#0f566a]",
            )}
          >
            <Icon aria-hidden="true" className="size-4.5" strokeWidth={1.9} />
          </span>
          <span
            className={cn(
              "inline-flex min-h-7 items-center rounded-full px-3 text-[10px] font-semibold uppercase tracking-[0.18em]",
              active ? "bg-white/12 text-white/88" : "bg-[#e8f5fa] text-[#0f6172]",
            )}
          >
            live
          </span>
        </div>
        <div className="mt-4">
          <p className="text-[15px] font-semibold tracking-[-0.03em]">{label}</p>
          <p className={cn("mt-2 text-sm leading-6", active ? "text-white/76" : "text-muted")}>{description}</p>
        </div>
      </Link>
    );
  }

  return (
    <div aria-disabled="true" className={cardClassName}>
      <div className="flex items-start justify-between gap-3">
        <span className="inline-flex size-10 items-center justify-center rounded-full border border-black/6 bg-white/56 text-[#537182]">
          <Icon aria-hidden="true" className="size-4.5" strokeWidth={1.9} />
        </span>
        <span className="inline-flex min-h-7 items-center rounded-full bg-black/5 px-3 text-[10px] font-semibold uppercase tracking-[0.18em] text-muted">
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
                    "inline-flex min-h-9 items-center justify-center rounded-full px-2 text-[11px] font-semibold tracking-[-0.01em] transition",
                    active
                      ? "bg-[linear-gradient(135deg,#123d55_0%,#216277_100%)] text-white shadow-[0_10px_24px_rgba(18,61,85,0.18)]"
                      : "border border-[#d9ecf2] bg-white text-[#0f566a] shadow-[0_8px_18px_rgba(36,94,132,0.08)] hover:bg-[#f5fbfd]",
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
                className="inline-flex min-h-9 items-center justify-center rounded-full border border-[#e3edf1] bg-[#f7fafb] px-2 text-[11px] font-semibold tracking-[-0.01em] text-muted"
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

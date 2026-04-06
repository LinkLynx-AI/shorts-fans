import Link from "next/link";

import type { CreatorProfileStats, CreatorSummary } from "../model/creator";
import { getCreatorInitials } from "../model/creator";
import { cn } from "@/shared/lib";
import { Avatar, AvatarFallback, AvatarImage } from "@/shared/ui";

type CreatorAvatarProps = {
  className?: string;
  creator: CreatorSummary;
};

type CreatorIdentityProps = {
  className?: string;
  creator: CreatorSummary;
  href?: string;
};

type CreatorStatListProps = {
  className?: string;
  stats: CreatorProfileStats;
  variant?: "default" | "creatorProfile";
};

function formatCompactCount(value: number): string {
  return new Intl.NumberFormat("en", {
    maximumFractionDigits: 0,
    notation: "compact",
  }).format(value);
}

/**
 * creator 用の avatar asset を表示する。
 */
export function CreatorAvatar({ className, creator }: CreatorAvatarProps) {
  return (
    <Avatar className={cn("size-12 border-white/68", className)}>
      <AvatarImage alt={creator.displayName} src={creator.avatar.url} />
      <AvatarFallback>{getCreatorInitials(creator.displayName)}</AvatarFallback>
    </Avatar>
  );
}

/**
 * creator 名と handle をまとめて表示する。
 */
export function CreatorIdentity({ className, creator, href }: CreatorIdentityProps) {
  const content = (
    <div className={cn("min-w-0", className)}>
      <p className="truncate text-sm font-semibold text-current">{creator.displayName}</p>
      <p className="truncate text-[13px] text-current/72">{creator.handle}</p>
    </div>
  );

  if (!href) {
    return content;
  }

  return (
    <Link className="transition hover:opacity-90" href={href}>
      {content}
    </Link>
  );
}

/**
 * creator profile 用の stat list を表示する。
 */
export function CreatorStatList({
  className,
  stats,
  variant = "default",
}: CreatorStatListProps) {
  const items = [
    { label: "shorts", value: stats.shortCount.toString() },
    { label: "fans", value: formatCompactCount(stats.fanCount) },
    { label: "views", value: formatCompactCount(stats.viewCount) },
  ] as const;

  return (
    <div
      className={cn(
        variant === "creatorProfile"
          ? "grid grid-cols-3 gap-4 text-left"
          : "grid grid-cols-3 gap-3 text-center",
        className,
      )}
    >
      {items.map((stat) => (
        <div key={stat.label} className="min-w-0">
          <strong
            className={cn(
              "block font-display font-semibold tracking-[-0.04em] text-foreground",
              variant === "creatorProfile" ? "text-[18px]" : "text-lg",
            )}
          >
            {stat.value}
          </strong>
          <span
            className={cn(
              "mt-1 block text-[11px] text-muted",
              variant === "creatorProfile" ? "text-[12px] leading-[1.25]" : "uppercase tracking-[0.14em]",
            )}
          >
            {stat.label}
          </span>
        </div>
      ))}
    </div>
  );
}

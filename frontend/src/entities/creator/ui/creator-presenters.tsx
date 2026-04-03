import Link from "next/link";

import type { CreatorStat, CreatorSummary } from "../model/creator";
import { getCreatorInitials } from "../model/creator";
import { cn } from "@/shared/lib";
import { Avatar, AvatarFallback } from "@/shared/ui";

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
  stats: readonly CreatorStat[];
};

/**
 * creator 用の gradient avatar を表示する。
 */
export function CreatorAvatar({ className, creator }: CreatorAvatarProps) {
  return (
    <Avatar
      className={cn("size-12 border-white/68", className)}
      style={{
        background: `linear-gradient(180deg, ${creator.avatar.from} 0%, ${creator.avatar.accent} 44%, ${creator.avatar.to} 100%)`,
      }}
    >
      <AvatarFallback>{getCreatorInitials(creator.name)}</AvatarFallback>
    </Avatar>
  );
}

/**
 * creator 名と handle をまとめて表示する。
 */
export function CreatorIdentity({ className, creator, href }: CreatorIdentityProps) {
  const content = (
    <div className={cn("min-w-0", className)}>
      <p className="truncate text-sm font-semibold text-current">{creator.name}</p>
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
export function CreatorStatList({ className, stats }: CreatorStatListProps) {
  return (
    <div className={cn("grid grid-cols-3 gap-3 text-center", className)}>
      {stats.map((stat) => (
        <div key={stat.label} className="min-w-0">
          <strong className="block font-display text-lg font-semibold tracking-[-0.04em] text-foreground">
            {stat.value}
          </strong>
          <span className="mt-1 block text-[11px] uppercase tracking-[0.14em] text-muted">{stat.label}</span>
        </div>
      ))}
    </div>
  );
}

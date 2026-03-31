import Link from "next/link";

import type { CreatorProfilePreview } from "@/entities/creator";
import { Avatar, AvatarFallback, Button } from "@/shared/ui";

type SubscriptionCtaCardProps = {
  creator: CreatorProfilePreview;
};

/**
 * creator page で限定導線のたたき台を表示する。
 */
export function SubscriptionCtaCard({ creator }: SubscriptionCtaCardProps) {
  return (
    <section className="rounded-[1.75rem] border border-white/80 bg-[#201410] p-6 text-white shadow-[0_24px_80px_rgba(35,16,8,0.22)]">
      <div className="flex items-start gap-4">
        <Avatar className="size-14 rounded-[1.35rem]">
          <AvatarFallback>{creator.displayName.slice(0, 2)}</AvatarFallback>
        </Avatar>
        <div className="flex-1">
          <p className="text-xs font-semibold uppercase tracking-[0.24em] text-amber-200/80">subscription cta</p>
          <h2 className="mt-2 text-2xl font-semibold tracking-tight">{creator.displayName}</h2>
          <p className="mt-1 text-sm text-stone-300">{creator.handle}</p>
        </div>
        <div className="rounded-full border border-white/15 bg-white/10 px-4 py-2 text-sm font-semibold">
          {creator.monthlyPriceLabel}
        </div>
      </div>
      <div className="mt-6 grid gap-3 sm:grid-cols-3">
        {Array.from({ length: 3 }).map((_, index) => (
          <div
            key={index}
            className="h-28 rounded-[1.25rem] border border-white/10 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.16),transparent_42%),linear-gradient(135deg,rgba(255,255,255,0.08),rgba(255,255,255,0.02))] blur-[0.2px]"
          />
        ))}
      </div>
      <div className="mt-6 grid gap-2 text-sm text-stone-300">
        <p>限定縦動画: {creator.lockedPosts} 本</p>
        <p>公開 short: {creator.publicShorts} 本</p>
        <p>非購読時は blur thumbnail と CTA を優先表示する前提です。</p>
      </div>
      <div className="mt-6 flex flex-wrap gap-3">
        <Button asChild>
          <Link href="/subscriptions">購読導線の箱を確認</Link>
        </Button>
        <Button asChild variant="secondary">
          <Link href="/">公開 short に戻る</Link>
        </Button>
      </div>
    </section>
  );
}

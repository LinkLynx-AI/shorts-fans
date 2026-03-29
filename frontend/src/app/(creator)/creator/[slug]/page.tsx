import type { CreatorProfilePreview } from "@/entities/creator";
import { SubscriptionCtaCard } from "@/features/subscription-cta";
import { RouteStage } from "@/widgets/route-stage";

const creatorPreviewBySlug: Record<string, CreatorProfilePreview> = {
  "atelier-rin": {
    displayName: "Atelier Rin",
    handle: "@atelier-rin",
    lockedPosts: 18,
    monthlyPriceLabel: "¥1,480 / month",
    publicShorts: 26,
    slug: "atelier-rin",
    teaser: "公開 short で惹きつけて、限定縦動画で深く回遊させる creator page のたたき台。",
  },
};

type CreatorPageProps = {
  params: Promise<{ slug: string }>;
};

export default async function CreatorPage({ params }: CreatorPageProps) {
  const { slug } = await params;
  const creator = creatorPreviewBySlug[slug] ?? {
    displayName: slug
      .split("-")
      .map((segment) => `${segment.slice(0, 1).toUpperCase()}${segment.slice(1)}`)
      .join(" "),
    handle: `@${slug}`,
    lockedPosts: 12,
    monthlyPriceLabel: "¥1,280 / month",
    publicShorts: 14,
    slug,
    teaser: "slug だけ渡されても route shell を立ち上げられるよう、最小 preview を server 側で組み立てています。",
  };

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-6xl flex-col gap-6 px-6 py-8 sm:px-10">
      <RouteStage
        eyebrow="creator / profile"
        title={`${creator.displayName} の公開面と限定導線を切り分ける。`}
        description={creator.teaser}
        highlights={[
          `${creator.publicShorts} 本の公開 short を起点に発見を継続`,
          `${creator.lockedPosts} 件の限定動画を blur thumbnail 前提で表示`,
          `subscription は ${creator.monthlyPriceLabel} を起点に別 feature で接続`,
        ]}
        actions={[
          { href: "/shorts", label: "shorts に戻る" },
          { href: "/subscriptions", label: "購読面を見る", variant: "secondary" },
        ]}
      >
        <SubscriptionCtaCard creator={creator} />
      </RouteStage>
    </main>
  );
}

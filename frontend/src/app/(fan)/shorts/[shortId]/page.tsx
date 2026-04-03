import { getFeedShortForTab, getShortById, getShortThemeStyle } from "@/entities/short";
import { ShortPoster } from "@/entities/short";
import { UnlockCta } from "@/features/unlock-entry";
import { DetailShell } from "@/widgets/detail-shell";
import { RouteStructurePanel } from "@/widgets/route-structure-panel";

export default async function ShortDetailPage({
  params,
}: {
  params: Promise<{ shortId: string }>;
}) {
  const { shortId } = await params;
  const short = getShortById(shortId) ?? getFeedShortForTab("recommended");

  return (
    <DetailShell backHref="/" style={getShortThemeStyle(short)} variant="immersive">
      <h1 className="font-display text-[30px] font-semibold tracking-[-0.05em] text-white">Short detail structure</h1>
      <ShortPoster className="mx-auto w-full max-w-[274px]" meta={`shortId: ${shortId}`} short={short} title="detail viewport placeholder" variant="hero" />
      <UnlockCta short={short} />
      <RouteStructurePanel
        description="short detail route は feed からの遷移先だけ先に定義し、本文と paywall 連携は `SHO-5` / `SHO-8` で実装します。"
        items={[
          {
            description: "単体 short viewport をここに配置する",
            key: "viewport",
            label: "Detail viewport slot",
          },
          {
            description: "creator block と caption はここに差し込む",
            key: "creator",
            label: "Creator block slot",
          },
          {
            description: "unlock / continue main の分岐は feature として後続で接続する",
            key: "cta",
            label: "Unlock CTA contract",
          },
        ]}
        title="Short detail route blueprint"
      />
    </DetailShell>
  );
}

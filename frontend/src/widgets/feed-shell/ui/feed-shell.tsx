import { ShortPoster, getShortThemeStyle } from "@/entities/short";
import type { FeedTab, ShortPreviewMeta } from "@/entities/short";
import { UnlockCta } from "@/features/unlock-entry";
import { SegmentedControl, SurfacePanel } from "@/shared/ui";
import { RouteStructurePanel } from "@/widgets/route-structure-panel";

type FeedShellProps = {
  activeTab: FeedTab;
  short: ShortPreviewMeta;
};

/**
 * fan feed の route shell を表示する。
 */
export function FeedShell({ activeTab, short }: FeedShellProps) {
  return (
    <section className="relative min-h-full overflow-y-auto px-4 pb-28 pt-6 text-white" style={getShortThemeStyle(short)}>
      <div className="absolute inset-0 bg-[linear-gradient(180deg,var(--short-bg-start)_0%,var(--short-bg-mid)_54%,var(--short-bg-end)_100%)]" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.22),transparent_34%)]" />

      <div className="relative space-y-5">
        <SegmentedControl
          ariaLabel="Feed sections"
          className="mx-auto"
          items={[
            {
              active: activeTab === "recommended",
              href: "/?tab=recommended",
              key: "recommended",
              label: "おすすめ",
            },
            {
              active: activeTab === "following",
              href: "/?tab=following",
              key: "following",
              label: "フォロー中",
            },
          ]}
          variant="underline"
        />

        <h1 className="font-display text-[30px] font-semibold tracking-[-0.05em] text-white">Feed shell</h1>

        <ShortPoster className="mx-auto w-full max-w-[274px]" meta="short viewport" short={short} title="surface placeholder" variant="hero" />

        <SurfacePanel className="space-y-3 px-4 py-4 text-foreground">
          <p className="font-display text-[11px] font-semibold uppercase tracking-[0.22em] text-accent">Shared anchors</p>
          <UnlockCta href={`/shorts/${short.id}`} short={short} />
          <p className="text-sm leading-6 text-muted">
            `short viewport`、`creator block`、`unlock CTA` の配置だけを固定しています。表示内容の実装は `SHO-5` で埋めます。
          </p>
        </SurfacePanel>

        <RouteStructurePanel
          description="fan mode default の route と common shell をここで固定し、feed 本文は後続 issue で実装します。"
          items={[
            {
              description: "short の縦型 viewport をここに差し込む",
              key: "viewport",
              label: "Short viewport slot",
            },
            {
              description: "creator 名、caption、profile 遷移をここに配置する",
              key: "creator-block",
              label: "Creator block slot",
            },
            {
              description: "おすすめ / フォロー中の tab 切替だけ先に確定する",
              key: "tabs",
              label: "Top tab structure",
            },
          ]}
          title="Feed route blueprint"
        />
      </div>
    </section>
  );
}

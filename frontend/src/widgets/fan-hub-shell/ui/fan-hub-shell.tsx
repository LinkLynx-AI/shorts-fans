import type { FanCollectionTab } from "@/entities/short";
import { SegmentedControl } from "@/shared/ui";
import { RouteStructurePanel } from "@/widgets/route-structure-panel";

type FanHubShellProps = {
  activeTab: FanCollectionTab;
};

/**
 * private consumer hub の route shell を表示する。
 */
export function FanHubShell({ activeTab }: FanHubShellProps) {
  return (
    <section className="min-h-full overflow-y-auto px-4 pb-28 pt-4 text-foreground">
      <p className="font-display text-[11px] font-semibold uppercase tracking-[0.24em] text-accent">private hub</p>
      <h1 className="mt-2 font-display text-[32px] font-semibold tracking-[-0.05em]">Fan hub structure</h1>

      <div className="mt-6">
        <SegmentedControl
          ariaLabel="Archive sections"
          items={[
            {
              active: activeTab === "pinned",
              href: "/fan?tab=pinned",
              key: "pinned",
              label: "Pinned",
            },
            {
              active: activeTab === "library",
              href: "/fan?tab=library",
              key: "library",
              label: "Library",
            },
          ]}
        />
      </div>

      <RouteStructurePanel
        description="fan profile は private consumer hub として扱い、各セクションの本体は `SHO-7` で実装します。"
        items={[
          {
            description: "fan profile header と stats の配置だけを定義する",
            key: "header",
            label: "Header slot",
          },
          {
            description: `${activeTab === "library" ? "Library" : "Pinned"} panel の中身を後続 task で差し込む`,
            key: "panel",
            label: "Active panel slot",
          },
          {
            description: "Following / Settings への遷移導線をここに集約する",
            key: "supporting",
            label: "Supporting routes",
          },
        ]}
        title="Fan route blueprint"
      />
    </section>
  );
}

import { Separator, SurfacePanel } from "@/shared/ui";

export type RouteStructureItem = {
  description: string;
  key: string;
  label: string;
};

type RouteStructurePanelProps = {
  description: string;
  eyebrow?: string;
  items: readonly RouteStructureItem[];
  title: string;
};

/**
 * route 単位の構造と後続実装ポイントを表示する。
 */
export function RouteStructurePanel({
  description,
  eyebrow = "Route structure",
  items,
  title,
}: RouteStructurePanelProps) {
  return (
    <SurfacePanel className="px-4 py-4">
      <p className="font-display text-[11px] font-semibold uppercase tracking-[0.24em] text-accent">{eyebrow}</p>
      <h2 className="mt-2 font-display text-xl font-semibold tracking-[-0.04em] text-foreground">{title}</h2>
      <p className="mt-2 text-sm leading-6 text-muted">{description}</p>

      <div className="mt-4 space-y-3">
        {items.map((item, index) => (
          <div key={item.key}>
            {index > 0 ? <Separator className="mb-3" /> : null}
            <div className="space-y-1">
              <p className="text-sm font-semibold text-foreground">{item.label}</p>
              <p className="text-sm leading-6 text-muted">{item.description}</p>
            </div>
          </div>
        ))}
      </div>
    </SurfacePanel>
  );
}

import type { FeedTab } from "@/entities/short";
import { getFeedSurfaceByTab, ImmersiveShortSurface } from "@/widgets/immersive-short-surface";

type FeedShellProps = {
  activeTab: FeedTab;
};

/**
 * fan feed の route shell を表示する。
 */
export function FeedShell({ activeTab }: FeedShellProps) {
  const surface = getFeedSurfaceByTab(activeTab);

  return <ImmersiveShortSurface activeTab={activeTab} mode="feed" surface={surface} />;
}

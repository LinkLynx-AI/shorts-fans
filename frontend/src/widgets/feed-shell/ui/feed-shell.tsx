import { getCreatorById } from "@/entities/creator";
import type { FeedTab, ShortPreviewMeta } from "@/entities/short";
import { ImmersiveShortSurface } from "@/widgets/immersive-short-surface";

type FeedShellProps = {
  activeTab: FeedTab;
  short: ShortPreviewMeta;
};

/**
 * fan feed の route shell を表示する。
 */
export function FeedShell({ activeTab, short }: FeedShellProps) {
  const creator = getCreatorById(short.creatorId);

  if (!creator) {
    throw new Error(`Unknown creator for short: ${short.id}`);
  }

  return <ImmersiveShortSurface activeTab={activeTab} creator={creator} mode="feed" short={short} />;
}

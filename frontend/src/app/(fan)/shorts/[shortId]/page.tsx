import { getCreatorById } from "@/entities/creator";
import { getFeedShortForTab, getShortById } from "@/entities/short";
import { ImmersiveShortSurface } from "@/widgets/immersive-short-surface";

export default async function ShortDetailPage({
  params,
}: {
  params: Promise<{ shortId: string }>;
}) {
  const { shortId } = await params;
  const short = getShortById(shortId) ?? getFeedShortForTab("recommended");
  const creator = getCreatorById(short.creatorId);

  if (!creator) {
    throw new Error(`Unknown creator for short: ${short.id}`);
  }

  return <ImmersiveShortSurface backHref="/" creator={creator} mode="detail" short={short} />;
}

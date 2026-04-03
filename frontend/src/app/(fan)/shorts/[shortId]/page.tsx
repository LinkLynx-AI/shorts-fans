import { notFound } from "next/navigation";

import { getShortSurfaceById, ImmersiveShortSurface } from "@/widgets/immersive-short-surface";

export default async function ShortDetailPage({
  params,
}: {
  params: Promise<{ shortId: string }>;
}) {
  const { shortId } = await params;
  const surface = getShortSurfaceById(shortId);

  if (!surface) {
    notFound();
  }

  return <ImmersiveShortSurface backHref="/" mode="detail" surface={surface} />;
}

import { notFound } from "next/navigation";

import { getMainPlaybackSurfaceById, MainPlaybackSurface } from "@/widgets/main-playback-surface";

function normalizeFromShortId(value: string | string[] | undefined): string | undefined {
  if (Array.isArray(value)) {
    return value[0];
  }

  return value;
}

export default async function MainPlaybackPage({
  params,
  searchParams,
}: {
  params: Promise<{ mainId: string }>;
  searchParams: Promise<{ fromShortId?: string | string[] }>;
}) {
  const [{ mainId }, { fromShortId }] = await Promise.all([params, searchParams]);
  const normalizedFromShortId = normalizeFromShortId(fromShortId);
  const surface = getMainPlaybackSurfaceById(mainId, normalizedFromShortId);

  if (!surface) {
    notFound();
  }

  return (
    <MainPlaybackSurface
      fallbackHref={normalizedFromShortId ? `/shorts/${normalizedFromShortId}` : "/"}
      surface={surface}
    />
  );
}

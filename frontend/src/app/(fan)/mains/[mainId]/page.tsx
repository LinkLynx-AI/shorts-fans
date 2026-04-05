import { notFound } from "next/navigation";

import { getMainPlaybackSurfaceById, MainPlaybackGate } from "@/widgets/main-playback-surface";

function normalizeFromShortId(value: string | string[] | undefined): string | undefined {
  if (Array.isArray(value)) {
    return value[0];
  }

  return value;
}

function normalizeGrant(value: string | string[] | undefined): string | undefined {
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
  searchParams: Promise<{ fromShortId?: string | string[]; grant?: string | string[] }>;
}) {
  const [{ mainId }, { fromShortId, grant }] = await Promise.all([params, searchParams]);
  const normalizedFromShortId = normalizeFromShortId(fromShortId);
  const normalizedGrant = normalizeGrant(grant);

  if (!normalizedFromShortId || !normalizedGrant) {
    notFound();
  }

  const surface = getMainPlaybackSurfaceById(mainId, normalizedFromShortId);

  if (!surface) {
    notFound();
  }

  return (
    <MainPlaybackGate
      fallbackHref={normalizedFromShortId ? `/shorts/${normalizedFromShortId}` : "/"}
      grantToken={normalizedGrant}
      surface={surface}
    />
  );
}

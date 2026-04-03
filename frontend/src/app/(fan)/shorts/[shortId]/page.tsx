import { notFound } from "next/navigation";
import { z } from "zod";

import { resolveShortDetailBackHref } from "@/features/creator-navigation";
import { getEnumQueryParam, getSingleQueryParam } from "@/shared/lib";
import { getShortSurfaceById, ImmersiveShortSurface } from "@/widgets/immersive-short-surface";

const paramsSchema = z.object({
  shortId: z.string().min(1),
});

export default async function ShortDetailPage({
  params,
  searchParams,
}: {
  params: Promise<{ shortId: string }>;
  searchParams: Promise<{
    creatorId?: string | string[];
    from?: string | string[];
    profileFrom?: string | string[];
    profileQ?: string | string[];
    profileShortId?: string | string[];
    profileTab?: string | string[];
  }>;
}) {
  const rawParams = await params;
  const rawSearchParams = await searchParams;
  const { shortId } = paramsSchema.parse(rawParams);
  const routeState = {
    creatorId: getSingleQueryParam(rawSearchParams.creatorId),
    from: getEnumQueryParam(rawSearchParams.from, ["creator"]),
    profileFrom: getEnumQueryParam(rawSearchParams.profileFrom, ["feed", "search", "short"]),
    profileQ: getSingleQueryParam(rawSearchParams.profileQ),
    profileShortId: getSingleQueryParam(rawSearchParams.profileShortId),
    profileTab: getEnumQueryParam(rawSearchParams.profileTab, ["following", "recommended"]),
  };
  const surface = getShortSurfaceById(shortId);

  if (!surface) {
    notFound();
  }

  return <ImmersiveShortSurface backHref={resolveShortDetailBackHref(routeState)} mode="detail" surface={surface} />;
}

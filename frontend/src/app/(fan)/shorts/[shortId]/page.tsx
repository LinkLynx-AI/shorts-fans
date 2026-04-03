import { notFound } from "next/navigation";
import { z } from "zod";

import { resolveShortDetailBackHref } from "@/features/creator-navigation";
import { getSingleQueryParam } from "@/shared/lib";
import { getShortSurfaceById, ImmersiveShortSurface } from "@/widgets/immersive-short-surface";

const paramsSchema = z.object({
  shortId: z.string().min(1),
});

const searchParamsSchema = z.object({
  creatorId: z.string().optional(),
  from: z.enum(["creator"]).optional(),
  profileFrom: z.enum(["feed", "search", "short"]).optional(),
  profileQ: z.string().optional(),
  profileShortId: z.string().optional(),
  profileTab: z.enum(["following", "recommended"]).optional(),
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
  const routeState = searchParamsSchema.parse({
    creatorId: getSingleQueryParam(rawSearchParams.creatorId),
    from: getSingleQueryParam(rawSearchParams.from),
    profileFrom: getSingleQueryParam(rawSearchParams.profileFrom),
    profileQ: getSingleQueryParam(rawSearchParams.profileQ),
    profileShortId: getSingleQueryParam(rawSearchParams.profileShortId),
    profileTab: getSingleQueryParam(rawSearchParams.profileTab),
  });
  const surface = getShortSurfaceById(shortId);

  if (!surface) {
    notFound();
  }

  return <ImmersiveShortSurface backHref={resolveShortDetailBackHref(routeState)} mode="detail" surface={surface} />;
}

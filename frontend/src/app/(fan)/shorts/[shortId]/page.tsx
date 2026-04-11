import { cookies } from "next/headers";
import { notFound } from "next/navigation";
import { z } from "zod";

import { getPublicShortDetail } from "@/entities/short";
import { viewerSessionCookieName } from "@/entities/viewer";
import { resolveShortDetailBackHref } from "@/features/creator-navigation";
import { ApiError } from "@/shared/api";
import { getEnumQueryParam, getSingleQueryParam } from "@/shared/lib";
import {
  buildDetailSurfaceFromApi,
  getShortSurfaceById,
  ImmersiveShortSurface,
} from "@/widgets/immersive-short-surface";

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
  const legacySurface = !shortId.startsWith("short_") ? getShortSurfaceById(shortId) : undefined;

  if (legacySurface) {
    return <ImmersiveShortSurface backHref={resolveShortDetailBackHref(routeState)} mode="detail" surface={legacySurface} />;
  }

  const cookieStore = await cookies();
  const sessionToken = cookieStore?.get?.(viewerSessionCookieName)?.value;
  let detail: Awaited<ReturnType<typeof getPublicShortDetail>>;

  try {
    detail = await getPublicShortDetail({
      sessionToken,
      shortId,
    });
  } catch (error) {
    if (error instanceof ApiError && error.code === "http" && error.status === 404) {
      notFound();
    }

    throw error;
  }

  return (
    <ImmersiveShortSurface
      backHref={resolveShortDetailBackHref(routeState)}
      mode="detail"
      surface={buildDetailSurfaceFromApi(detail)}
    />
  );
}

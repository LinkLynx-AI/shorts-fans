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
    fanTab?: string | string[];
    from?: string | string[];
    profileFrom?: string | string[];
    profileQ?: string | string[];
    profileShortFanTab?: string | string[];
    profileShortId?: string | string[];
    profileTab?: string | string[];
  }>;
}) {
  const rawParams = await params;
  const rawSearchParams = await searchParams;
  const { shortId } = paramsSchema.parse(rawParams);
  const routeState = {
    creatorId: getSingleQueryParam(rawSearchParams.creatorId),
    fanTab: getEnumQueryParam(rawSearchParams.fanTab, ["library", "pinned"]),
    from: getEnumQueryParam(rawSearchParams.from, ["creator", "fan"]),
    profileFrom: getEnumQueryParam(rawSearchParams.profileFrom, ["feed", "search", "short"]),
    profileQ: getSingleQueryParam(rawSearchParams.profileQ),
    profileShortFanTab: getEnumQueryParam(rawSearchParams.profileShortFanTab, ["library", "pinned"]),
    profileShortId: getSingleQueryParam(rawSearchParams.profileShortId),
    profileTab: getEnumQueryParam(rawSearchParams.profileTab, ["following", "recommended"]),
  };
  const creatorProfileOrigin = {
    from: "short" as const,
    shortFanTab: routeState.from === "fan" ? routeState.fanTab : undefined,
    shortId,
  };
  const legacySurface = !shortId.startsWith("short_") ? getShortSurfaceById(shortId) : undefined;

  if (legacySurface) {
    return (
      <ImmersiveShortSurface
        backHref={resolveShortDetailBackHref(routeState)}
        creatorProfileOrigin={creatorProfileOrigin}
        mode="detail"
        surface={legacySurface}
      />
    );
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
      creatorProfileOrigin={creatorProfileOrigin}
      mode="detail"
      surface={buildDetailSurfaceFromApi(detail)}
    />
  );
}

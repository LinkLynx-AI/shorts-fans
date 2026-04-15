import { notFound, redirect } from "next/navigation";
import { z } from "zod";

import { getShortById } from "@/entities/short";
import { buildCreatorProfileHref } from "@/features/creator-navigation";
import { viewerSessionCookieName } from "@/entities/viewer";
import { buildFanLoginHref } from "@/features/fan-auth";
import { getFanAuthGateState } from "@/features/fan-auth-gate";
import { parseMockMainPlaybackGrantContext } from "@/features/unlock-entry";
import { ApiError } from "@/shared/api";
import { getEnumQueryParam } from "@/shared/lib";
import { readMockSignedToken } from "@/shared/lib/mock-signed-token";
import { cookies } from "next/headers";
import {
  LibraryMainReel,
  getMainPlaybackSurfaceById,
  loadLibraryMainReelState,
  MainPlaybackLockedState,
  MainPlaybackSurface,
  requestMainPlaybackSurface,
} from "@/widgets/main-playback-surface";

const mainPlaybackParamsSchema = z.object({
  mainId: z.string().min(1),
});

const firstSearchParamValueSchema = z
  .union([z.array(z.string().min(1)).nonempty(), z.string().min(1)])
  .transform<string>((value) => {
    if (!Array.isArray(value)) {
      return value;
    }

    const [firstValue] = value;

    if (!firstValue) {
      throw new Error("Missing search param value");
    }

    return firstValue;
  });

function parseOptionalSearchParam(value: unknown): string | undefined {
  const parsed = firstSearchParamValueSchema.safeParse(value);

  if (!parsed.success) {
    return undefined;
  }

  return parsed.data;
}

function parseMainPlaybackParams(value: unknown) {
  const parsed = mainPlaybackParamsSchema.safeParse(value);

  if (!parsed.success) {
    notFound();
  }

  return parsed.data;
}

export default async function MainPlaybackPage({
  params,
  searchParams,
}: {
  params: Promise<{ mainId: string }>;
  searchParams: Promise<{
    fanTab?: string | string[];
    from?: string | string[];
    fromShortId?: string | string[];
    grant?: string | string[];
  }>;
}) {
  const viewerState = await getFanAuthGateState();

  if (!viewerState.hasSession) {
    redirect(buildFanLoginHref());
  }

  const [rawParams, rawSearchParams] = await Promise.all([params, searchParams]);
  const { mainId } = parseMainPlaybackParams(rawParams);
  const normalizedFromShortId = parseOptionalSearchParam(rawSearchParams.fromShortId);
  const normalizedGrant = parseOptionalSearchParam(rawSearchParams.grant);
  const routeFrom = getEnumQueryParam(rawSearchParams.from, ["fan"]);
  const routeFanTab = getEnumQueryParam(rawSearchParams.fanTab, ["library", "pinned"]);

  if (routeFrom === "fan" && routeFanTab === "library" && normalizedFromShortId && !normalizedGrant) {
    const cookieStore = await cookies();
    const sessionToken = cookieStore.get(viewerSessionCookieName)?.value;
    const reelState = await loadLibraryMainReelState({
      mainId,
      sessionToken,
    });

    if (!reelState) {
      notFound();
    }

    return (
      <LibraryMainReel
        backHref="/fan?tab=library"
        initialIndex={reelState.initialIndex}
        items={reelState.items}
      />
    );
  }

  if (!normalizedFromShortId || !normalizedGrant) {
    notFound();
  }

  const fallbackHref = `/shorts/${normalizedFromShortId}`;
  const creatorProfileHref = (creatorId: string) =>
    buildCreatorProfileHref(
      creatorId,
      routeFrom === "fan"
        ? {
            from: "short",
            shortFanTab: routeFanTab,
            shortId: normalizedFromShortId,
          }
        : {
            from: "short",
            shortId: normalizedFromShortId,
          },
    );

  if (normalizedFromShortId.startsWith("short_")) {
    const cookieStore = await cookies();
    const sessionToken = cookieStore.get(viewerSessionCookieName)?.value;
    let surface: Awaited<ReturnType<typeof requestMainPlaybackSurface>>;

    try {
      surface = await requestMainPlaybackSurface({
        fromShortId: normalizedFromShortId,
        grant: normalizedGrant,
        mainId,
        sessionToken,
      });
    } catch (error) {
      if (error instanceof ApiError && error.code === "http") {
        if (error.status === 403) {
          return <MainPlaybackLockedState fallbackHref={fallbackHref} />;
        }

        if (error.status === 404) {
          notFound();
        }
      }

      throw error;
    }

    return (
      <MainPlaybackSurface
        creatorProfileHref={creatorProfileHref(surface.creator.id)}
        fallbackHref={fallbackHref}
        surface={surface}
      />
    );
  }

  const entryShort = getShortById(normalizedFromShortId);

  if (!entryShort || entryShort.canonicalMainId !== mainId) {
    notFound();
  }

  const grantPayload = readMockSignedToken(normalizedGrant);
  const parsedGrantContext = grantPayload
    ? parseMockMainPlaybackGrantContext(grantPayload.context)
    : null;

  if (
    !parsedGrantContext ||
    parsedGrantContext.mainId !== mainId ||
    parsedGrantContext.fromShortId !== normalizedFromShortId
  ) {
    return <MainPlaybackLockedState fallbackHref={fallbackHref} />;
  }

  const surface = getMainPlaybackSurfaceById(
    mainId,
    normalizedFromShortId,
    parsedGrantContext.grantKind,
  );

  if (!surface) {
    notFound();
  }

  return (
    <MainPlaybackSurface
      creatorProfileHref={creatorProfileHref(surface.creator.id)}
      fallbackHref={fallbackHref}
      surface={surface}
    />
  );
}

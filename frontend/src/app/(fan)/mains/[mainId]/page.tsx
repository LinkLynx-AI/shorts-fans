import { notFound, redirect } from "next/navigation";
import { z } from "zod";

import { getShortById } from "@/entities/short";
import { viewerSessionCookieName } from "@/entities/viewer";
import { buildFanLoginHref } from "@/features/fan-auth";
import { getFanAuthGateState } from "@/features/fan-auth-gate";
import { parseMockMainPlaybackGrantContext } from "@/features/unlock-entry";
import { ApiError } from "@/shared/api";
import { readMockSignedToken } from "@/shared/lib/mock-signed-token";
import { cookies } from "next/headers";
import {
  getMainPlaybackSurfaceById,
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

const mainPlaybackSearchParamsSchema = z.object({
  fromShortId: firstSearchParamValueSchema,
  grant: firstSearchParamValueSchema,
});

function parseMainPlaybackParams(value: unknown) {
  const parsed = mainPlaybackParamsSchema.safeParse(value);

  if (!parsed.success) {
    notFound();
  }

  return parsed.data;
}

function parseMainPlaybackSearchParams(value: unknown): {
  fromShortId: string;
  grant: string;
} {
  const parsed = mainPlaybackSearchParamsSchema.safeParse(value);

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
  searchParams: Promise<{ fromShortId?: string | string[]; grant?: string | string[] }>;
}) {
  const viewerState = await getFanAuthGateState();

  if (!viewerState.hasSession) {
    redirect(buildFanLoginHref());
  }

  const [rawParams, rawSearchParams] = await Promise.all([params, searchParams]);
  const { mainId } = parseMainPlaybackParams(rawParams);
  const { fromShortId: normalizedFromShortId, grant: normalizedGrant } =
    parseMainPlaybackSearchParams(rawSearchParams);

  const fallbackHref = `/shorts/${normalizedFromShortId}`;

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

    return <MainPlaybackSurface fallbackHref={fallbackHref} surface={surface} />;
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

  return <MainPlaybackSurface fallbackHref={fallbackHref} surface={surface} />;
}

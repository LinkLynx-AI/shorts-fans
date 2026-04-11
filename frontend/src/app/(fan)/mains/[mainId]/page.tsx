import { cookies } from "next/headers";
import { notFound, redirect } from "next/navigation";
import { z } from "zod";

import { viewerSessionCookieName } from "@/entities/viewer";
import { buildFanLoginHref } from "@/features/fan-auth";
import { getFanAuthGateState } from "@/features/fan-auth-gate";
import {
  loadMainPlaybackSurface,
  MainPlaybackLockedState,
  MainPlaybackSurface,
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
    return redirect(buildFanLoginHref());
  }

  const [rawParams, rawSearchParams] = await Promise.all([params, searchParams]);
  const { mainId } = parseMainPlaybackParams(rawParams);
  const { fromShortId: normalizedFromShortId, grant: normalizedGrant } = parseMainPlaybackSearchParams(rawSearchParams);

  const fallbackHref = `/shorts/${normalizedFromShortId}`;
  const cookieStore = await cookies();
  const sessionToken = cookieStore.get(viewerSessionCookieName)?.value;
  const result = await loadMainPlaybackSurface(mainId, {
    fromShortId: normalizedFromShortId,
    grant: normalizedGrant,
    sessionToken,
  });

  if (result.kind === "auth_required") {
    return redirect(buildFanLoginHref());
  }

  if (result.kind === "locked") {
    return <MainPlaybackLockedState fallbackHref={fallbackHref} />;
  }

  if (result.kind === "not_found") {
    return notFound();
  }

  return <MainPlaybackSurface fallbackHref={fallbackHref} surface={result.surface} />;
}

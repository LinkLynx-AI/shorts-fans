import { notFound } from "next/navigation";
import { z } from "zod";

import { getEnumQueryParam, getSingleQueryParam } from "@/shared/lib";
import { CreatorProfileShell, getCreatorProfileShellState } from "@/widgets/creator-profile-shell";

const paramsSchema = z.object({
  creatorId: z.string().min(1),
});

export default async function CreatorProfilePage({
  params,
  searchParams,
}: {
  params: Promise<{ creatorId: string }>;
  searchParams: Promise<{
    creatorDisplayName?: string | string[];
    creatorHandle?: string | string[];
    from?: string | string[];
    q?: string | string[];
    shortId?: string | string[];
    tab?: string | string[];
  }>;
}) {
  const rawParams = await params;
  const rawSearchParams = await searchParams;
  const { creatorId } = paramsSchema.parse(rawParams);
  const creatorHandle = getSingleQueryParam(rawSearchParams.creatorHandle);
  const routeState = {
    creatorDisplayName: getSingleQueryParam(rawSearchParams.creatorDisplayName),
    creatorHandle: creatorHandle?.startsWith("@")
      ? (creatorHandle as `@${string}`)
      : undefined,
    from: getEnumQueryParam(rawSearchParams.from, ["feed", "search", "short"]),
    q: getSingleQueryParam(rawSearchParams.q),
    shortId: getSingleQueryParam(rawSearchParams.shortId),
    tab: getEnumQueryParam(rawSearchParams.tab, ["following", "recommended"]),
  };
  const state = getCreatorProfileShellState(creatorId, {
    displayName: routeState.creatorDisplayName,
    handle: routeState.creatorHandle,
  });

  if (!state) {
    notFound();
  }

  return <CreatorProfileShell routeState={routeState} state={state} />;
}

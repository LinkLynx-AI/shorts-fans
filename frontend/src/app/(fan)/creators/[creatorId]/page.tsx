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
    from?: string | string[];
    q?: string | string[];
    shortId?: string | string[];
    tab?: string | string[];
  }>;
}) {
  const rawParams = await params;
  const rawSearchParams = await searchParams;
  const { creatorId } = paramsSchema.parse(rawParams);
  const routeState = {
    from: getEnumQueryParam(rawSearchParams.from, ["feed", "search", "short"]),
    q: getSingleQueryParam(rawSearchParams.q),
    shortId: getSingleQueryParam(rawSearchParams.shortId),
    tab: getEnumQueryParam(rawSearchParams.tab, ["following", "recommended"]),
  };
  const state = getCreatorProfileShellState(creatorId);

  if (!state) {
    notFound();
  }

  return <CreatorProfileShell routeState={routeState} state={state} />;
}

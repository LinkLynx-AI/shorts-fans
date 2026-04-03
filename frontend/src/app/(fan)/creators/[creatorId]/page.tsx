import { notFound } from "next/navigation";
import { z } from "zod";

import { getSingleQueryParam } from "@/shared/lib";
import { CreatorProfileShell, getCreatorProfileShellState } from "@/widgets/creator-profile-shell";

const paramsSchema = z.object({
  creatorId: z.string().min(1),
});

const searchParamsSchema = z.object({
  from: z.enum(["feed", "search", "short"]).optional(),
  q: z.string().optional(),
  shortId: z.string().optional(),
  tab: z.enum(["following", "recommended"]).optional(),
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
  const routeState = searchParamsSchema.parse({
    from: getSingleQueryParam(rawSearchParams.from),
    q: getSingleQueryParam(rawSearchParams.q),
    shortId: getSingleQueryParam(rawSearchParams.shortId),
    tab: getSingleQueryParam(rawSearchParams.tab),
  });
  const state = getCreatorProfileShellState(creatorId);

  if (!state) {
    notFound();
  }

  return <CreatorProfileShell routeState={routeState} state={state} />;
}

import { cookies } from "next/headers";
import { notFound } from "next/navigation";
import { z } from "zod";

import { viewerSessionCookieName } from "@/entities/viewer";
import { getEnumQueryParam, getSingleQueryParam } from "@/shared/lib";
import { CreatorProfileShell, loadCreatorProfileShellState } from "@/widgets/creator-profile-shell";

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
    shortFanTab?: string | string[];
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
    shortFanTab: getEnumQueryParam(rawSearchParams.shortFanTab, ["library", "pinned"]),
    shortId: getSingleQueryParam(rawSearchParams.shortId),
    tab: getEnumQueryParam(rawSearchParams.tab, ["following", "recommended"]),
  };
  const cookieStore = await cookies();
  const sessionToken = cookieStore.get(viewerSessionCookieName)?.value;
  const state = await loadCreatorProfileShellState(creatorId, {
    sessionToken,
  });

  if (!state) {
    notFound();
  }

  return (
    <CreatorProfileShell
      key={[
        creatorId,
        state.kind,
        state.viewer.isFollowing ? "following" : "not-following",
        state.stats.fanCount.toString(),
        sessionToken ? "session" : "guest",
      ].join(":")}
      routeState={routeState}
      state={state}
    />
  );
}

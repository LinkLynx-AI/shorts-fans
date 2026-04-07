import { redirect } from "next/navigation";

import { getFanHubState, normalizeFanHubTab } from "@/entities/fan-profile";
import { buildFanLoginHref } from "@/features/fan-auth";
import { getFanAuthGateState } from "@/features/fan-auth-gate";
import { FanHubShell } from "@/widgets/fan-hub-shell";

export default async function FanPage({
  searchParams,
}: {
  searchParams: Promise<{ tab?: string | string[] }>;
}) {
  const viewerState = await getFanAuthGateState();

  if (!viewerState.hasSession) {
    redirect(buildFanLoginHref());
  }

  const { tab } = await searchParams;

  return <FanHubShell state={getFanHubState(normalizeFanHubTab(tab))} />;
}

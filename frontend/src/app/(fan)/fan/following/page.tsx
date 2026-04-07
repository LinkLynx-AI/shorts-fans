import { redirect } from "next/navigation";

import { listFollowingItems } from "@/entities/fan-profile";
import { buildFanLoginHref } from "@/features/fan-auth";
import { getFanAuthGateState } from "@/features/fan-auth-gate";
import { FollowingShell } from "@/widgets/following-shell";

export default async function FollowingPage() {
  const viewerState = await getFanAuthGateState();

  if (!viewerState.hasSession) {
    redirect(buildFanLoginHref());
  }

  return <FollowingShell items={listFollowingItems()} />;
}

import { redirect } from "next/navigation";

import { buildFanLoginHref } from "@/features/fan-auth";
import { getFanAuthGateState } from "@/features/fan-auth-gate";
import { CreatorRegistrationSuccessPanel } from "@/features/creator-entry";

export default async function CreatorSuccessPage() {
  const viewerState = await getFanAuthGateState();
  const currentViewer = viewerState.currentViewer;

  if (!viewerState.hasSession || currentViewer === null) {
    redirect(buildFanLoginHref());
    return null;
  }

  if (!currentViewer.canAccessCreatorMode) {
    redirect("/fan/creator/register");
    return null;
  }

  if (currentViewer.activeMode === "creator") {
    redirect("/creator");
    return null;
  }

  return <CreatorRegistrationSuccessPanel />;
}

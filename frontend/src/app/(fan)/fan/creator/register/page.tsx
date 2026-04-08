import { redirect } from "next/navigation";

import { buildFanLoginHref } from "@/features/fan-auth";
import { getFanAuthGateState } from "@/features/fan-auth-gate";
import { CreatorRegistrationPanel } from "@/features/creator-entry";

export default async function CreatorRegisterPage() {
  const viewerState = await getFanAuthGateState();
  const currentViewer = viewerState.currentViewer;

  if (!viewerState.hasSession || currentViewer === null) {
    redirect(buildFanLoginHref());
    return null;
  }

  if (currentViewer.canAccessCreatorMode) {
    redirect(currentViewer.activeMode === "creator" ? "/creator" : "/fan/creator/success");
    return null;
  }

  return <CreatorRegistrationPanel />;
}

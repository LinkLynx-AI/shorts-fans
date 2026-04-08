import { redirect } from "next/navigation";

import { buildFanLoginHref } from "@/features/fan-auth";
import { getFanAuthGateState } from "@/features/fan-auth-gate";

export default async function CreatorPage() {
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

  if (currentViewer.activeMode !== "creator") {
    redirect("/fan/creator/success");
    return null;
  }

  return (
    <main className="mx-auto min-h-svh w-full max-w-[408px] bg-white">
      <h1 className="sr-only">Creator mode</h1>
    </main>
  );
}

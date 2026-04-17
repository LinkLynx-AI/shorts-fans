import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { buildFanLoginHref } from "@/features/fan-auth";
import { getFanAuthGateState } from "@/features/fan-auth-gate";
import {
  CreatorRegistrationSuccessPanel,
  fetchCreatorRegistration,
  getCreatorEntryErrorCode,
} from "@/features/creator-entry";
import { viewerSessionCookieName } from "@/entities/viewer";

export default async function CreatorSuccessPage() {
  const viewerState = await getFanAuthGateState();
  const currentViewer = viewerState.currentViewer;

  if (!viewerState.hasSession || currentViewer === null) {
    redirect(buildFanLoginHref());
    return null;
  }

  if (currentViewer.canAccessCreatorMode) {
    redirect(currentViewer.activeMode === "creator" ? "/creator" : "/fan");
    return null;
  }

  const sessionToken = (await cookies()).get(viewerSessionCookieName)?.value;
  if (!sessionToken) {
    redirect(buildFanLoginHref());
    return null;
  }

  let registration = null;
  let hasRegistrationState = false;
  try {
    registration = await fetchCreatorRegistration({ sessionToken });
    hasRegistrationState = true;
  } catch (error) {
    if (getCreatorEntryErrorCode(error) === "not_found") {
      redirect("/fan/settings/profile");
      return null;
    }
  }

  if (hasRegistrationState && registration?.state !== "submitted") {
    redirect("/fan/creator/register");
    return null;
  }

  return <CreatorRegistrationSuccessPanel />;
}

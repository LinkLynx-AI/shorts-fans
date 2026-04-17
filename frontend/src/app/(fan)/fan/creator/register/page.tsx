import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { viewerSessionCookieName } from "@/entities/viewer";
import { buildFanLoginHref } from "@/features/fan-auth";
import { getFanAuthGateState } from "@/features/fan-auth-gate";
import {
  CreatorRegistrationPanel,
  fetchCreatorRegistration,
  getCreatorEntryErrorCode,
} from "@/features/creator-entry";

export default async function CreatorRegisterPage() {
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
  try {
    registration = await fetchCreatorRegistration({ sessionToken });
  } catch (error) {
    if (getCreatorEntryErrorCode(error) === "not_found") {
      redirect("/fan/settings/profile");
      return null;
    }
  }

  if (registration?.actions.canEnterCreatorMode) {
    redirect(currentViewer.activeMode === "creator" ? "/creator" : "/fan");
    return null;
  }

  if (registration?.state === "submitted") {
    redirect("/fan/creator/success");
    return null;
  }

  return <CreatorRegistrationPanel initialRegistration={registration} />;
}

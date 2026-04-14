import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { viewerSessionCookieName } from "@/entities/viewer";
import { buildFanLoginHref } from "@/features/fan-auth";
import { getFanAuthGateState } from "@/features/fan-auth-gate";
import {
  getViewerProfile,
  ViewerProfileSettingsPanel,
} from "@/features/viewer-profile";

export default async function FanProfileSettingsPage() {
  const viewerState = await getFanAuthGateState();
  const currentViewer = viewerState.currentViewer;

  if (!viewerState.hasSession || currentViewer === null) {
    redirect(buildFanLoginHref());
    return null;
  }

  const cookieStore = await cookies();
  const sessionToken = cookieStore.get(viewerSessionCookieName)?.value;
  const profile = await getViewerProfile({
    ...(sessionToken ? { sessionToken } : {}),
  });

  return (
    <ViewerProfileSettingsPanel
      initialValues={{
        avatarUrl: profile.avatar?.url ?? null,
        displayName: profile.displayName,
        handle: profile.handle,
      }}
    />
  );
}

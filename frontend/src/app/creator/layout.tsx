import type { ReactNode } from "react";

import {
  CurrentViewerProvider,
  ViewerSessionProvider,
} from "@/entities/viewer";
import { getFanAuthGateState } from "@/features/fan-auth-gate";

function resolveViewerSessionKey(hasSession: boolean): string {
  return hasSession ? "session-present" : "session-missing";
}

function resolveCurrentViewerKey(
  currentViewer: Awaited<ReturnType<typeof getFanAuthGateState>>["currentViewer"],
): string {
  if (currentViewer === null) {
    return "anonymous";
  }

  return `${currentViewer.id}:${currentViewer.activeMode}:${currentViewer.canAccessCreatorMode}`;
}

export default async function CreatorLayout({ children }: { children: ReactNode }) {
  const viewerState = await getFanAuthGateState();

  return (
    <ViewerSessionProvider
      hasSession={viewerState.hasSession}
      key={resolveViewerSessionKey(viewerState.hasSession)}
    >
      <CurrentViewerProvider
        currentViewer={viewerState.currentViewer}
        key={resolveCurrentViewerKey(viewerState.currentViewer)}
      >
        {children}
      </CurrentViewerProvider>
    </ViewerSessionProvider>
  );
}

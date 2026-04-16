import type { ReactNode } from "react";

import {
  CurrentViewerProvider,
  ViewerSessionProvider,
} from "@/entities/viewer";
import { FanAuthDialogProvider } from "@/features/fan-auth";
import { getFanAuthGateState } from "@/features/fan-auth-gate";
import { FanBottomNavigation } from "@/features/fan-navigation";

const fanBottomNavigationOffsetClassName =
  "pb-[calc(76px+env(safe-area-inset-bottom,0px))]";

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

export default async function FanLayout({ children }: { children: ReactNode }) {
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
        <FanAuthDialogProvider>
          <div className="min-h-svh bg-background sm:px-4">
            <div className="mx-auto min-h-svh w-full max-w-[408px] overflow-hidden bg-white text-foreground sm:border sm:border-border sm:shadow-[var(--device-shadow)]">
              <div
                className={`relative min-h-svh overflow-hidden ${fanBottomNavigationOffsetClassName}`}
              >
                {children}
              </div>
            </div>
            <div className="fixed inset-x-0 bottom-0 z-30 sm:px-4">
              <div className="mx-auto w-full max-w-[408px]">
                <FanBottomNavigation />
              </div>
            </div>
          </div>
        </FanAuthDialogProvider>
      </CurrentViewerProvider>
    </ViewerSessionProvider>
  );
}

import type { ReactNode } from "react";

import {
  CurrentViewerProvider,
  getCurrentViewerBootstrap,
  viewerSessionCookieName,
} from "@/entities/viewer";
import { FanBottomNavigation } from "@/features/fan-navigation";
import { cookies } from "next/headers";

export default async function FanLayout({ children }: { children: ReactNode }) {
  const cookieStore = await cookies();
  const sessionToken = cookieStore.get(viewerSessionCookieName)?.value;
  let currentViewer = null;

  try {
    currentViewer = await getCurrentViewerBootstrap(
      sessionToken ? { sessionToken } : {},
    );
  } catch {
    currentViewer = null;
  }

  return (
    <CurrentViewerProvider currentViewer={currentViewer}>
      <div className="relative mx-auto min-h-svh w-full max-w-[408px] overflow-hidden bg-white text-foreground">
        <div className="relative min-h-svh overflow-hidden pb-[76px]">{children}</div>
        <div className="absolute inset-x-0 bottom-0 z-30">
          <FanBottomNavigation />
        </div>
      </div>
    </CurrentViewerProvider>
  );
}

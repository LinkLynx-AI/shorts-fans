import { cookies } from "next/headers";
import { cache } from "react";

import type { CurrentViewer } from "@/entities/viewer";
import {
  getCurrentViewerBootstrap,
  viewerSessionCookieName,
} from "@/entities/viewer";

export type FanAuthGateState = {
  currentViewer: CurrentViewer | null;
  hasSession: boolean;
};

/**
 * request cookie と viewer bootstrap から fan auth gate 用 state を解決する。
 */
export const getFanAuthGateState = cache(async (): Promise<FanAuthGateState> => {
  const cookieStore = await cookies();
  const sessionToken = cookieStore.get(viewerSessionCookieName)?.value;

  if (!sessionToken) {
    return {
      currentViewer: null,
      hasSession: false,
    };
  }

  try {
    return {
      currentViewer: await getCurrentViewerBootstrap({ sessionToken }),
      hasSession: true,
    };
  } catch {
    return {
      currentViewer: null,
      hasSession: true,
    };
  }
});

"use client";

import { createContext, useContext, type ReactNode } from "react";

const ViewerSessionContext = createContext(false);

type ViewerSessionProviderProps = {
  children: ReactNode;
  hasSession: boolean;
};

/**
 * request cookie に viewer session があるかを context として配下に渡す。
 */
export function ViewerSessionProvider({
  children,
  hasSession,
}: ViewerSessionProviderProps) {
  return (
    <ViewerSessionContext.Provider value={hasSession}>
      {children}
    </ViewerSessionContext.Provider>
  );
}

/**
 * 現在の request に viewer session cookie があるかを参照する。
 */
export function useHasViewerSession(): boolean {
  return useContext(ViewerSessionContext);
}

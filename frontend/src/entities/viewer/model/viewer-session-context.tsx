"use client";

import { createContext, useContext, type ReactNode } from "react";

const ViewerSessionContext = createContext(false);

type ViewerSessionProviderProps = {
  children: ReactNode;
  hasSession: boolean;
};

/**
 * protected fan flow を継続できる viewer session があるかを context として配下に渡す。
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
 * 現在の request で protected fan flow を継続できるかを参照する。
 */
export function useHasViewerSession(): boolean {
  return useContext(ViewerSessionContext);
}

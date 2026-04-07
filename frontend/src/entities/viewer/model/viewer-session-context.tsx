"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";

const ViewerSessionContext = createContext(false);
const ViewerSessionOverrideContext = createContext<((hasSession: boolean) => void) | null>(null);

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
  const [resolvedHasSession, setResolvedHasSession] = useState(hasSession);

  useEffect(() => {
    setResolvedHasSession(hasSession);
  }, [hasSession]);

  return (
    <ViewerSessionOverrideContext.Provider value={setResolvedHasSession}>
      <ViewerSessionContext.Provider value={resolvedHasSession}>
        {children}
      </ViewerSessionContext.Provider>
    </ViewerSessionOverrideContext.Provider>
  );
}

/**
 * 現在の request で protected fan flow を継続できるかを参照する。
 */
export function useHasViewerSession(): boolean {
  return useContext(ViewerSessionContext);
}

/**
 * 現在の request に対する viewer session 判定を client 側から同期する。
 */
export function useSetViewerSession(): (hasSession: boolean) => void {
  const setViewerSession = useContext(ViewerSessionOverrideContext);

  if (setViewerSession === null) {
    throw new Error("ViewerSessionProvider is required");
  }

  return setViewerSession;
}

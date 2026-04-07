"use client";

import {
  createContext,
  useContext,
  useState,
  type ReactNode,
} from "react";

import type { CurrentViewer } from "./current-viewer";

const CurrentViewerContext = createContext<CurrentViewer | null>(null);
const CurrentViewerOverrideContext = createContext<((viewer: CurrentViewer | null) => void) | null>(
  null,
);

type CurrentViewerProviderProps = {
  children: ReactNode;
  currentViewer: CurrentViewer | null;
};

/**
 * current viewer state を context として配下に渡す。
 */
export function CurrentViewerProvider({
  children,
  currentViewer,
}: CurrentViewerProviderProps) {
  const [resolvedCurrentViewer, setResolvedCurrentViewer] = useState(currentViewer);

  return (
    <CurrentViewerOverrideContext.Provider value={setResolvedCurrentViewer}>
      <CurrentViewerContext.Provider value={resolvedCurrentViewer}>
        {children}
      </CurrentViewerContext.Provider>
    </CurrentViewerOverrideContext.Provider>
  );
}

/**
 * app bootstrap で取得した current viewer state を参照する。
 */
export function useCurrentViewer(): CurrentViewer | null {
  return useContext(CurrentViewerContext);
}

/**
 * app bootstrap current viewer を client 側から同期する。
 */
export function useSetCurrentViewer(): (viewer: CurrentViewer | null) => void {
  const setCurrentViewer = useContext(CurrentViewerOverrideContext);

  if (setCurrentViewer === null) {
    throw new Error("CurrentViewerProvider is required");
  }

  return setCurrentViewer;
}

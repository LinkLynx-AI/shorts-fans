"use client";

import { createContext, useContext, type ReactNode } from "react";

import type { CurrentViewer } from "./current-viewer";

const CurrentViewerContext = createContext<CurrentViewer | null>(null);

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
  return (
    <CurrentViewerContext.Provider value={currentViewer}>
      {children}
    </CurrentViewerContext.Provider>
  );
}

/**
 * app bootstrap で取得した current viewer state を参照する。
 */
export function useCurrentViewer(): CurrentViewer | null {
  return useContext(CurrentViewerContext);
}

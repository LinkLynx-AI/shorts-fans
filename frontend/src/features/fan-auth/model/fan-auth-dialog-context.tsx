"use client";

import { useRouter } from "next/navigation";
import {
  createContext,
  startTransition,
  useContext,
  useState,
  type ReactNode,
} from "react";

import {
  getCurrentViewerBootstrap,
  useSetCurrentViewer,
  useSetViewerSession,
} from "@/entities/viewer";

import { FanAuthDialog } from "../ui/fan-auth-dialog";

type FanAuthDialogCloseBehavior = "back" | "stay";

type OpenFanAuthDialogOptions = {
  afterAuthenticatedHref?: string;
  closeBehavior?: FanAuthDialogCloseBehavior;
};

type FanAuthDialogContextValue = {
  closeFanAuthDialog: () => void;
  isFanAuthDialogOpen: boolean;
  openFanAuthDialog: (options?: OpenFanAuthDialogOptions) => void;
};

const FanAuthDialogContext = createContext<FanAuthDialogContextValue | null>(null);

/**
 * fan auth modal の共通 open / close state を fan layout 配下へ提供する。
 */
export function FanAuthDialogProvider({ children }: { children: ReactNode }) {
  const router = useRouter();
  const setCurrentViewer = useSetCurrentViewer();
  const setViewerSession = useSetViewerSession();
  const [afterAuthenticatedHref, setAfterAuthenticatedHref] = useState<string | null>(null);
  const [closeBehavior, setCloseBehavior] = useState<FanAuthDialogCloseBehavior>("stay");
  const [isFanAuthDialogOpen, setIsFanAuthDialogOpen] = useState(false);

  const resetFanAuthDialogState = () => {
    setIsFanAuthDialogOpen(false);
    setAfterAuthenticatedHref(null);
    setCloseBehavior("stay");
  };

  const closeFanAuthDialog = () => {
    const shouldReturnToPreviousRoute = closeBehavior === "back";

    resetFanAuthDialogState();

    if (!shouldReturnToPreviousRoute) {
      return;
    }

    startTransition(() => {
      router.back();
    });
  };

  const openFanAuthDialog = (options?: OpenFanAuthDialogOptions) => {
    setAfterAuthenticatedHref(options?.afterAuthenticatedHref ?? null);
    setCloseBehavior(options?.closeBehavior ?? "stay");
    setIsFanAuthDialogOpen(true);
  };

  const handleAuthenticated = async () => {
    const nextHref = afterAuthenticatedHref;
    const currentViewer = await getCurrentViewerBootstrap({
      credentials: "include",
    }).catch(() => null);

    if (currentViewer === null) {
      setCurrentViewer(null);
      setViewerSession(false);

      return "認証自体は完了しましたが、状態反映の確認に失敗しました。画面を更新して確認してください。";
    }

    setCurrentViewer(currentViewer);
    setViewerSession(currentViewer !== null);
    resetFanAuthDialogState();

    startTransition(() => {
      if (nextHref) {
        router.push(nextHref);
        return;
      }

      router.refresh();
    });

    return null;
  };

  return (
    <FanAuthDialogContext.Provider
      value={{
        closeFanAuthDialog,
        isFanAuthDialogOpen,
        openFanAuthDialog,
      }}
    >
      {children}
      <FanAuthDialog
        onAuthenticated={handleAuthenticated}
        onOpenChange={(open) => {
          if (open) {
            setIsFanAuthDialogOpen(true);
            return;
          }

          closeFanAuthDialog();
        }}
        open={isFanAuthDialogOpen}
      />
    </FanAuthDialogContext.Provider>
  );
}

/**
 * fan auth modal を任意の surface から開閉する。
 */
export function useFanAuthDialog(): FanAuthDialogContextValue {
  const context = useContext(FanAuthDialogContext);

  if (context === null) {
    throw new Error("FanAuthDialogProvider is required");
  }

  return context;
}

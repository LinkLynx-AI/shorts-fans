"use client";

import { useRouter } from "next/navigation";
import {
  createContext,
  startTransition,
  useCallback,
  useContext,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";

import {
  type CurrentViewer,
  getCurrentViewerBootstrap,
  useCurrentViewer,
  useHasViewerSession,
  useSetCurrentViewer,
  useSetViewerSession,
} from "@/entities/viewer";

import { type FanAuthMode } from "./fan-auth";
import { FanAuthDialog } from "../ui/fan-auth-dialog";

type FanAuthDialogCloseBehavior = "back" | "stay";
type FanAuthDialogPostAuthNavigation = "none" | "refresh";

type OpenFanAuthDialogOptions = {
  afterAuthenticatedHref?: string;
  allowClose?: boolean;
  closeBehavior?: FanAuthDialogCloseBehavior;
  closeFallbackHref?: string;
  initialMode?: FanAuthMode;
  onAfterAuthenticated?: ((currentViewer?: CurrentViewer | null) => Promise<void> | void) | undefined;
  postAuthNavigation?: FanAuthDialogPostAuthNavigation;
};

type FanAuthDialogControlsValue = {
  closeFanAuthDialog: () => void;
  openFanAuthDialog: (options?: OpenFanAuthDialogOptions) => void;
};

type FanAuthDialogContextValue = FanAuthDialogControlsValue & {
  isFanAuthDialogOpen: boolean;
};

type HandleAuthenticatedOptions = {
  afterViewerSynced?: () => Promise<string | null> | string | null;
  authenticatedMode?: FanAuthMode;
};

const FanAuthDialogControlsContext = createContext<FanAuthDialogControlsValue | null>(null);
const FanAuthDialogStateContext = createContext<{ isFanAuthDialogOpen: boolean } | null>(null);

function getPostAuthRecoveryErrorMessage() {
  return "認証は完了しましたが、元の画面の復元に失敗しました。もう一度お試しください。";
}

/**
 * fan auth modal の共通 open / close state を fan layout 配下へ提供する。
 */
export function FanAuthDialogProvider({ children }: { children: ReactNode }) {
  const router = useRouter();
  const currentViewer = useCurrentViewer();
  const hasViewerSession = useHasViewerSession();
  const setCurrentViewer = useSetCurrentViewer();
  const setViewerSession = useSetViewerSession();
  const [afterAuthenticatedHref, setAfterAuthenticatedHref] = useState<string | null>(null);
  const [allowClose, setAllowClose] = useState(true);
  const [authSessionKey, setAuthSessionKey] = useState(0);
  const [initialMode, setInitialMode] = useState<FanAuthMode>("sign-in");
  const [isFanAuthDialogOpen, setIsFanAuthDialogOpen] = useState(false);
  const [onAfterAuthenticated, setOnAfterAuthenticated] = useState<
    ((currentViewer?: CurrentViewer | null) => Promise<void> | void) | null
  >(null);
  const [postAuthNavigation, setPostAuthNavigation] = useState<FanAuthDialogPostAuthNavigation>("refresh");
  const allowCloseRef = useRef(true);
  const closeBehaviorRef = useRef<FanAuthDialogCloseBehavior>("stay");
  const closeFallbackHrefRef = useRef<string | null>(null);

  const resetFanAuthDialogState = () => {
    allowCloseRef.current = true;
    closeBehaviorRef.current = "stay";
    closeFallbackHrefRef.current = null;
    setIsFanAuthDialogOpen(false);
    setAfterAuthenticatedHref(null);
    setAllowClose(true);
    setInitialMode("sign-in");
    setOnAfterAuthenticated(null);
    setPostAuthNavigation("refresh");
  };

  const closeFanAuthDialog = useCallback(() => {
    if (!allowCloseRef.current) {
      return;
    }

    const shouldReturnToPreviousRoute = closeBehaviorRef.current === "back";
    const fallbackHref = closeFallbackHrefRef.current;

    resetFanAuthDialogState();

    if (!shouldReturnToPreviousRoute) {
      return;
    }

    startTransition(() => {
      if (window.history.length <= 1 && fallbackHref) {
        router.push(fallbackHref);
        return;
      }

      router.back();
    });
  }, [router]);

  const openFanAuthDialog = useCallback((options?: OpenFanAuthDialogOptions) => {
    const nextAllowClose = options?.allowClose ?? true;
    const nextCloseBehavior = options?.closeBehavior ?? "stay";
    const nextCloseFallbackHref = options?.closeFallbackHref ?? null;

    allowCloseRef.current = nextAllowClose;
    closeBehaviorRef.current = nextCloseBehavior;
    closeFallbackHrefRef.current = nextCloseFallbackHref;
    setAfterAuthenticatedHref(options?.afterAuthenticatedHref ?? null);
    setAllowClose(nextAllowClose);
    setInitialMode(options?.initialMode ?? "sign-in");
    setOnAfterAuthenticated(() => options?.onAfterAuthenticated ?? null);
    setPostAuthNavigation(
      options?.postAuthNavigation ?? (options?.initialMode === "re-auth" ? "none" : "refresh"),
    );
    setAuthSessionKey((currentKey) => currentKey + 1);
    setIsFanAuthDialogOpen(true);
  }, []);

  const handleAuthenticated = async (
    options?: HandleAuthenticatedOptions,
  ) => {
    const nextHref = afterAuthenticatedHref;
    const isReAuthFlow =
      options?.authenticatedMode === "re-auth" ||
      (options?.authenticatedMode === undefined && initialMode === "re-auth");
    const shouldRefresh = !isReAuthFlow && postAuthNavigation === "refresh";
    const postAuthAction = onAfterAuthenticated;

    if (isReAuthFlow) {
      try {
        await postAuthAction?.(currentViewer);
      } catch {
        return getPostAuthRecoveryErrorMessage();
      }

      resetFanAuthDialogState();
      return null;
    }

    const authenticatedViewer = await getCurrentViewerBootstrap({
      credentials: "include",
    }).catch(() => null);

    if (authenticatedViewer === null) {
      if (hasViewerSession) {
        setCurrentViewer(null);
        setViewerSession(false);
      }

      return "認証自体は完了しましたが、状態反映の確認に失敗しました。画面を更新して確認してください。";
    }

    setCurrentViewer(authenticatedViewer);
    setViewerSession(true);

    const postSyncErrorMessage = await options?.afterViewerSynced?.();

    if (postSyncErrorMessage) {
      return postSyncErrorMessage;
    }

    try {
      await postAuthAction?.(authenticatedViewer);
    } catch {
      return getPostAuthRecoveryErrorMessage();
    }

    resetFanAuthDialogState();

    startTransition(() => {
      if (nextHref) {
        router.push(nextHref);
        return;
      }

      if (shouldRefresh) {
        router.refresh();
      }
    });

    return null;
  };

  const controlsValue = useMemo<FanAuthDialogControlsValue>(
    () => ({
      closeFanAuthDialog,
      openFanAuthDialog,
    }),
    [closeFanAuthDialog, openFanAuthDialog],
  );
  const stateValue = useMemo(
    () => ({
      isFanAuthDialogOpen,
    }),
    [isFanAuthDialogOpen],
  );

  return (
    <FanAuthDialogControlsContext.Provider value={controlsValue}>
      <FanAuthDialogStateContext.Provider value={stateValue}>
        {children}
        <FanAuthDialog
          allowClose={allowClose}
          initialMode={initialMode}
          onAuthenticated={handleAuthenticated}
          onFallbackToSignIn={() => {
            allowCloseRef.current = true;
            setAllowClose(true);
          }}
          onOpenChange={(open) => {
            if (open) {
              setIsFanAuthDialogOpen(true);
              return;
            }

            closeFanAuthDialog();
          }}
          open={isFanAuthDialogOpen}
          sessionKey={authSessionKey}
        />
      </FanAuthDialogStateContext.Provider>
    </FanAuthDialogControlsContext.Provider>
  );
}

/**
 * fan auth modal の open / close controls だけを取得する。
 */
export function useFanAuthDialogControls(): FanAuthDialogControlsValue {
  const controls = useContext(FanAuthDialogControlsContext);

  if (controls === null) {
    throw new Error("FanAuthDialogProvider is required");
  }

  return controls;
}

/**
 * fan auth modal を任意の surface から開閉する。
 */
export function useFanAuthDialog(): FanAuthDialogContextValue {
  const controls = useFanAuthDialogControls();
  const state = useContext(FanAuthDialogStateContext);

  if (state === null) {
    throw new Error("FanAuthDialogProvider is required");
  }

  return {
    ...controls,
    ...state,
  };
}

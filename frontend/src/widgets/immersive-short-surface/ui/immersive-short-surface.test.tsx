import userEvent from "@testing-library/user-event";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

import {
  CreatorFollowApiError,
  updateCreatorFollow,
} from "@/entities/creator";
import { CurrentViewerProvider, ViewerSessionProvider } from "@/entities/viewer";
import {
  useFanAuthDialog,
  useFanAuthDialogControls,
} from "@/features/fan-auth";
import {
  normalizeUnlockSurface,
  requestCardSetupSession,
  requestCardSetupToken,
  requestMainAccessEntry,
  requestMainPurchase,
  requestUnlockSurfaceByShortId,
  type UnlockSurfaceModel,
} from "@/features/unlock-entry";
import { ApiError } from "@/shared/api";
import {
  buildDetailSurfaceFromApi,
  buildFeedSurfaceFromApiItem,
  getFeedSurfaceByTab,
  getShortSurfaceById,
} from "@/widgets/immersive-short-surface";

import { ImmersiveShortSurface } from "./immersive-short-surface";

const push = vi.fn();

vi.mock("@/entities/creator", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/entities/creator")>();

  return {
    ...actual,
    updateCreatorFollow: vi.fn(),
  };
});

vi.mock("@/features/fan-auth", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/features/fan-auth")>();

  return {
    ...actual,
    useFanAuthDialogControls: vi.fn(),
    useFanAuthDialog: vi.fn(),
  };
});

vi.mock("@/features/unlock-entry", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/features/unlock-entry")>();
  const mockedRequestCardSetupSession = vi.fn(actual.requestCardSetupSession);
  const mockedRequestCardSetupToken = vi.fn(actual.requestCardSetupToken);
  const mockedRequestMainAccessEntry = vi.fn(actual.requestMainAccessEntry);
  const mockedRequestMainPurchase = vi.fn(actual.requestMainPurchase);
  const mockedRequestUnlockSurfaceByShortId = vi.fn(actual.requestUnlockSurfaceByShortId);
  const buildMockPaywallTitle = (caption: string) => {
    const normalizedCaption = caption.trim().replace(/[。.!?]+$/u, "");

    return normalizedCaption ? `${normalizedCaption} の続きを見る` : "この short の続きを見る";
  };
  const buildLegacyUnlockLabel = (unlock: UnlockSurfaceModel) => {
    switch (unlock.unlockCta.state) {
      case "continue_main":
        return "Continue main";
      case "owner_preview":
        return "Owner preview";
      case "setup_required":
      case "unlock_available": {
        const priceLabel = `¥${unlock.main.priceJpy.toLocaleString("ja-JP")}`;
        const minutesLabel = `${Math.max(1, Math.round(unlock.main.durationSeconds / 60))}分`;

        return `Unlock ${priceLabel} | ${minutesLabel}`;
      }
      case "unavailable":
        return "Unlock";
    }
  };
  const buildPurchaseFlowLabel = (
    selection: {
      mode: "new_card" | "saved_card";
      paymentMethodId?: string;
    },
    unlock: UnlockSurfaceModel,
  ) => {
    switch (unlock.purchase.state) {
      case "already_purchased":
        return "Continue main";
      case "owner_preview":
        return "Owner preview";
      case "purchase_ready":
      case "setup_required":
        return selection.mode === "saved_card" ? `Purchase ¥${unlock.main.priceJpy.toLocaleString("ja-JP")}` : null;
      default:
        return null;
    }
  };

  return {
    ...actual,
    UnlockPaywallDialog: ({
      acceptAge,
      acceptTerms,
      cardSetupSession,
      isSubmitting = false,
      onAcceptAgeChange,
      onAcceptTermsChange,
      onCardPaymentTokenCreated,
      onClose,
      onConfirm,
      onPaymentSelectionChange,
      open,
      selection,
      unlock,
      usePurchaseFlow = true,
    }: {
      acceptAge: boolean;
      acceptTerms: boolean;
      cardSetupSession?: object | null;
      isSubmitting?: boolean;
      onAcceptAgeChange: (checked: boolean) => void;
      onAcceptTermsChange: (checked: boolean) => void;
      onCardPaymentTokenCreated: (paymentTokenId: string) => void;
      onClose: () => void;
      onConfirm: () => void;
      onPaymentSelectionChange: (selection: {
        mode: "new_card" | "saved_card";
        paymentMethodId?: string;
      }) => void;
      open: boolean;
      selection: {
        mode: "new_card" | "saved_card";
        paymentMethodId?: string;
      };
      unlock: UnlockSurfaceModel;
      usePurchaseFlow?: boolean;
    }) => {
      if (!open) {
        return null;
      }

      const supportsSelection =
        usePurchaseFlow &&
        (unlock.purchase.state === "purchase_ready" || unlock.purchase.state === "setup_required");
      const consentSatisfied =
        (!unlock.setup.requiresAgeConfirmation || acceptAge) &&
        (!unlock.setup.requiresTermsAcceptance || acceptTerms);
      const confirmEnabled =
        consentSatisfied && !isSubmitting;
      const primaryActionLabel = usePurchaseFlow
        ? buildPurchaseFlowLabel(selection, unlock)
        : buildLegacyUnlockLabel(unlock);

      return (
        <div aria-label={buildMockPaywallTitle(unlock.short.caption)} role="dialog">
          {unlock.setup.requiresAgeConfirmation ? (
            <label>
              <input
                checked={acceptAge}
                onChange={(event) => {
                  onAcceptAgeChange(event.target.checked);
                }}
                type="checkbox"
              />
              <span>18歳以上であり、年齢確認に同意する</span>
            </label>
          ) : null}
          {unlock.setup.requiresTermsAcceptance ? (
            <label>
              <input
                checked={acceptTerms}
                onChange={(event) => {
                  onAcceptTermsChange(event.target.checked);
                }}
                type="checkbox"
              />
              <span>利用規約とポリシーに同意し、main purchase へ進む</span>
            </label>
          ) : null}
          {supportsSelection && unlock.purchase.savedPaymentMethods.length > 0 ? (
            unlock.purchase.savedPaymentMethods.map((method) => (
              <label key={method.paymentMethodId}>
                <input
                  checked={selection.mode === "saved_card" && selection.paymentMethodId === method.paymentMethodId}
                  name="paywall-payment-method"
                  onChange={() => {
                    onPaymentSelectionChange({
                      mode: "saved_card",
                      paymentMethodId: method.paymentMethodId,
                    });
                  }}
                  type="radio"
                />
                <span>{`Saved card ${method.last4}`}</span>
              </label>
            ))
          ) : null}
          {supportsSelection ? (
            <label>
              <input
                checked={selection.mode === "new_card"}
                name="paywall-payment-method"
                onChange={() => {
                  onPaymentSelectionChange({
                    mode: "new_card",
                  });
                }}
                type="radio"
              />
              <span>新しい card を使う</span>
            </label>
          ) : null}
          {selection.mode === "new_card" && !consentSatisfied ? (
            <span>年齢確認と利用規約への同意を完了すると card widget を表示します。</span>
          ) : null}
          {selection.mode === "new_card" && cardSetupSession && consentSatisfied ? (
            <button
              onClick={() => {
                onCardPaymentTokenCreated("widget-payment-token");
              }}
              type="button"
            >
              Mock card widget
            </button>
          ) : null}
          {primaryActionLabel ? (
            <button disabled={!confirmEnabled} onClick={onConfirm} type="button">
              {primaryActionLabel}
            </button>
          ) : null}
          {isSubmitting ? <span>Submitting purchase</span> : null}
          <button onClick={onClose} type="button">
            閉じる
          </button>
        </div>
      );
    },
    requestCardSetupSession: mockedRequestCardSetupSession,
    requestCardSetupToken: mockedRequestCardSetupToken,
    requestMainAccessEntry: mockedRequestMainAccessEntry,
    requestMainPurchase: mockedRequestMainPurchase,
    requestUnlockSurfaceByShortId: mockedRequestUnlockSurfaceByShortId,
    useUnlockPaywallController: (options: Parameters<typeof actual.useUnlockPaywallController>[0]) =>
      actual.useUnlockPaywallController({
        ...options,
        deps: {
          requestCardSetupSession: mockedRequestCardSetupSession,
          requestCardSetupToken: mockedRequestCardSetupToken,
          requestMainAccessEntry: mockedRequestMainAccessEntry,
          requestMainPurchase: mockedRequestMainPurchase,
          requestUnlockSurfaceByShortId: mockedRequestUnlockSurfaceByShortId,
        },
      }),
  };
});

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push,
  }),
}));

const mockedUpdateCreatorFollow = vi.mocked(updateCreatorFollow);
const mockedUseFanAuthDialogControls = vi.mocked(useFanAuthDialogControls);
const mockedUseFanAuthDialog = vi.mocked(useFanAuthDialog);
const mockedRequestCardSetupSession = vi.mocked(requestCardSetupSession);
const mockedRequestCardSetupToken = vi.mocked(requestCardSetupToken);
const mockedRequestMainAccessEntry = vi.mocked(requestMainAccessEntry);
const mockedRequestMainPurchase = vi.mocked(requestMainPurchase);
const mockedRequestUnlockSurfaceByShortId = vi.mocked(requestUnlockSurfaceByShortId);
const openFanAuthDialog = vi.fn();

function renderWithViewerSession(
  ui: React.ReactElement,
  {
    currentViewer = null,
    hasSession,
  }: {
    currentViewer?: {
      activeMode: "creator" | "fan";
      canAccessCreatorMode: boolean;
      id: string;
    } | null;
    hasSession: boolean;
  },
) {
  return render(
    <ViewerSessionProvider hasSession={hasSession}>
      <CurrentViewerProvider currentViewer={currentViewer}>
        {ui}
      </CurrentViewerProvider>
    </ViewerSessionProvider>,
  );
}

function createApiFeedSurface(state: "continue_main" | "owner_preview" | "setup_required" | "unlock_available") {
  return buildFeedSurfaceFromApiItem({
    creator: {
      avatar: null,
      bio: "night preview specialist",
      displayName: "Mina Rei",
      handle: "@minarei",
      id: "creator_mina_rei",
    },
    short: {
      caption: "quiet rooftop preview",
      canonicalMainId: "main_mina_quiet_rooftop",
      creatorId: "creator_mina_rei",
      id: "short_mina_rooftop",
      media: {
        durationSeconds: 16,
        id: "asset_short_mina_rooftop",
        kind: "video",
        posterUrl: "https://cdn.example.com/shorts/poster.jpg",
        url: "https://cdn.example.com/shorts/playback.mp4",
      },
      previewDurationSeconds: 16,
    },
    unlockCta: {
      mainDurationSeconds: 480,
      priceJpy: 1800,
      resumePositionSeconds: state === "continue_main" ? 120 : null,
      state,
    },
    viewer: {
      isFollowingCreator: false,
      isPinned: true,
    },
  });
}

function createApiDetailSurface(state: "continue_main" | "owner_preview" | "setup_required" | "unlock_available") {
  return buildDetailSurfaceFromApi({
    creator: {
      avatar: null,
      bio: "night preview specialist",
      displayName: "Mina Rei",
      handle: "@minarei",
      id: "creator_mina_rei",
    },
    short: {
      caption: "quiet rooftop preview",
      canonicalMainId: "main_mina_quiet_rooftop",
      creatorId: "creator_mina_rei",
      id: "short_mina_rooftop",
      media: {
        durationSeconds: 16,
        id: "asset_short_mina_rooftop",
        kind: "video",
        posterUrl: "https://cdn.example.com/shorts/poster.jpg",
        url: "https://cdn.example.com/shorts/playback.mp4",
      },
      previewDurationSeconds: 16,
    },
    unlockCta: {
      mainDurationSeconds: 480,
      priceJpy: 1800,
      resumePositionSeconds: state === "continue_main" ? 120 : null,
      state,
    },
    viewer: {
      isFollowingCreator: false,
      isPinned: true,
    },
  });
}

function createResolvedApiUnlock({
  accessReason = "unlock_required",
  accessStatus = "locked",
  entryToken = "resolved-entry-token",
  purchaseState,
  requiresAgeConfirmation = false,
  requiresCardSetup = false,
  requiresTermsAcceptance = false,
  savedPaymentMethods = [],
  unlockCtaState,
}: {
  accessReason?: UnlockSurfaceModel["access"]["reason"];
  accessStatus?: UnlockSurfaceModel["access"]["status"];
  entryToken?: string;
  purchaseState: UnlockSurfaceModel["purchase"]["state"];
  requiresAgeConfirmation?: boolean;
  requiresCardSetup?: boolean;
  requiresTermsAcceptance?: boolean;
  savedPaymentMethods?: UnlockSurfaceModel["purchase"]["savedPaymentMethods"];
  unlockCtaState: UnlockSurfaceModel["unlockCta"]["state"];
}): UnlockSurfaceModel {
  const baseUnlock = createApiFeedSurface("setup_required").unlock;

  return normalizeUnlockSurface({
    access: {
      mainId: baseUnlock.main.id,
      reason: accessReason,
      status: accessStatus,
    },
    creator: baseUnlock.creator,
    entryContext: {
      accessEntryPath: baseUnlock.entryContext.accessEntryPath,
      purchasePath: baseUnlock.entryContext.purchasePath,
      token: entryToken,
    },
    main: baseUnlock.main,
    purchase: {
      pendingReason: purchaseState === "purchase_pending" ? "provider_processing" : null,
      savedPaymentMethods,
      setup: {
        required: requiresAgeConfirmation || requiresCardSetup || requiresTermsAcceptance,
        requiresAgeConfirmation,
        requiresCardSetup,
        requiresTermsAcceptance,
      },
      state: purchaseState,
      supportedCardBrands: ["visa", "mastercard", "jcb", "american_express"],
    },
    short: baseUnlock.short,
    unlockCta: {
      mainDurationSeconds:
        unlockCtaState === "continue_main" || unlockCtaState === "owner_preview" ? null : baseUnlock.main.durationSeconds,
      priceJpy:
        unlockCtaState === "continue_main" || unlockCtaState === "owner_preview" ? null : baseUnlock.main.priceJpy,
      resumePositionSeconds: unlockCtaState === "continue_main" ? 120 : null,
      state: unlockCtaState,
    },
  });
}

function createMainLockedApiError() {
  return new ApiError("main is not available for unlock", {
    code: "http",
    details: JSON.stringify({
      error: {
        code: "main_locked",
        message: "main is not available for unlock",
      },
    }),
    status: 403,
  });
}

function createMockCardSetupSession() {
  return {
    apiBaseUrl: "https://api.ccbill.test",
    apiKey: "widget-api-key",
    clientAccount: "900000",
    currency: "JPY" as const,
    initialPeriod: "1",
    initialPrice: "1800.00",
    sessionToken: "card-setup-session-token",
    subAccount: "0001",
  };
}

describe("ImmersiveShortSurface", () => {
  const feedSurface = getFeedSurfaceByTab("recommended");
  const detailSurface = getShortSurfaceById("rooftop");
  const continueMainSurface = getShortSurfaceById("softlight");
  const directUnlockSurface = getShortSurfaceById("afterrain");
  const ownerPreviewSurface = getShortSurfaceById("balcony");
  const feedDialogTitle = "quiet rooftop preview の続きを見る";
  const detailDialogTitle = detailSurface ? "quiet rooftop preview の続きを見る" : "";
  const pinnedDetailOrigin = {
    from: "short" as const,
    shortFanTab: "pinned" as const,
    shortId: "rooftop",
  };

  beforeEach(() => {
    vi.stubEnv("NEXT_PUBLIC_API_BASE_URL", "https://api.example.com");
    mockedUpdateCreatorFollow.mockReset();
    mockedRequestCardSetupSession.mockReset();
    mockedRequestCardSetupToken.mockReset();
    mockedRequestMainAccessEntry.mockReset();
    mockedRequestMainPurchase.mockReset();
    mockedUseFanAuthDialogControls.mockReset();
    mockedRequestUnlockSurfaceByShortId.mockReset();
    mockedUseFanAuthDialog.mockReset();
    openFanAuthDialog.mockReset();
    mockedUseFanAuthDialogControls.mockReturnValue({
      closeFanAuthDialog: vi.fn(),
      openFanAuthDialog,
    });
    mockedUseFanAuthDialog.mockReturnValue({
      closeFanAuthDialog: vi.fn(),
      isFanAuthDialogOpen: false,
      openFanAuthDialog,
    });
  });

  afterEach(() => {
    push.mockReset();
    vi.unstubAllGlobals();
    vi.unstubAllEnvs();
  });

  it("opens the mini paywall for setup-required feed content", async () => {
    const user = userEvent.setup();
    mockedRequestUnlockSurfaceByShortId.mockResolvedValue(feedSurface.unlock);

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: true },
    );

    expect(screen.getByRole("link", { name: /For You/i })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("button", { name: /Unlock/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Mina Rei/i })).toHaveAttribute(
      "href",
      "/creators/creator_mina_rei?from=feed&tab=recommended",
    );
    expect(screen.queryByRole("link", { name: /Back/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Search" })).not.toBeInTheDocument();
    const feedActionRail = screen.getByTestId("feed-action-rail");
    const pinButton = screen.getByRole("button", { name: "Pinned short" });
    expect(feedActionRail).toHaveStyle("bottom: calc(228px + env(safe-area-inset-bottom, 0px))");
    expect(pinButton).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "プロフィールへ" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Share" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "More options" })).not.toBeInTheDocument();
    expect(screen.getByText("Follow")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    expect(await screen.findByRole("dialog", { name: feedDialogTitle })).toBeInTheDocument();
  });

  it("updates the feed playback progress bar from video metadata and timeupdate", () => {
    const { container } = renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: true },
    );

    const video = container.querySelector("video");

    if (!video) {
      throw new Error("video element missing");
    }

    const progressFill = screen.getByTestId("feed-playback-progress-fill");

    expect(progressFill).toHaveStyle("transform: scaleX(0)");

    Object.defineProperty(video, "duration", {
      configurable: true,
      value: 16,
    });
    Object.defineProperty(video, "currentTime", {
      configurable: true,
      value: 4,
      writable: true,
    });

    fireEvent(video, new Event("loadedmetadata"));

    expect(progressFill).toHaveStyle("transform: scaleX(0.25)");

    video.currentTime = 8;
    fireEvent(video, new Event("timeupdate"));

    expect(progressFill).toHaveStyle("transform: scaleX(0.5)");
  });

  it("seeks the feed video when the playback progress bar is clicked", () => {
    const { container } = renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: true },
    );

    const video = container.querySelector("video");

    if (!video) {
      throw new Error("video element missing");
    }

    Object.defineProperty(video, "duration", {
      configurable: true,
      value: 16,
    });
    Object.defineProperty(video, "currentTime", {
      configurable: true,
      value: 0,
      writable: true,
    });

    const progressBar = screen.getByTestId("feed-playback-progress-bar");

    vi.spyOn(progressBar, "getBoundingClientRect").mockReturnValue({
      bottom: 20,
      height: 20,
      left: 0,
      right: 100,
      toJSON: () => ({}),
      top: 0,
      width: 100,
      x: 0,
      y: 0,
    });

    fireEvent.click(progressBar, {
      clientX: 25,
    });

    expect(video.currentTime).toBe(4);
    expect(screen.getByTestId("feed-playback-progress-fill")).toHaveStyle("transform: scaleX(0.25)");
  });

  it("uses the short duration fallback when the feed video duration is unavailable during seek", () => {
    const { container } = renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: true },
    );

    const video = container.querySelector("video");

    if (!video) {
      throw new Error("video element missing");
    }

    Object.defineProperty(video, "duration", {
      configurable: true,
      value: Number.NaN,
    });
    Object.defineProperty(video, "currentTime", {
      configurable: true,
      value: 0,
      writable: true,
    });

    const progressBar = screen.getByTestId("feed-playback-progress-bar");

    vi.spyOn(progressBar, "getBoundingClientRect").mockReturnValue({
      bottom: 20,
      height: 20,
      left: 0,
      right: 100,
      toJSON: () => ({}),
      top: 0,
      width: 100,
      x: 0,
      y: 0,
    });

    fireEvent.click(progressBar, {
      clientX: 75,
    });

    expect(video.currentTime).toBe(12);
    expect(screen.getByTestId("feed-playback-progress-fill")).toHaveStyle("transform: scaleX(0.75)");
  });

  it("resets the feed playback progress bar when the feed short changes", async () => {
    const { rerender, container } = renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: true },
    );

    const video = container.querySelector("video");

    if (!video) {
      throw new Error("video element missing");
    }

    Object.defineProperty(video, "duration", {
      configurable: true,
      value: 16,
    });
    Object.defineProperty(video, "currentTime", {
      configurable: true,
      value: 8,
      writable: true,
    });

    fireEvent(video, new Event("loadedmetadata"));

    expect(screen.getByTestId("feed-playback-progress-fill")).toHaveStyle("transform: scaleX(0.5)");

    const nextSurface = {
      ...feedSurface,
      short: {
        ...feedSurface.short,
        id: "short_mina_rooftop_alt",
        media: {
          ...feedSurface.short.media,
          id: "asset_short_mina_rooftop_alt",
          url: "https://cdn.example.com/shorts/playback-alt.mp4",
        },
      },
      unlock: {
        ...feedSurface.unlock,
        short: {
          ...feedSurface.unlock.short,
          id: "short_mina_rooftop_alt",
          media: {
            ...feedSurface.unlock.short.media,
            id: "asset_short_mina_rooftop_alt",
            url: "https://cdn.example.com/shorts/playback-alt.mp4",
          },
        },
      },
    };

    rerender(
      <ViewerSessionProvider hasSession>
        <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={nextSurface} />
      </ViewerSessionProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("feed-playback-progress-fill")).toHaveStyle("transform: scaleX(0)");
    });
  });

  it("resolves the unlock surface before opening the API-backed paywall", async () => {
    const user = userEvent.setup();
    const resolvedUnlock = createResolvedApiUnlock({
      purchaseState: "setup_required",
      requiresAgeConfirmation: true,
      requiresCardSetup: true,
      requiresTermsAcceptance: true,
      unlockCtaState: "setup_required",
    });

    mockedRequestUnlockSurfaceByShortId.mockResolvedValue(resolvedUnlock);
    mockedRequestCardSetupSession.mockResolvedValue(createMockCardSetupSession());

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={createApiFeedSurface("setup_required")} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    expect(await screen.findByRole("dialog", { name: feedDialogTitle })).toBeInTheDocument();
    expect(mockedRequestUnlockSurfaceByShortId).toHaveBeenCalledWith({
      shortId: "short_mina_rooftop",
    });
    expect(mockedRequestCardSetupSession).toHaveBeenCalledWith({
      entryToken: "resolved-entry-token",
      fromShortId: "short_mina_rooftop",
      mainId: "main_mina_quiet_rooftop",
    });
  });

  it("reuses the resolved unlock state when reopening the API-backed paywall", async () => {
    const user = userEvent.setup();
    const resolvedUnlock = createResolvedApiUnlock({
      purchaseState: "setup_required",
      requiresAgeConfirmation: true,
      requiresCardSetup: true,
      requiresTermsAcceptance: true,
      unlockCtaState: "setup_required",
    });

    mockedRequestUnlockSurfaceByShortId.mockResolvedValue(resolvedUnlock);
    mockedRequestCardSetupSession.mockResolvedValue(createMockCardSetupSession());

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={createApiFeedSurface("setup_required")} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));
    expect(await screen.findByRole("dialog", { name: feedDialogTitle })).toBeInTheDocument();

    expect(mockedRequestUnlockSurfaceByShortId).toHaveBeenCalledTimes(1);
    expect(mockedRequestCardSetupSession).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole("button", { name: "閉じる" }));
    expect(screen.queryByRole("dialog", { name: feedDialogTitle })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Unlock/i }));
    expect(await screen.findByRole("dialog", { name: feedDialogTitle })).toBeInTheDocument();

    expect(mockedRequestUnlockSurfaceByShortId).toHaveBeenCalledTimes(1);
    expect(mockedRequestCardSetupSession).toHaveBeenCalledTimes(1);
  });

  it("shows a saved-card purchase CTA for API-backed direct unlock feed content", async () => {
    const user = userEvent.setup();
    const surface = createApiFeedSurface("unlock_available");
    const resolvedUnlock = createResolvedApiUnlock({
      purchaseState: "purchase_ready",
      savedPaymentMethods: [
        {
          brand: "visa",
          last4: "4242",
          paymentMethodId: "paymeth_saved_visa",
        },
      ],
      unlockCtaState: "unlock_available",
    });

    mockedRequestUnlockSurfaceByShortId.mockResolvedValue(resolvedUnlock);

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={surface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));
    expect(await screen.findByRole("dialog", { name: feedDialogTitle })).toBeInTheDocument();

    expect(mockedRequestUnlockSurfaceByShortId).toHaveBeenCalledWith({
      shortId: "short_mina_rooftop",
    });
    expect(screen.getByRole("button", { name: "Purchase ¥1,800" })).toBeInTheDocument();
  });

  it("tokenizes a new card in the API-backed paywall before purchase submission", async () => {
    const user = userEvent.setup();
    const surface = createApiFeedSurface("setup_required");
    const resolvedUnlock = createResolvedApiUnlock({
      purchaseState: "setup_required",
      requiresAgeConfirmation: true,
      requiresCardSetup: true,
      requiresTermsAcceptance: true,
      unlockCtaState: "setup_required",
    });

    mockedRequestUnlockSurfaceByShortId.mockResolvedValue(resolvedUnlock);
    mockedRequestCardSetupSession.mockResolvedValue(createMockCardSetupSession());
    mockedRequestCardSetupToken.mockResolvedValue({
      cardSetupToken: "opaque-card-setup-token",
    });
    mockedRequestMainPurchase.mockResolvedValue({
      access: {
        mainId: "main_mina_quiet_rooftop",
        reason: "unlock_required",
        status: "locked",
      },
      entryContext: null,
      purchase: {
        canRetry: true,
        failureReason: "purchase_declined",
        status: "failed",
      },
    });

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={surface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));
    expect(await screen.findByRole("dialog", { name: feedDialogTitle })).toBeInTheDocument();

    await user.click(screen.getByLabelText("18歳以上であり、年齢確認に同意する"));
    await user.click(screen.getByLabelText("利用規約とポリシーに同意し、main purchase へ進む"));
    await user.click(screen.getByRole("button", { name: "Mock card widget" }));

    await waitFor(() => {
      expect(mockedRequestCardSetupToken).toHaveBeenCalledWith({
        cardSetupSessionToken: "card-setup-session-token",
        entryToken: "resolved-entry-token",
        fromShortId: "short_mina_rooftop",
        mainId: "main_mina_quiet_rooftop",
        paymentTokenId: "widget-payment-token",
      });
    });

    await waitFor(() => {
      expect(mockedRequestMainPurchase).toHaveBeenCalledWith({
        acceptedAge: true,
        acceptedTerms: true,
        entryToken: "resolved-entry-token",
        fromShortId: "short_mina_rooftop",
        mainId: "main_mina_quiet_rooftop",
        paymentMethod: {
          cardSetupToken: "opaque-card-setup-token",
          mode: "new_card",
        },
        purchasePath: resolvedUnlock.entryContext.purchasePath,
      });
    });
  });

  it("opens main playback after a new-card purchase succeeds", async () => {
    const user = userEvent.setup();
    const surface = createApiFeedSurface("setup_required");
    const resolvedUnlock = createResolvedApiUnlock({
      purchaseState: "setup_required",
      requiresAgeConfirmation: true,
      requiresCardSetup: true,
      requiresTermsAcceptance: true,
      unlockCtaState: "setup_required",
    });

    mockedRequestUnlockSurfaceByShortId.mockResolvedValue(resolvedUnlock);
    mockedRequestCardSetupSession.mockResolvedValue(createMockCardSetupSession());
    mockedRequestCardSetupToken.mockResolvedValue({
      cardSetupToken: "opaque-card-setup-token",
    });
    mockedRequestMainPurchase.mockResolvedValue({
      access: {
        mainId: "main_mina_quiet_rooftop",
        reason: "purchased",
        status: "unlocked",
      },
      entryContext: {
        accessEntryPath: "/api/fan/mains/main_mina_quiet_rooftop/access-entry",
        purchasePath: "/api/fan/mains/main_mina_quiet_rooftop/purchase",
        token: "purchase-success-entry-token",
      },
      purchase: {
        canRetry: false,
        failureReason: null,
        status: "succeeded",
      },
    });
    mockedRequestMainAccessEntry.mockResolvedValue({
      href: "/mains/main_mina_quiet_rooftop?grant=purchase-success",
    });

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={surface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));
    expect(await screen.findByRole("dialog", { name: feedDialogTitle })).toBeInTheDocument();

    await user.click(screen.getByLabelText("18歳以上であり、年齢確認に同意する"));
    await user.click(screen.getByLabelText("利用規約とポリシーに同意し、main purchase へ進む"));
    await user.click(screen.getByRole("button", { name: "Mock card widget" }));

    await waitFor(() => {
      expect(mockedRequestMainAccessEntry).toHaveBeenCalledWith({
        entryToken: "purchase-success-entry-token",
        fromShortId: "short_mina_rooftop",
        mainId: "main_mina_quiet_rooftop",
        routePath: "/api/fan/mains/main_mina_quiet_rooftop/access-entry",
      });
      expect(push).toHaveBeenCalledWith("/mains/main_mina_quiet_rooftop?grant=purchase-success");
    });
  });

  it("does not exchange or purchase a new card before required consent is completed", async () => {
    const user = userEvent.setup();
    const surface = createApiFeedSurface("setup_required");
    const resolvedUnlock = createResolvedApiUnlock({
      purchaseState: "setup_required",
      requiresAgeConfirmation: true,
      requiresCardSetup: true,
      requiresTermsAcceptance: true,
      unlockCtaState: "setup_required",
    });

    mockedRequestUnlockSurfaceByShortId.mockResolvedValue(resolvedUnlock);
    mockedRequestCardSetupSession.mockResolvedValue(createMockCardSetupSession());

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={surface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));
    expect(await screen.findByRole("dialog", { name: feedDialogTitle })).toBeInTheDocument();

    expect(
      screen.getByText("年齢確認と利用規約への同意を完了すると card widget を表示します。"),
    ).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Mock card widget" })).not.toBeInTheDocument();
    expect(mockedRequestCardSetupToken).not.toHaveBeenCalled();
    expect(mockedRequestMainPurchase).not.toHaveBeenCalled();
  });

  it("recovers a stale card-setup session into the current continue-main access state", async () => {
    const user = userEvent.setup();
    const surface = createApiFeedSurface("setup_required");
    const initialUnlock = createResolvedApiUnlock({
      purchaseState: "setup_required",
      requiresAgeConfirmation: true,
      requiresCardSetup: true,
      requiresTermsAcceptance: true,
      unlockCtaState: "setup_required",
    });
    const recoveredUnlock = createResolvedApiUnlock({
      accessReason: "purchased",
      accessStatus: "unlocked",
      entryToken: "recovered-entry-token",
      purchaseState: "already_purchased",
      unlockCtaState: "continue_main",
    });

    mockedRequestUnlockSurfaceByShortId
      .mockResolvedValueOnce(initialUnlock)
      .mockResolvedValue(recoveredUnlock);
    mockedRequestCardSetupSession.mockRejectedValue(createMainLockedApiError());
    mockedRequestMainAccessEntry.mockResolvedValue({
      href: "/mains/main_mina_quiet_rooftop?grant=recovered-session",
    });

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={surface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    await waitFor(() => {
      expect(mockedRequestMainAccessEntry).toHaveBeenCalledWith({
        entryToken: "recovered-entry-token",
        fromShortId: "short_mina_rooftop",
        mainId: "main_mina_quiet_rooftop",
        routePath: recoveredUnlock.entryContext.accessEntryPath,
      });
      expect(push).toHaveBeenCalledWith("/mains/main_mina_quiet_rooftop?grant=recovered-session");
    });
  });

  it("recovers a stale card token exchange into the current continue-main access state", async () => {
    const user = userEvent.setup();
    const surface = createApiFeedSurface("setup_required");
    const initialUnlock = createResolvedApiUnlock({
      purchaseState: "setup_required",
      requiresAgeConfirmation: true,
      requiresCardSetup: true,
      requiresTermsAcceptance: true,
      unlockCtaState: "setup_required",
    });
    const recoveredUnlock = createResolvedApiUnlock({
      accessReason: "purchased",
      accessStatus: "unlocked",
      entryToken: "recovered-token-exchange-entry-token",
      purchaseState: "already_purchased",
      unlockCtaState: "continue_main",
    });

    mockedRequestUnlockSurfaceByShortId
      .mockResolvedValueOnce(initialUnlock)
      .mockResolvedValue(recoveredUnlock);
    mockedRequestCardSetupSession.mockResolvedValue(createMockCardSetupSession());
    mockedRequestCardSetupToken.mockRejectedValue(createMainLockedApiError());
    mockedRequestMainAccessEntry.mockResolvedValue({
      href: "/mains/main_mina_quiet_rooftop?grant=recovered-token-exchange",
    });

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={surface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));
    expect(await screen.findByRole("dialog", { name: feedDialogTitle })).toBeInTheDocument();

    await user.click(screen.getByLabelText("18歳以上であり、年齢確認に同意する"));
    await user.click(screen.getByLabelText("利用規約とポリシーに同意し、main purchase へ進む"));
    await user.click(screen.getByRole("button", { name: "Mock card widget" }));

    await waitFor(() => {
      expect(mockedRequestMainAccessEntry).toHaveBeenCalledWith({
        entryToken: "recovered-token-exchange-entry-token",
        fromShortId: "short_mina_rooftop",
        mainId: "main_mina_quiet_rooftop",
        routePath: recoveredUnlock.entryContext.accessEntryPath,
      });
      expect(push).toHaveBeenCalledWith("/mains/main_mina_quiet_rooftop?grant=recovered-token-exchange");
    });
  });

  it("refreshes the paywall state after a stale saved-card purchase hits main_locked", async () => {
    const user = userEvent.setup();
    const surface = createApiFeedSurface("unlock_available");
    const initialUnlock = createResolvedApiUnlock({
      entryToken: "stale-entry-token",
      purchaseState: "purchase_ready",
      savedPaymentMethods: [
        {
          brand: "visa",
          last4: "4242",
          paymentMethodId: "paymeth_saved_visa",
        },
      ],
      unlockCtaState: "unlock_available",
    });
    const refreshedUnlock = createResolvedApiUnlock({
      entryToken: "refreshed-entry-token",
      purchaseState: "purchase_ready",
      savedPaymentMethods: [
        {
          brand: "visa",
          last4: "4242",
          paymentMethodId: "paymeth_saved_visa",
        },
      ],
      unlockCtaState: "unlock_available",
    });

    mockedRequestUnlockSurfaceByShortId
      .mockResolvedValueOnce(initialUnlock)
      .mockResolvedValue(refreshedUnlock);
    mockedRequestMainPurchase
      .mockRejectedValueOnce(createMainLockedApiError())
      .mockResolvedValue({
        access: {
          mainId: "main_mina_quiet_rooftop",
          reason: "unlock_required",
          status: "locked",
        },
        entryContext: null,
        purchase: {
          canRetry: true,
          failureReason: "purchase_declined",
          status: "failed",
        },
      });

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={surface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));
    expect(await screen.findByRole("dialog", { name: feedDialogTitle })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Purchase ¥1,800" }));

    await waitFor(() => {
      expect(mockedRequestUnlockSurfaceByShortId).toHaveBeenNthCalledWith(2, {
        shortId: "short_mina_rooftop",
      });
      expect(screen.getByRole("button", { name: "Purchase ¥1,800" })).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: "Purchase ¥1,800" }));

    await waitFor(() => {
      expect(mockedRequestMainPurchase).toHaveBeenNthCalledWith(2, {
        acceptedAge: false,
        acceptedTerms: false,
        entryToken: "refreshed-entry-token",
        fromShortId: "short_mina_rooftop",
        mainId: "main_mina_quiet_rooftop",
        paymentMethod: {
          mode: "saved_card",
          paymentMethodId: "paymeth_saved_visa",
        },
        purchasePath: refreshedUnlock.entryContext.purchasePath,
      });
    });
  });

  it("updates the feed follow CTA after an authenticated follow succeeds", async () => {
    const user = userEvent.setup();

    mockedUpdateCreatorFollow.mockResolvedValue({
      stats: {
        fanCount: 12,
      },
      viewer: {
        isFollowing: true,
      },
    });

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Follow" }));

    await waitFor(() => {
      expect(mockedUpdateCreatorFollow).toHaveBeenCalledWith({
        action: "follow",
        creatorId: feedSurface.creator.id,
      });
      expect(screen.getByRole("button", { name: "Following" })).toHaveAttribute("aria-pressed", "true");
    });
  });

  it("keeps the feed follow CTA pending and ignores duplicate clicks", async () => {
    const user = userEvent.setup();
    let resolveUpdate: (() => void) | undefined;

    mockedUpdateCreatorFollow.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveUpdate = () => {
            resolve({
              stats: {
                fanCount: 12,
              },
              viewer: {
                isFollowing: true,
              },
            });
          };
        }),
    );

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: true },
    );

    const followButton = screen.getByRole("button", { name: "Follow" });

    await user.click(followButton);
    await user.click(screen.getByRole("button", { name: "Following..." }));

    expect(mockedUpdateCreatorFollow).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("button", { name: "Following..." })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Following..." })).toHaveAttribute("aria-busy", "true");

    resolveUpdate?.();

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Following" })).toBeInTheDocument();
    });
  });

  it("opens the shared auth dialog when an unauthenticated viewer presses feed follow", async () => {
    const user = userEvent.setup();

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: false },
    );

    await user.click(screen.getByRole("button", { name: "Follow" }));

    expect(openFanAuthDialog).toHaveBeenCalledTimes(1);
    expect(mockedUpdateCreatorFollow).not.toHaveBeenCalled();
  });

  it("reopens the shared auth dialog when the feed follow request returns auth_required", async () => {
    const user = userEvent.setup();

    mockedUpdateCreatorFollow.mockRejectedValue(
      new CreatorFollowApiError("auth_required", "creator follow requires authentication", {
        requestId: "req_creator_follow_put_auth_required_001",
        status: 401,
      }),
    );

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Follow" }));

    await waitFor(() => {
      expect(openFanAuthDialog).toHaveBeenCalledTimes(1);
    });
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("shows an inline error when feed follow fails", async () => {
    const user = userEvent.setup();

    mockedUpdateCreatorFollow.mockRejectedValue(new Error("boom"));

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Follow" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "フォロー状態を更新できませんでした。少し時間を置いてから再度お試しください。",
    );
  });

  it("opens the shared auth dialog before an unauthenticated viewer can open the paywall", async () => {
    const user = userEvent.setup();

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: false },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    expect(openFanAuthDialog).toHaveBeenCalledWith(
      expect.objectContaining({
        onAfterAuthenticated: expect.any(Function),
        postAuthNavigation: "none",
      }),
    );
  });

  it("resolves the unlock surface before reopening the paywall after auth", async () => {
    const user = userEvent.setup();
    const surface = createApiFeedSurface("unlock_available");
    const resolvedUnlock = createResolvedApiUnlock({
      purchaseState: "purchase_ready",
      unlockCtaState: "unlock_available",
    });

    mockedRequestUnlockSurfaceByShortId.mockResolvedValue(resolvedUnlock);
    mockedRequestCardSetupSession.mockResolvedValue(createMockCardSetupSession());

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={surface} />,
      { hasSession: false },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    const options = openFanAuthDialog.mock.calls[0]?.[0];

    if (!options?.onAfterAuthenticated) {
      throw new Error("unlock recovery callback missing");
    }

    await act(async () => {
      await options.onAfterAuthenticated?.();
    });

    await waitFor(() => {
      expect(mockedRequestUnlockSurfaceByShortId).toHaveBeenCalledWith({
        shortId: surface.short.id,
      });
      expect(screen.getByRole("dialog", { name: feedDialogTitle })).toBeInTheDocument();
    });
  });

  it("keeps the paywall closed when auth recovery resolves to continue-main", async () => {
    const user = userEvent.setup();
    const surface = createApiFeedSurface("unlock_available");
    const resolvedUnlock = createApiFeedSurface("continue_main").unlock;

    mockedRequestUnlockSurfaceByShortId.mockResolvedValue(resolvedUnlock);

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={surface} />,
      { hasSession: false },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    const options = openFanAuthDialog.mock.calls[0]?.[0];

    if (!options?.onAfterAuthenticated) {
      throw new Error("unlock recovery callback missing");
    }

    await act(async () => {
      await options.onAfterAuthenticated?.();
    });

    await waitFor(() => {
      expect(mockedRequestUnlockSurfaceByShortId).toHaveBeenCalledWith({
        shortId: surface.short.id,
      });
      expect(screen.queryByRole("dialog", { name: feedDialogTitle })).not.toBeInTheDocument();
      expect(screen.getByRole("button", { name: /Continue main/i })).toBeInTheDocument();
    });
  });

  it("keeps the paywall closed when auth recovery resolves to owner-preview", async () => {
    const user = userEvent.setup();
    const surface = createApiFeedSurface("unlock_available");
    const resolvedUnlock = createApiFeedSurface("owner_preview").unlock;

    mockedRequestUnlockSurfaceByShortId.mockResolvedValue(resolvedUnlock);

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={surface} />,
      { hasSession: false },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    const options = openFanAuthDialog.mock.calls[0]?.[0];

    if (!options?.onAfterAuthenticated) {
      throw new Error("unlock recovery callback missing");
    }

    await act(async () => {
      await options.onAfterAuthenticated?.();
    });

    await waitFor(() => {
      expect(mockedRequestUnlockSurfaceByShortId).toHaveBeenCalledWith({
        shortId: surface.short.id,
      });
      expect(screen.queryByRole("dialog", { name: feedDialogTitle })).not.toBeInTheDocument();
      expect(screen.getByRole("button", { name: /Owner preview/i })).toBeInTheDocument();
    });
  });

  it("refreshes stale continue-main auth recovery into the current paywall state", async () => {
    const user = userEvent.setup();
    const surface = createApiFeedSurface("continue_main");
    const resolvedUnlock = createResolvedApiUnlock({
      purchaseState: "setup_required",
      requiresAgeConfirmation: true,
      requiresCardSetup: true,
      requiresTermsAcceptance: true,
      unlockCtaState: "setup_required",
    });

    mockedRequestUnlockSurfaceByShortId.mockResolvedValue(resolvedUnlock);
    mockedRequestCardSetupSession.mockResolvedValue(createMockCardSetupSession());

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={surface} />,
      { hasSession: false },
    );

    await user.click(screen.getByRole("button", { name: /Continue main/i }));

    const options = openFanAuthDialog.mock.calls[0]?.[0];

    if (!options?.onAfterAuthenticated) {
      throw new Error("unlock recovery callback missing");
    }

    await act(async () => {
      await options.onAfterAuthenticated?.();
    });

    await waitFor(() => {
      expect(mockedRequestUnlockSurfaceByShortId).toHaveBeenCalledWith({
        shortId: surface.short.id,
      });
      expect(screen.getByRole("dialog", { name: feedDialogTitle })).toBeInTheDocument();
    });
  });

  it("drops a stale resolved unlock immediately when the parent provides newer unlock props for the same short", async () => {
    const user = userEvent.setup();
    const initialSurface = createApiFeedSurface("unlock_available");
    const resolvedUnlock = createApiFeedSurface("continue_main").unlock;
    const nextSurface = createApiFeedSurface("owner_preview");

    mockedRequestUnlockSurfaceByShortId.mockResolvedValue(resolvedUnlock);

    const view = renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={initialSurface} />,
      { hasSession: false },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    const options = openFanAuthDialog.mock.calls[0]?.[0];

    if (!options?.onAfterAuthenticated) {
      throw new Error("unlock recovery callback missing");
    }

    await act(async () => {
      await options.onAfterAuthenticated?.();
    });

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /Continue main/i })).toBeInTheDocument();
    });

    view.rerender(
      <ViewerSessionProvider hasSession={false}>
        <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={nextSurface} />
      </ViewerSessionProvider>,
    );

    expect(screen.getByRole("button", { name: /Owner preview/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Continue main/i })).not.toBeInTheDocument();
  });

  it("preserves paywall setup selections across auth-required recovery", async () => {
    const user = userEvent.setup();
    vi.stubGlobal(
      "fetch",
      vi.fn<typeof fetch>().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: {
              code: "auth_required",
              message: "login required",
            },
          }),
          {
            headers: {
              "Content-Type": "application/json",
            },
            status: 401,
          },
        ),
      ),
    );

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));
    await user.click(screen.getByLabelText("18歳以上であり、年齢確認に同意する"));
    await user.click(screen.getByLabelText("利用規約とポリシーに同意し、main purchase へ進む"));
    await user.click(screen.getByRole("button", { name: "Unlock ¥1,800 | 8分" }));

    const options = openFanAuthDialog.mock.calls[0]?.[0];

    if (!options?.onAfterAuthenticated) {
      throw new Error("unlock recovery callback missing");
    }

    await act(async () => {
      await options.onAfterAuthenticated?.();
    });

    await waitFor(() => {
      expect(screen.getByRole("dialog", { name: feedDialogTitle })).toBeInTheDocument();
      expect(screen.getByLabelText("18歳以上であり、年齢確認に同意する")).toBeChecked();
      expect(screen.getByLabelText("利用規約とポリシーに同意し、main purchase へ進む")).toBeChecked();
    });
  });

  it("restores the paywall after re-auth success when fresh-auth blocks unlock entry", async () => {
    const user = userEvent.setup();
    vi.stubGlobal(
      "fetch",
      vi.fn<typeof fetch>().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: {
              code: "fresh_auth_required",
              message: "recent auth required",
            },
          }),
          {
            headers: {
              "Content-Type": "application/json",
            },
            status: 403,
          },
        ),
      ),
    );

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));
    await user.click(screen.getByLabelText("18歳以上であり、年齢確認に同意する"));
    await user.click(screen.getByLabelText("利用規約とポリシーに同意し、main purchase へ進む"));
    await user.click(screen.getByRole("button", { name: "Unlock ¥1,800 | 8分" }));

    const options = openFanAuthDialog.mock.calls[0]?.[0];

    expect(options).toEqual(
      expect.objectContaining({
        allowClose: false,
        initialMode: "re-auth",
        onAfterAuthenticated: expect.any(Function),
        postAuthNavigation: "none",
      }),
    );

    await act(async () => {
      await options.onAfterAuthenticated?.();
    });

    await waitFor(() => {
      expect(screen.getByRole("dialog", { name: feedDialogTitle })).toBeInTheDocument();
      expect(screen.getByLabelText("18歳以上であり、年齢確認に同意する")).toBeChecked();
      expect(screen.getByLabelText("利用規約とポリシーに同意し、main purchase へ進む")).toBeChecked();
    });
  });

  it("resolves the unlock surface before reopening detail paywall after auth", async () => {
    const user = userEvent.setup();
    const surface = {
      ...createApiDetailSurface("setup_required"),
      mainEntryEnabled: true,
    };
    const resolvedUnlock = createResolvedApiUnlock({
      purchaseState: "setup_required",
      requiresAgeConfirmation: true,
      requiresCardSetup: true,
      requiresTermsAcceptance: true,
      unlockCtaState: "setup_required",
    });

    mockedRequestUnlockSurfaceByShortId.mockResolvedValue(resolvedUnlock);
    mockedRequestCardSetupSession.mockResolvedValue(createMockCardSetupSession());

    renderWithViewerSession(
      <ImmersiveShortSurface
        backHref="/"
        creatorProfileOrigin={{ from: "short", shortId: "short_mina_rooftop" }}
        mode="detail"
        surface={surface}
      />,
      { hasSession: false },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    const options = openFanAuthDialog.mock.calls[0]?.[0];

    if (!options?.onAfterAuthenticated) {
      throw new Error("unlock recovery callback missing");
    }

    await act(async () => {
      await options.onAfterAuthenticated?.();
    });

    await waitFor(() => {
      expect(mockedRequestUnlockSurfaceByShortId).toHaveBeenCalledWith({
        shortId: surface.short.id,
      });
      expect(screen.getByRole("dialog", { name: detailDialogTitle })).toBeInTheDocument();
    });
  });

  it("renders detail mode with back navigation and the same creator block", async () => {
    if (!detailSurface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    renderWithViewerSession(
      <ImmersiveShortSurface backHref="/" creatorProfileOrigin={pinnedDetailOrigin} mode="detail" surface={detailSurface} />,
      { hasSession: true },
    );

    expect(screen.getByRole("link", { name: /Back/i })).toHaveAttribute("href", "/");
    expect(screen.getByRole("link", { name: /Mina Rei/i })).toHaveAttribute(
      "href",
      "/creators/creator_mina_rei?from=short&shortFanTab=pinned&shortId=rooftop",
    );
    expect(screen.getByRole("heading", { level: 1, name: "Short detail" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /For You/i })).not.toBeInTheDocument();
    expect(screen.getByText(detailSurface.short.caption)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Unlock/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Following" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    expect(screen.getByRole("dialog", { name: detailDialogTitle })).toBeInTheDocument();
  });

  it("renders feed-like detail presentation with back navigation for creator and pinned sources", () => {
    if (!detailSurface) {
      throw new Error("fixture missing");
    }

    renderWithViewerSession(
      <ImmersiveShortSurface
        backHref="/fan?tab=pinned"
        creatorProfileOrigin={pinnedDetailOrigin}
        mode="detail"
        presentation="feedLike"
        surface={detailSurface}
      />,
      { hasSession: true },
    );

    expect(screen.getByRole("link", { name: /Back/i })).toHaveAttribute("href", "/fan?tab=pinned");
    expect(screen.getByRole("link", { name: /Mina Rei/i })).toHaveAttribute(
      "href",
      "/creators/creator_mina_rei?from=short&shortFanTab=pinned&shortId=rooftop",
    );
    expect(screen.getByRole("heading", { level: 1, name: "Short detail" })).toBeInTheDocument();
    expect(screen.getByTestId("feed-action-rail")).toBeInTheDocument();
    expect(screen.getByTestId("feed-playback-progress-bar")).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /For You/i })).not.toBeInTheDocument();
    expect(screen.getByText(detailSurface.short.caption)).toBeInTheDocument();
  });

  it("keeps the feed accessibility heading when feed mode uses the feed-like presentation branch", () => {
    renderWithViewerSession(<ImmersiveShortSurface activeTab="recommended" mode="feed" surface={feedSurface} />, {
      hasSession: true,
    });

    expect(screen.getByRole("heading", { level: 1, name: "Feed" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { level: 1, name: "Short detail" })).not.toBeInTheDocument();
  });

  it("updates the detail follow CTA after an authenticated unfollow succeeds", async () => {
    if (!detailSurface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    mockedUpdateCreatorFollow.mockResolvedValue({
      stats: {
        fanCount: 11,
      },
      viewer: {
        isFollowing: false,
      },
    });

    renderWithViewerSession(
      <ImmersiveShortSurface backHref="/" creatorProfileOrigin={pinnedDetailOrigin} mode="detail" surface={detailSurface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Following" }));

    await waitFor(() => {
      expect(mockedUpdateCreatorFollow).toHaveBeenCalledWith({
        action: "unfollow",
        creatorId: detailSurface.creator.id,
      });
      expect(screen.getByRole("button", { name: "Follow" })).toHaveAttribute("aria-pressed", "false");
    });
  });

  it("updates the detail pin CTA after an authenticated unpin succeeds", async () => {
    if (!detailSurface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    vi.stubGlobal(
      "fetch",
      vi.fn<typeof fetch>().mockResolvedValue(
        new Response(
          JSON.stringify({
            data: {
              viewer: {
                isPinned: false,
              },
            },
            error: null,
            meta: {
              page: null,
              requestId: "req_short_pin_delete_success_001",
            },
          }),
          {
            headers: {
              "Content-Type": "application/json",
            },
            status: 200,
          },
        ),
      ),
    );

    renderWithViewerSession(
      <ImmersiveShortSurface backHref="/" creatorProfileOrigin={pinnedDetailOrigin} mode="detail" surface={detailSurface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: "Pinned short" }));

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Pin short" })).toHaveAttribute("aria-pressed", "false");
    });
  });

  it("opens the shared auth dialog when detail pin is tapped without a session", async () => {
    if (!detailSurface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    renderWithViewerSession(
      <ImmersiveShortSurface backHref="/" creatorProfileOrigin={pinnedDetailOrigin} mode="detail" surface={detailSurface} />,
      { hasSession: false },
    );

    await user.click(screen.getByRole("button", { name: "Pinned short" }));

    expect(openFanAuthDialog).toHaveBeenCalledWith({
      postAuthNavigation: "none",
    });
  });

  it("reopens the shared auth dialog when continue-main receives auth_required", async () => {
    if (!continueMainSurface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    vi.stubGlobal(
      "fetch",
      vi.fn<typeof fetch>().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: {
              code: "auth_required",
              message: "login required",
            },
          }),
          {
            headers: {
              "Content-Type": "application/json",
            },
            status: 401,
          },
        ),
      ),
    );

    renderWithViewerSession(
      <ImmersiveShortSurface
        backHref="/"
        creatorProfileOrigin={{ from: "short", shortId: "softlight" }}
        mode="detail"
        surface={continueMainSurface}
      />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: /Continue main/i }));

    await waitFor(() => {
      expect(openFanAuthDialog).toHaveBeenCalledWith(
        expect.objectContaining({
          postAuthNavigation: "none",
        }),
      );
    });
  });

  it("opens the shared re-auth dialog when continue-main receives fresh_auth_required", async () => {
    if (!continueMainSurface) {
      throw new Error("fixture missing");
    }

    const user = userEvent.setup();

    vi.stubGlobal(
      "fetch",
      vi.fn<typeof fetch>().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: {
              code: "fresh_auth_required",
              message: "recent auth required",
            },
          }),
          {
            headers: {
              "Content-Type": "application/json",
            },
            status: 403,
          },
        ),
      ),
    );

    renderWithViewerSession(
      <ImmersiveShortSurface
        backHref="/"
        creatorProfileOrigin={{ from: "short", shortId: "softlight" }}
        mode="detail"
        surface={continueMainSurface}
      />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: /Continue main/i }));

    await waitFor(() => {
      expect(openFanAuthDialog).toHaveBeenCalledWith({
        allowClose: false,
        initialMode: "re-auth",
        postAuthNavigation: "none",
      });
    });
  });

  it("renders direct-unlock detail content as an action button", () => {
    if (!directUnlockSurface) {
      throw new Error("fixture missing");
    }

    renderWithViewerSession(
      <ImmersiveShortSurface
        backHref="/"
        creatorProfileOrigin={{ from: "short", shortId: "afterrain" }}
        mode="detail"
        surface={directUnlockSurface}
      />,
      { hasSession: true },
    );

    expect(screen.getByRole("button", { name: /Unlock/i })).toBeInTheDocument();
  });

  it("renders owner-preview detail content as an action button", () => {
    if (!ownerPreviewSurface) {
      throw new Error("fixture missing");
    }

    renderWithViewerSession(
      <ImmersiveShortSurface
        backHref="/"
        creatorProfileOrigin={{ from: "short", shortId: "balcony" }}
        mode="detail"
        surface={ownerPreviewSurface}
      />,
      { hasSession: true },
    );

    expect(screen.getByRole("button", { name: /Owner preview/i })).toBeInTheDocument();
  });

  it("renders creator initials when the creator has no custom avatar", () => {
    renderWithViewerSession(
      <ImmersiveShortSurface
        activeTab="recommended"
        mode="feed"
        surface={{ ...feedSurface, creator: { ...feedSurface.creator, avatar: null } }}
      />,
      { hasSession: true },
    );

    expect(screen.getAllByText("MR").length).toBeGreaterThan(0);
  });

  it("falls back to a generic paywall title when the short has no caption", async () => {
    const user = userEvent.setup();
    const surface = {
      ...feedSurface,
      short: {
        ...feedSurface.short,
        caption: "",
      },
      unlock: {
        ...feedSurface.unlock,
        short: {
          ...feedSurface.unlock.short,
          caption: "",
        },
      },
    };
    mockedRequestUnlockSurfaceByShortId.mockResolvedValue(surface.unlock);

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={surface} />,
      { hasSession: true },
    );

    await user.click(screen.getByRole("button", { name: /Unlock/i }));

    expect(await screen.findByRole("dialog", { name: "この short の続きを見る" })).toBeInTheDocument();
  });

  it("renders feed mode without short theme lookup for unknown short ids", () => {
    const surface = {
      ...feedSurface,
      short: {
        ...feedSurface.short,
        id: "short_dbcc1756d3d9406988e6860c7348609c",
      },
      unlock: {
        ...feedSurface.unlock,
        short: {
          ...feedSurface.unlock.short,
          id: "short_dbcc1756d3d9406988e6860c7348609c",
        },
      },
    };

    renderWithViewerSession(
      <ImmersiveShortSurface activeTab="recommended" mode="feed" surface={surface} />,
      { hasSession: true },
    );

    expect(screen.getByRole("link", { name: /For You/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Unlock/i })).toBeInTheDocument();
  });
});

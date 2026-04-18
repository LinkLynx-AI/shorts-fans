import userEvent from "@testing-library/user-event";
import { render, screen } from "@testing-library/react";

import { normalizeUnlockSurface, type UnlockSurfaceModel } from "@/features/unlock-entry";

import { UnlockPaywallDialog, type PaywallPaymentSelection } from "./unlock-paywall-dialog";

vi.mock("./ccbill-payment-widget", () => ({
  CCBillPaymentWidget: ({
    onPaymentTokenCreated,
  }: {
    onPaymentTokenCreated: (paymentTokenId: string) => void;
  }) => (
    <button
      onClick={() => {
        onPaymentTokenCreated("widget-payment-token");
      }}
      type="button"
    >
      Mock card widget
    </button>
  ),
}));

function createUnlockModel({
  pendingReason = null,
  purchaseState,
  savedPaymentMethods = [],
}: {
  pendingReason?: UnlockSurfaceModel["purchase"]["pendingReason"];
  purchaseState: UnlockSurfaceModel["purchase"]["state"];
  savedPaymentMethods?: UnlockSurfaceModel["purchase"]["savedPaymentMethods"];
}) {
  const accessReason =
    purchaseState === "already_purchased"
      ? "purchased"
      : purchaseState === "owner_preview"
        ? "owner_preview"
        : "unlock_required";
  const accessStatus =
    purchaseState === "already_purchased"
      ? "unlocked"
      : purchaseState === "owner_preview"
        ? "owner"
        : "locked";
  const unlockCtaState =
    purchaseState === "already_purchased"
      ? "continue_main"
      : purchaseState === "owner_preview"
        ? "owner_preview"
        : purchaseState === "purchase_ready"
          ? "unlock_available"
          : "setup_required";

  return normalizeUnlockSurface({
    access: {
      mainId: "main_mina_quiet_rooftop",
      reason: accessReason,
      status: accessStatus,
    },
    creator: {
      avatar: null,
      bio: "night preview specialist",
      displayName: "Mina Rei",
      handle: "@minarei",
      id: "creator_mina_rei",
    },
    entryContext: {
      accessEntryPath: "/api/fan/mains/main_mina_quiet_rooftop/access-entry",
      purchasePath: "/api/fan/mains/main_mina_quiet_rooftop/purchase",
      token: "entry-token",
    },
    main: {
      durationSeconds: 480,
      id: "main_mina_quiet_rooftop",
      priceJpy: 1800,
    },
    purchase: {
      pendingReason,
      savedPaymentMethods,
      setup: {
        required: purchaseState === "setup_required",
        requiresAgeConfirmation: purchaseState === "setup_required",
        requiresCardSetup: purchaseState === "setup_required",
        requiresTermsAcceptance: purchaseState === "setup_required",
      },
      state: purchaseState,
      supportedCardBrands: ["visa", "mastercard", "jcb", "american_express"],
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
      mainDurationSeconds: purchaseState === "already_purchased" ? null : 480,
      priceJpy: purchaseState === "already_purchased" ? null : 1800,
      resumePositionSeconds: purchaseState === "already_purchased" ? 120 : null,
      state: unlockCtaState,
    },
  });
}

function renderDialog({
  selection,
  unlock,
}: {
  selection: PaywallPaymentSelection;
  unlock: UnlockSurfaceModel;
}) {
  return render(
    <UnlockPaywallDialog
      acceptAge={false}
      acceptTerms={false}
      cardSetupSession={{
        apiBaseUrl: "https://api.ccbill.test",
        apiKey: "widget-api-key",
        clientAccount: "900000",
        currency: "JPY",
        initialPeriod: "1",
        initialPrice: "1800.00",
        sessionToken: "card-setup-session-token",
        subAccount: "0001",
      }}
      onAcceptAgeChange={vi.fn()}
      onAcceptTermsChange={vi.fn()}
      onCardPaymentTokenCreated={vi.fn()}
      onClose={vi.fn()}
      onConfirm={vi.fn()}
      onPaymentSelectionChange={vi.fn()}
      open
      selection={selection}
      unlock={unlock}
    />,
  );
}

describe("UnlockPaywallDialog", () => {
  it("renders saved-card purchase controls for purchase-ready state", () => {
    renderDialog({
      selection: {
        mode: "saved_card",
        paymentMethodId: "paymeth_saved_visa",
      },
      unlock: createUnlockModel({
        purchaseState: "purchase_ready",
        savedPaymentMethods: [
          {
            brand: "visa",
            last4: "4242",
            paymentMethodId: "paymeth_saved_visa",
          },
        ],
      }),
    });

    expect(screen.getByText("4 ブランドの card だけで purchase できます。")).toBeInTheDocument();
    expect(screen.getByText("Visa •••• 4242")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Purchase ¥1,800" })).toBeInTheDocument();
  });

  it("renders the embedded widget for new-card selection and forwards widget tokens", async () => {
    const user = userEvent.setup();
    const onCardPaymentTokenCreated = vi.fn();

    render(
      <UnlockPaywallDialog
        acceptAge
        acceptTerms
        cardSetupSession={{
          apiBaseUrl: "https://api.ccbill.test",
          apiKey: "widget-api-key",
          clientAccount: "900000",
          currency: "JPY",
          initialPeriod: "1",
          initialPrice: "1800.00",
          sessionToken: "card-setup-session-token",
          subAccount: "0001",
        }}
        onAcceptAgeChange={vi.fn()}
        onAcceptTermsChange={vi.fn()}
        onCardPaymentTokenCreated={onCardPaymentTokenCreated}
        onClose={vi.fn()}
        onConfirm={vi.fn()}
        onPaymentSelectionChange={vi.fn()}
        open
        selection={{
          mode: "new_card",
        }}
        unlock={createUnlockModel({
          purchaseState: "setup_required",
        })}
      />,
    );

    expect(screen.queryByRole("button", { name: /Purchase ¥/ })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Mock card widget" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Mock card widget" }));

    expect(onCardPaymentTokenCreated).toHaveBeenCalledWith("widget-payment-token");
  });

  it("withholds the new-card widget until required consent is completed", () => {
    const { rerender } = render(
      <UnlockPaywallDialog
        acceptAge={false}
        acceptTerms={false}
        cardSetupSession={{
          apiBaseUrl: "https://api.ccbill.test",
          apiKey: "widget-api-key",
          clientAccount: "900000",
          currency: "JPY",
          initialPeriod: "1",
          initialPrice: "1800.00",
          sessionToken: "card-setup-session-token",
          subAccount: "0001",
        }}
        onAcceptAgeChange={vi.fn()}
        onAcceptTermsChange={vi.fn()}
        onCardPaymentTokenCreated={vi.fn()}
        onClose={vi.fn()}
        onConfirm={vi.fn()}
        onPaymentSelectionChange={vi.fn()}
        open
        selection={{
          mode: "new_card",
        }}
        unlock={createUnlockModel({
          purchaseState: "setup_required",
        })}
      />,
    );

    expect(screen.queryByRole("button", { name: "Mock card widget" })).not.toBeInTheDocument();
    expect(screen.getByText("年齢確認と利用規約への同意を完了すると card widget を表示します。")).toBeInTheDocument();

    rerender(
      <UnlockPaywallDialog
        acceptAge
        acceptTerms
        cardSetupSession={{
          apiBaseUrl: "https://api.ccbill.test",
          apiKey: "widget-api-key",
          clientAccount: "900000",
          currency: "JPY",
          initialPeriod: "1",
          initialPrice: "1800.00",
          sessionToken: "card-setup-session-token",
          subAccount: "0001",
        }}
        onAcceptAgeChange={vi.fn()}
        onAcceptTermsChange={vi.fn()}
        onCardPaymentTokenCreated={vi.fn()}
        onClose={vi.fn()}
        onConfirm={vi.fn()}
        onPaymentSelectionChange={vi.fn()}
        open
        selection={{
          mode: "new_card",
        }}
        unlock={createUnlockModel({
          purchaseState: "setup_required",
        })}
      />,
    );

    expect(screen.getByRole("button", { name: "Mock card widget" })).toBeInTheDocument();
  });

  it("renders continue and owner actions for already-purchased and owner-preview states", () => {
    const { rerender } = renderDialog({
      selection: {
        mode: "saved_card",
        paymentMethodId: "paymeth_saved_visa",
      },
      unlock: createUnlockModel({
        purchaseState: "already_purchased",
        savedPaymentMethods: [
          {
            brand: "visa",
            last4: "4242",
            paymentMethodId: "paymeth_saved_visa",
          },
        ],
      }),
    });

    expect(screen.getByText("購入済みのため、そのまま再開できます。")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Continue main" })).toBeInTheDocument();

    rerender(
      <UnlockPaywallDialog
        acceptAge={false}
        acceptTerms={false}
        cardSetupSession={{
          apiBaseUrl: "https://api.ccbill.test",
          apiKey: "widget-api-key",
          clientAccount: "900000",
          currency: "JPY",
          initialPeriod: "1",
          initialPrice: "1800.00",
          sessionToken: "card-setup-session-token",
          subAccount: "0001",
        }}
        onAcceptAgeChange={vi.fn()}
        onAcceptTermsChange={vi.fn()}
        onCardPaymentTokenCreated={vi.fn()}
        onClose={vi.fn()}
        onConfirm={vi.fn()}
        onPaymentSelectionChange={vi.fn()}
        open
        selection={{
          mode: "saved_card",
          paymentMethodId: "paymeth_saved_visa",
        }}
        unlock={createUnlockModel({
          purchaseState: "owner_preview",
          savedPaymentMethods: [
            {
              brand: "visa",
              last4: "4242",
              paymentMethodId: "paymeth_saved_visa",
            },
          ],
        })}
      />,
    );

    expect(screen.getByText("owner preview で続きへ進めます。")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Owner preview" })).toBeInTheDocument();
  });

  it("renders pending and unavailable states without purchase CTA", () => {
    const { rerender } = renderDialog({
      selection: {
        mode: "saved_card",
        paymentMethodId: "paymeth_saved_visa",
      },
      unlock: createUnlockModel({
        pendingReason: "provider_processing",
        purchaseState: "purchase_pending",
        savedPaymentMethods: [
          {
            brand: "visa",
            last4: "4242",
            paymentMethodId: "paymeth_saved_visa",
          },
        ],
      }),
    });

    expect(screen.getByText("決済処理を確認しています。")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Purchase ¥/ })).not.toBeInTheDocument();
    expect(screen.getByText("pending reason: provider_processing")).toBeInTheDocument();

    rerender(
      <UnlockPaywallDialog
        acceptAge={false}
        acceptTerms={false}
        cardSetupSession={{
          apiBaseUrl: "https://api.ccbill.test",
          apiKey: "widget-api-key",
          clientAccount: "900000",
          currency: "JPY",
          initialPeriod: "1",
          initialPrice: "1800.00",
          sessionToken: "card-setup-session-token",
          subAccount: "0001",
        }}
        onAcceptAgeChange={vi.fn()}
        onAcceptTermsChange={vi.fn()}
        onCardPaymentTokenCreated={vi.fn()}
        onClose={vi.fn()}
        onConfirm={vi.fn()}
        onPaymentSelectionChange={vi.fn()}
        open
        selection={{
          mode: "saved_card",
          paymentMethodId: "paymeth_saved_visa",
        }}
        unlock={createUnlockModel({
          purchaseState: "unavailable",
        })}
      />,
    );

    expect(screen.getByText("現在は利用できません。")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Purchase ¥/ })).not.toBeInTheDocument();
    expect(screen.getByText("現在は purchase action を表示していません。")).toBeInTheDocument();
  });

  it("renders new-card loading and error messaging", () => {
    render(
      <UnlockPaywallDialog
        acceptAge
        acceptTerms
        cardSetupErrorMessage="card setup session の準備に失敗しました。"
        cardSetupSession={null}
        isLoadingCardSetupSession
        onAcceptAgeChange={vi.fn()}
        onAcceptTermsChange={vi.fn()}
        onCardPaymentTokenCreated={vi.fn()}
        onClose={vi.fn()}
        onConfirm={vi.fn()}
        onPaymentSelectionChange={vi.fn()}
        open
        purchaseErrorMessage="purchase を完了できませんでした。"
        selection={{
          mode: "new_card",
        }}
        unlock={createUnlockModel({
          purchaseState: "setup_required",
        })}
      />,
    );

    expect(screen.getByText("card setup session を準備しています。")).toBeInTheDocument();
    expect(screen.getByText("card setup session の準備に失敗しました。")).toBeInTheDocument();
    expect(screen.getByText("purchase を完了できませんでした。")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Purchase ¥/ })).not.toBeInTheDocument();
  });
});

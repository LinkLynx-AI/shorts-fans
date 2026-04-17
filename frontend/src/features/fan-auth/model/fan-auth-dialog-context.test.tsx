import userEvent from "@testing-library/user-event";
import { render, screen, waitFor } from "@testing-library/react";

import {
  CurrentViewerProvider,
  ViewerSessionProvider,
} from "@/entities/viewer";
import {
  FanAuthApiError,
  FanAuthDialogProvider,
  useFanAuthDialog,
} from "@/features/fan-auth";

const getCurrentViewerBootstrapMock = vi.hoisted(() => vi.fn());
const mockedRouter = vi.hoisted(() => ({
  back: vi.fn(),
  forward: vi.fn(),
  prefetch: vi.fn(),
  push: vi.fn(),
  refresh: vi.fn(),
  replace: vi.fn(),
}));
const reAuthenticateFanMock = vi.hoisted(() => vi.fn());
const signInFanMock = vi.hoisted(() => vi.fn());

vi.mock("next/navigation", async () => {
  const actual = await vi.importActual<typeof import("next/navigation")>("next/navigation");

  return {
    ...actual,
    useRouter: () => mockedRouter,
  };
});

vi.mock("@/features/fan-auth/api/request-fan-auth", () => ({
  confirmFanPasswordReset: vi.fn(),
  confirmFanSignUp: vi.fn(),
  reAuthenticateFan: reAuthenticateFanMock,
  signInFan: signInFanMock,
  signUpFan: vi.fn(),
  startFanPasswordReset: vi.fn(),
}));

vi.mock("@/entities/viewer", async () => {
  const actual = await vi.importActual<typeof import("@/entities/viewer")>("@/entities/viewer");

  return {
    ...actual,
    getCurrentViewerBootstrap: getCurrentViewerBootstrapMock,
  };
});

function FanAuthDialogTrigger({
  afterAuthenticatedHref = "/fan",
  initialMode = "sign-in",
  onAfterAuthenticated,
}: {
  afterAuthenticatedHref?: string | undefined;
  initialMode?: "re-auth" | "sign-in";
  onAfterAuthenticated?: (() => Promise<void> | void) | undefined;
}) {
  const { openFanAuthDialog } = useFanAuthDialog();

  return (
    <button
      onClick={() =>
        openFanAuthDialog({
          afterAuthenticatedHref,
          initialMode,
          onAfterAuthenticated,
        })
      }
      type="button"
    >
      open auth dialog
    </button>
  );
}

describe("FanAuthDialogProvider", () => {
  beforeEach(() => {
    getCurrentViewerBootstrapMock.mockReset();
    mockedRouter.back.mockReset();
    mockedRouter.forward.mockReset();
    mockedRouter.prefetch.mockReset();
    mockedRouter.push.mockReset();
    mockedRouter.refresh.mockReset();
    mockedRouter.replace.mockReset();
    reAuthenticateFanMock.mockReset();
    signInFanMock.mockReset();
  });

  it("keeps the modal open and shows a recoverable error when bootstrap resolves null after auth success", async () => {
    const user = userEvent.setup();

    signInFanMock.mockResolvedValue(undefined);
    getCurrentViewerBootstrapMock.mockResolvedValue(null);

    render(
      <ViewerSessionProvider hasSession={false}>
        <CurrentViewerProvider currentViewer={null}>
          <FanAuthDialogProvider>
            <FanAuthDialogTrigger />
          </FanAuthDialogProvider>
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    await user.click(screen.getByRole("button", { name: "open auth dialog" }));
    await user.type(screen.getByRole("textbox", { name: "Email" }), "fan@example.com");
    await user.type(screen.getByLabelText("Password"), "VeryStrongPass123!");
    await user.click(screen.getByRole("button", { name: "サインインを続ける" }));

    await waitFor(() => {
      expect(getCurrentViewerBootstrapMock).toHaveBeenCalledWith({
        credentials: "include",
      });
    });

    expect(
      await screen.findByRole("alert"),
    ).toHaveTextContent(
      "認証自体は完了しましたが、状態反映の確認に失敗しました。画面を更新して確認してください。",
    );
    expect(screen.getByRole("dialog", { name: "続けるには認証が必要です" })).toBeInTheDocument();
  });

  it("closes re-auth without refreshing the route or re-reading bootstrap", async () => {
    const user = userEvent.setup();
    const onAfterAuthenticated = vi.fn();

    reAuthenticateFanMock.mockResolvedValue(undefined);

    render(
      <ViewerSessionProvider hasSession>
        <CurrentViewerProvider
          currentViewer={{
            activeMode: "fan",
            canAccessCreatorMode: false,
            id: "viewer_123",
          }}
        >
          <FanAuthDialogProvider>
            <FanAuthDialogTrigger
              afterAuthenticatedHref="/fan"
              initialMode="re-auth"
              onAfterAuthenticated={onAfterAuthenticated}
            />
          </FanAuthDialogProvider>
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    await user.click(screen.getByRole("button", { name: "open auth dialog" }));
    await user.type(screen.getByPlaceholderText("現在のパスワード"), "VeryStrongPass123!");
    await user.click(screen.getByRole("button", { name: "認証を続ける" }));

    await waitFor(() => {
      expect(reAuthenticateFanMock).toHaveBeenCalledWith({
        password: "VeryStrongPass123!",
      });
      expect(onAfterAuthenticated).toHaveBeenCalledTimes(1);
    });

    expect(getCurrentViewerBootstrapMock).not.toHaveBeenCalled();
    expect(mockedRouter.push).not.toHaveBeenCalled();
    expect(mockedRouter.refresh).not.toHaveBeenCalled();
    expect(screen.queryByRole("dialog", { name: "認証を確認してください" })).not.toBeInTheDocument();
  });

  it("re-reads bootstrap when re-auth falls back to sign-in", async () => {
    const user = userEvent.setup();

    reAuthenticateFanMock.mockRejectedValueOnce(
      new FanAuthApiError("auth_required", "session is missing"),
    );
    signInFanMock.mockResolvedValue(undefined);
    getCurrentViewerBootstrapMock.mockResolvedValue({
      activeMode: "fan",
      canAccessCreatorMode: false,
      id: "viewer_123",
    });

    render(
      <ViewerSessionProvider hasSession>
        <CurrentViewerProvider currentViewer={null}>
          <FanAuthDialogProvider>
            <FanAuthDialogTrigger initialMode="re-auth" />
          </FanAuthDialogProvider>
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    await user.click(screen.getByRole("button", { name: "open auth dialog" }));
    await user.type(screen.getByPlaceholderText("現在のパスワード"), "VeryStrongPass123!");
    await user.click(screen.getByRole("button", { name: "認証を続ける" }));

    expect(
      await screen.findByRole("alert"),
    ).toHaveTextContent(
      "セッションが確認できませんでした。もう一度ログインしてください。",
    );
    expect(screen.getByRole("button", { name: "閉じる" })).toBeEnabled();

    await user.type(screen.getByRole("textbox", { name: "Email" }), "fan@example.com");
    await user.click(screen.getByRole("button", { name: "サインインを続ける" }));

    await waitFor(() => {
      expect(signInFanMock).toHaveBeenCalledWith({
        email: "fan@example.com",
        password: "VeryStrongPass123!",
      });
      expect(getCurrentViewerBootstrapMock).toHaveBeenCalledWith({
        credentials: "include",
      });
      expect(mockedRouter.push).toHaveBeenCalledWith("/fan");
    });

    expect(screen.queryByRole("dialog", { name: "続けるには認証が必要です" })).not.toBeInTheDocument();
  });
});

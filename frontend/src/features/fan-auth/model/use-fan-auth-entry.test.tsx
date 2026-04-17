import userEvent from "@testing-library/user-event";
import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";

import {
  FanAuthApiError,
  useFanAuthEntry,
  type FanAuthMode,
} from "@/features/fan-auth";

const emptyAvatarField = {
  canClear: false,
  fileName: null,
  inputAccept: "image/jpeg,image/png,image/webp",
  isError: false,
  kind: "empty" as const,
  message: "未設定でも登録できます。",
  previewUrl: null,
};

const mockedRouter = vi.hoisted(() => ({
  back: vi.fn(),
  forward: vi.fn(),
  prefetch: vi.fn(),
  push: vi.fn(),
  refresh: vi.fn(),
  replace: vi.fn(),
}));

const confirmFanPasswordResetMock = vi.hoisted(() => vi.fn());
const confirmFanSignUpMock = vi.hoisted(() => vi.fn());
const clearAvatarSelectionMock = vi.hoisted(() => vi.fn());
const getAvatarSubmissionErrorMock = vi.hoisted(() => vi.fn(() => null));
const getProfileValidationErrorMock = vi.hoisted(() => vi.fn((): string | null => null));
const reAuthenticateFanMock = vi.hoisted(() => vi.fn());
const resetDraftMock = vi.hoisted(() => vi.fn());
const selectAvatarFileMock = vi.hoisted(() => vi.fn());
const setDisplayNameMock = vi.hoisted(() => vi.fn());
const setHandleMock = vi.hoisted(() => vi.fn());
const signInFanMock = vi.hoisted(() => vi.fn());
const signUpFanMock = vi.hoisted(() => vi.fn());
const startFanPasswordResetMock = vi.hoisted(() => vi.fn());
const updateViewerProfileMock = vi.hoisted(() => vi.fn());
const uploadAvatarIfNeededMock = vi.hoisted(() => vi.fn());

vi.mock("next/navigation", async () => {
  const actual = await vi.importActual<typeof import("next/navigation")>("next/navigation");

  return {
    ...actual,
    useRouter: () => mockedRouter,
  };
});

vi.mock("../api/request-fan-auth", () => ({
  confirmFanPasswordReset: confirmFanPasswordResetMock,
  confirmFanSignUp: confirmFanSignUpMock,
  reAuthenticateFan: reAuthenticateFanMock,
  signInFan: signInFanMock,
  signUpFan: signUpFanMock,
  startFanPasswordReset: startFanPasswordResetMock,
}));

vi.mock("@/features/viewer-profile", () => ({
  updateViewerProfile: updateViewerProfileMock,
  useViewerProfileDraft: () => ({
    avatar: emptyAvatarField,
    avatarInputKey: 0,
    clearAvatarSelection: clearAvatarSelectionMock,
    displayName: "Mina",
    getAvatarSubmissionError: getAvatarSubmissionErrorMock,
    getProfileValidationError: getProfileValidationErrorMock,
    handle: "@mina",
    resetDraft: resetDraftMock,
    selectAvatarFile: selectAvatarFileMock,
    setDisplayName: setDisplayNameMock,
    setHandle: setHandleMock,
    uploadAvatarIfNeeded: uploadAvatarIfNeededMock,
  }),
}));

function FanAuthEntryConsumer(props: {
  initialMode?: FanAuthMode;
  onAuthenticated?: (
    options?: {
      afterViewerSynced?: () => Promise<string | null> | string | null;
      authenticatedMode?: FanAuthMode;
    },
  ) => Promise<string | null> | string | null;
}) {
  const {
    canResend,
    confirmationCode,
    deliveryDestinationHint,
    email,
    errorMessage,
    infoMessage,
    mode,
    newPassword,
    password,
    resend,
    setConfirmationCode,
    setEmail,
    setMode,
    setNewPassword,
    setPassword,
    submit,
  } = useFanAuthEntry(props);

  return (
    <div>
      <p data-testid="mode">{mode}</p>
      <p data-testid="can-resend">{canResend ? "yes" : "no"}</p>
      <p data-testid="email">{email}</p>
      <p data-testid="password">{password}</p>
      <p data-testid="confirmation-code">{confirmationCode}</p>
      <p data-testid="new-password">{newPassword}</p>
      <p data-testid="delivery-hint">{deliveryDestinationHint ?? ""}</p>
      <p data-testid="info-message">{infoMessage ?? ""}</p>
      <button onClick={() => setMode("sign-up")} type="button">
        mode-sign-up
      </button>
      <button onClick={() => setMode("sign-in")} type="button">
        mode-sign-in
      </button>
      <button onClick={() => setMode("password-reset-request")} type="button">
        mode-password-reset
      </button>
      <button onClick={() => setEmail("fan@example.com")} type="button">
        set-email
      </button>
      <button onClick={() => setEmail("other@example.com")} type="button">
        set-other-email
      </button>
      <button onClick={() => setPassword("VeryStrongPass123!")} type="button">
        set-password
      </button>
      <button onClick={() => setPassword("AnotherStrongPass456!")} type="button">
        set-other-password
      </button>
      <button onClick={() => setConfirmationCode("123456")} type="button">
        set-code
      </button>
      <button onClick={() => setNewPassword("EvenStrongerPass123!")} type="button">
        set-new-password
      </button>
      <button onClick={() => void resend()} type="button">
        resend
      </button>
      <button onClick={() => void submit()} type="button">
        submit
      </button>
      {errorMessage ? <p role="alert">{errorMessage}</p> : null}
    </div>
  );
}

describe("useFanAuthEntry", () => {
  beforeEach(() => {
    mockedRouter.back.mockReset();
    mockedRouter.forward.mockReset();
    mockedRouter.prefetch.mockReset();
    mockedRouter.push.mockReset();
    mockedRouter.refresh.mockReset();
    mockedRouter.replace.mockReset();
    confirmFanPasswordResetMock.mockReset();
    confirmFanSignUpMock.mockReset();
    clearAvatarSelectionMock.mockReset();
    getAvatarSubmissionErrorMock.mockReset();
    getProfileValidationErrorMock.mockReset();
    reAuthenticateFanMock.mockReset();
    resetDraftMock.mockReset();
    selectAvatarFileMock.mockReset();
    setDisplayNameMock.mockReset();
    setHandleMock.mockReset();
    signInFanMock.mockReset();
    signUpFanMock.mockReset();
    startFanPasswordResetMock.mockReset();
    updateViewerProfileMock.mockReset();
    uploadAvatarIfNeededMock.mockReset();

    getAvatarSubmissionErrorMock.mockReturnValue(null);
    getProfileValidationErrorMock.mockReturnValue(null);
    uploadAvatarIfNeededMock.mockResolvedValue(undefined);
  });

  it("transitions from sign-up into confirmation mode and stores the delivery hint", async () => {
    const user = userEvent.setup();

    signUpFanMock.mockResolvedValue({
      deliveryDestinationHint: "f***@example.com",
      nextStep: "confirm_sign_up",
    });

    render(<FanAuthEntryConsumer initialMode="sign-up" />);

    await user.click(screen.getByRole("button", { name: "set-email" }));
    await user.click(screen.getByRole("button", { name: "set-password" }));
    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(signUpFanMock).toHaveBeenCalledWith({
        displayName: "Mina",
        email: "fan@example.com",
        handle: "@mina",
        password: "VeryStrongPass123!",
      });
    });

    expect(screen.getByTestId("mode")).toHaveTextContent("confirm-sign-up");
    expect(screen.getByTestId("can-resend")).toHaveTextContent("yes");
    expect(screen.getByTestId("delivery-hint")).toHaveTextContent("f***@example.com");
    expect(screen.getByTestId("info-message")).toHaveTextContent(
      "確認コードを送信しました。メールを確認してください。",
    );
  });

  it("completes sign-up confirmation and runs avatar initialization after viewer sync", async () => {
    const user = userEvent.setup();
    const onAuthenticated = vi.fn(async (options?: {
      afterViewerSynced?: () => Promise<string | null> | string | null;
      authenticatedMode?: FanAuthMode;
    }) => {
      return options?.afterViewerSynced?.() ?? null;
    });

    signUpFanMock.mockResolvedValue({
      deliveryDestinationHint: "f***@example.com",
      nextStep: "confirm_sign_up",
    });
    confirmFanSignUpMock.mockResolvedValue(undefined);
    uploadAvatarIfNeededMock.mockResolvedValue("avatar-token");
    updateViewerProfileMock.mockResolvedValue(undefined);

    render(<FanAuthEntryConsumer initialMode="sign-up" onAuthenticated={onAuthenticated} />);

    await user.click(screen.getByRole("button", { name: "set-email" }));
    await user.click(screen.getByRole("button", { name: "set-password" }));
    await user.click(screen.getByRole("button", { name: "submit" }));
    await user.click(screen.getByRole("button", { name: "set-code" }));
    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(confirmFanSignUpMock).toHaveBeenCalledWith({
        confirmationCode: "123456",
        email: "fan@example.com",
      });
      expect(updateViewerProfileMock).toHaveBeenCalledWith({
        avatarUploadToken: "avatar-token",
        displayName: "Mina",
        handle: "@mina",
      });
      expect(onAuthenticated).toHaveBeenCalledTimes(1);
    });
  });

  it("moves sign-in into confirmation mode when the backend requires sign-up confirmation", async () => {
    const user = userEvent.setup();

    signInFanMock.mockRejectedValue(
      new FanAuthApiError("confirmation_required", "confirmation is required"),
    );

    render(<FanAuthEntryConsumer />);

    await user.click(screen.getByRole("button", { name: "set-email" }));
    await user.click(screen.getByRole("button", { name: "set-password" }));
    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(screen.getByTestId("mode")).toHaveTextContent("confirm-sign-up");
      expect(screen.getByTestId("can-resend")).toHaveTextContent("no");
      expect(screen.getByTestId("info-message")).toHaveTextContent(
        "確認コードを入力して登録を完了してください。",
      );
    });
    expect(resetDraftMock).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: "resend" }));

    await waitFor(() => {
      expect(screen.getByTestId("mode")).toHaveTextContent("sign-up");
      expect(screen.getByTestId("info-message")).toHaveTextContent(
        "確認コードを再送するには登録情報を入力し直してください。",
      );
    });
    expect(screen.getByTestId("password")).toHaveTextContent("");
    expect(signUpFanMock).not.toHaveBeenCalled();
  });

  it("resends confirmation when sign-in confirmation recovery keeps the accepted sign-up draft for the same email", async () => {
    const user = userEvent.setup();

    signInFanMock.mockRejectedValue(
      new FanAuthApiError("confirmation_required", "confirmation is required"),
    );
    signUpFanMock.mockResolvedValue({
      deliveryDestinationHint: "f***@example.com",
      nextStep: "confirm_sign_up",
    });

    render(<FanAuthEntryConsumer initialMode="sign-up" />);

    await user.click(screen.getByRole("button", { name: "set-email" }));
    await user.click(screen.getByRole("button", { name: "set-password" }));
    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(screen.getByTestId("mode")).toHaveTextContent("confirm-sign-up");
    });

    await user.click(screen.getByRole("button", { name: "mode-sign-in" }));
    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(screen.getByTestId("mode")).toHaveTextContent("confirm-sign-up");
      expect(screen.getByTestId("can-resend")).toHaveTextContent("yes");
    });

    await user.click(screen.getByRole("button", { name: "resend" }));

    await waitFor(() => {
      expect(signUpFanMock).toHaveBeenCalledWith({
        displayName: "Mina",
        email: "fan@example.com",
        handle: "@mina",
        password: "VeryStrongPass123!",
      });
      expect(screen.getByTestId("info-message")).toHaveTextContent(
        "確認コードを再送しました。メールを確認してください。",
      );
    });
  });

  it("drops the accepted sign-up draft when sign-in confirmation recovery targets a different email", async () => {
    const user = userEvent.setup();

    signInFanMock.mockRejectedValue(
      new FanAuthApiError("confirmation_required", "confirmation is required"),
    );
    signUpFanMock.mockResolvedValue({
      deliveryDestinationHint: "f***@example.com",
      nextStep: "confirm_sign_up",
    });

    render(<FanAuthEntryConsumer initialMode="sign-up" />);

    await user.click(screen.getByRole("button", { name: "set-email" }));
    await user.click(screen.getByRole("button", { name: "set-password" }));
    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(screen.getByTestId("mode")).toHaveTextContent("confirm-sign-up");
    });

    await user.click(screen.getByRole("button", { name: "mode-sign-in" }));
    await user.click(screen.getByRole("button", { name: "set-other-email" }));
    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(screen.getByTestId("mode")).toHaveTextContent("confirm-sign-up");
      expect(screen.getByTestId("can-resend")).toHaveTextContent("no");
      expect(screen.getByTestId("delivery-hint")).toHaveTextContent("");
    });

    expect(resetDraftMock).not.toHaveBeenCalled();
  });

  it("drops the accepted sign-up draft when sign-in confirmation recovery changes the password", async () => {
    const user = userEvent.setup();

    signInFanMock.mockRejectedValue(
      new FanAuthApiError("confirmation_required", "confirmation is required"),
    );
    signUpFanMock.mockResolvedValue({
      deliveryDestinationHint: "f***@example.com",
      nextStep: "confirm_sign_up",
    });

    render(<FanAuthEntryConsumer initialMode="sign-up" />);

    await user.click(screen.getByRole("button", { name: "set-email" }));
    await user.click(screen.getByRole("button", { name: "set-password" }));
    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(screen.getByTestId("mode")).toHaveTextContent("confirm-sign-up");
    });

    await user.click(screen.getByRole("button", { name: "mode-sign-in" }));
    await user.click(screen.getByRole("button", { name: "set-other-password" }));
    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(screen.getByTestId("mode")).toHaveTextContent("confirm-sign-up");
      expect(screen.getByTestId("can-resend")).toHaveTextContent("no");
    });

    expect(resetDraftMock).not.toHaveBeenCalled();
  });

  it("clears password when confirm-sign-up returns to sign-up without a resend-safe draft", async () => {
    const user = userEvent.setup();

    signInFanMock.mockRejectedValue(
      new FanAuthApiError("confirmation_required", "confirmation is required"),
    );

    render(<FanAuthEntryConsumer />);

    await user.click(screen.getByRole("button", { name: "set-email" }));
    await user.click(screen.getByRole("button", { name: "set-password" }));
    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(screen.getByTestId("mode")).toHaveTextContent("confirm-sign-up");
      expect(screen.getByTestId("can-resend")).toHaveTextContent("no");
    });

    await user.click(screen.getByRole("button", { name: "mode-sign-up" }));

    expect(screen.getByTestId("mode")).toHaveTextContent("sign-up");
    expect(screen.getByTestId("password")).toHaveTextContent("");
  });

  it("requires password re-entry when resend-safe sign-up draft becomes invalid", async () => {
    const user = userEvent.setup();

    signUpFanMock.mockResolvedValue({
      deliveryDestinationHint: "f***@example.com",
      nextStep: "confirm_sign_up",
    });

    render(<FanAuthEntryConsumer initialMode="sign-up" />);

    await user.click(screen.getByRole("button", { name: "set-email" }));
    await user.click(screen.getByRole("button", { name: "set-password" }));
    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(screen.getByTestId("mode")).toHaveTextContent("confirm-sign-up");
    });

    getProfileValidationErrorMock.mockReturnValue("display name is invalid");

    await user.click(screen.getByRole("button", { name: "resend" }));

    await waitFor(() => {
      expect(screen.getByTestId("mode")).toHaveTextContent("sign-up");
      expect(screen.getByTestId("password")).toHaveTextContent("");
      expect(screen.getByTestId("info-message")).toHaveTextContent(
        "確認コードを再送するには登録情報を入力し直してください。",
      );
    });

    expect(resetDraftMock).not.toHaveBeenCalled();
    expect(signUpFanMock).toHaveBeenCalledTimes(1);
  });

  it("does not re-confirm sign-up when post-confirm sync fails and the user retries", async () => {
    const user = userEvent.setup();
    const onAuthenticated = vi.fn(async (options?: {
      afterViewerSynced?: () => Promise<string | null> | string | null;
      authenticatedMode?: FanAuthMode;
    }) => {
      return options?.afterViewerSynced?.() ?? null;
    });

    signUpFanMock.mockResolvedValue({
      deliveryDestinationHint: "f***@example.com",
      nextStep: "confirm_sign_up",
    });
    confirmFanSignUpMock.mockResolvedValue(undefined);
    uploadAvatarIfNeededMock.mockResolvedValue("avatar-token");
    updateViewerProfileMock.mockRejectedValueOnce(new Error("boom")).mockResolvedValueOnce(undefined);

    render(<FanAuthEntryConsumer initialMode="sign-up" onAuthenticated={onAuthenticated} />);

    await user.click(screen.getByRole("button", { name: "set-email" }));
    await user.click(screen.getByRole("button", { name: "set-password" }));
    await user.click(screen.getByRole("button", { name: "submit" }));
    await user.click(screen.getByRole("button", { name: "set-code" }));
    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(confirmFanSignUpMock).toHaveBeenCalledTimes(1);
      expect(updateViewerProfileMock).toHaveBeenCalledTimes(1);
      expect(screen.getByRole("alert")).toHaveTextContent(
        "avatar の初期化に失敗しました。少し時間を置いてから再度お試しください。",
      );
    });

    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(confirmFanSignUpMock).toHaveBeenCalledTimes(1);
      expect(updateViewerProfileMock).toHaveBeenCalledTimes(2);
      expect(onAuthenticated).toHaveBeenCalledTimes(2);
    });
  });

  it("returns password-reset confirmation back to sign-in while preserving only the email", async () => {
    const user = userEvent.setup();

    startFanPasswordResetMock.mockResolvedValue({
      deliveryDestinationHint: "f***@example.com",
      nextStep: "confirm_password_reset",
    });
    confirmFanPasswordResetMock.mockResolvedValue(undefined);

    render(<FanAuthEntryConsumer />);

    await user.click(screen.getByRole("button", { name: "mode-password-reset" }));
    await user.click(screen.getByRole("button", { name: "set-email" }));
    await user.click(screen.getByRole("button", { name: "submit" }));
    await user.click(screen.getByRole("button", { name: "set-code" }));
    await user.click(screen.getByRole("button", { name: "set-new-password" }));
    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(confirmFanPasswordResetMock).toHaveBeenCalledWith({
        confirmationCode: "123456",
        email: "fan@example.com",
        newPassword: "EvenStrongerPass123!",
      });
    });

    expect(screen.getByTestId("mode")).toHaveTextContent("sign-in");
    expect(screen.getByTestId("password")).toHaveTextContent("");
    expect(screen.getByTestId("info-message")).toHaveTextContent(
      "パスワードを更新しました。サインインを続けてください。",
    );
  });

  it("falls back to sign-in when re-auth loses the current session", async () => {
    const user = userEvent.setup();

    reAuthenticateFanMock.mockRejectedValue(
      new FanAuthApiError("auth_required", "current session is missing"),
    );

    render(<FanAuthEntryConsumer initialMode="re-auth" />);

    await user.click(screen.getByRole("button", { name: "set-password" }));
    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(screen.getByTestId("mode")).toHaveTextContent("sign-in");
      expect(screen.getByRole("alert")).toHaveTextContent(
        "セッションが確認できませんでした。もう一度ログインしてください。",
      );
    });
  });
});

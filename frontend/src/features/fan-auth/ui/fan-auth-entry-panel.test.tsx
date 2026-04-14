import userEvent from "@testing-library/user-event";
import { render, screen, waitFor } from "@testing-library/react";

import {
  FanAuthApiError,
  FanAuthEntryPanel,
  useFanAuthEntry,
} from "@/features/fan-auth";

const authenticateFanWithEmailMock = vi.hoisted(() => vi.fn());

vi.mock("@/features/fan-auth/api/request-fan-auth", () => ({
  authenticateFanWithEmail: authenticateFanWithEmailMock,
}));

function FanAuthEntryPanelHarness(props: {
  onAuthenticated?: () => Promise<string | null> | string | null;
}) {
  const {
    avatar,
    avatarInputKey,
    clearAvatarSelection,
    displayName,
    email,
    errorMessage,
    handle,
    isSubmitting,
    mode,
    selectAvatarFile,
    setDisplayName,
    setEmail,
    setHandle,
    submit,
    switchMode,
  } = useFanAuthEntry(props);

  return (
    <FanAuthEntryPanel
      avatar={avatar}
      avatarInputKey={avatarInputKey}
      clearAvatarSelection={clearAvatarSelection}
      displayName={displayName}
      email={email}
      errorMessage={errorMessage}
      handle={handle}
      isSubmitting={isSubmitting}
      mode={mode}
      onAvatarSelect={selectAvatarFile}
      onDisplayNameChange={setDisplayName}
      onEmailChange={setEmail}
      onHandleChange={setHandle}
      onModeSwitch={switchMode}
      onSubmit={submit}
    />
  );
}

describe("FanAuthEntryPanel", () => {
  beforeEach(() => {
    authenticateFanWithEmailMock.mockReset();
  });

  it("starts in sign-in mode and preserves email when switching to sign-up", async () => {
    const user = userEvent.setup();

    render(<FanAuthEntryPanelHarness />);

    const emailInput = screen.getByRole("textbox", { name: "Email" });
    await user.type(emailInput, "fan@example.com");

    expect(screen.getByRole("button", { name: "サインインを続ける" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "サインアップへ" }));

    expect(screen.getByRole("button", { name: "新規登録を続ける" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Email" })).toHaveValue("fan@example.com");
    expect(screen.getByRole("textbox", { name: "Display name" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Handle" })).toBeInTheDocument();
  });

  it("runs the shared auth flow and notifies the caller on success", async () => {
    const user = userEvent.setup();
    const onAuthenticated = vi.fn();

    authenticateFanWithEmailMock.mockResolvedValue(undefined);

    render(<FanAuthEntryPanelHarness onAuthenticated={onAuthenticated} />);

    await user.type(screen.getByRole("textbox", { name: "Email" }), "fan@example.com");
    await user.click(screen.getByRole("button", { name: "サインインを続ける" }));

    await waitFor(() => {
      expect(authenticateFanWithEmailMock).toHaveBeenCalledWith({
        email: "fan@example.com",
        mode: "sign-in",
      });
      expect(onAuthenticated).toHaveBeenCalledTimes(1);
    });
  });

  it("passes shared profile fields when signing up", async () => {
    const user = userEvent.setup();

    authenticateFanWithEmailMock.mockResolvedValue(undefined);

    render(<FanAuthEntryPanelHarness />);

    await user.click(screen.getByRole("button", { name: "サインアップへ" }));
    await user.type(screen.getByRole("textbox", { name: "Email" }), "fan@example.com");
    await user.type(screen.getByRole("textbox", { name: "Display name" }), "Mina");
    await user.type(screen.getByRole("textbox", { name: "Handle" }), "@mina");
    await user.click(screen.getByRole("button", { name: "新規登録を続ける" }));

    await waitFor(() => {
      expect(authenticateFanWithEmailMock).toHaveBeenCalledWith({
        displayName: "Mina",
        email: "fan@example.com",
        handle: "@mina",
        mode: "sign-up",
      });
    });
  });

  it("renders mapped contract errors inside the panel", async () => {
    const user = userEvent.setup();

    authenticateFanWithEmailMock.mockRejectedValue(
      new FanAuthApiError("email_not_found", "email was not found"),
    );

    render(<FanAuthEntryPanelHarness />);

    await user.type(screen.getByRole("textbox", { name: "Email" }), "fan@example.com");
    await user.click(screen.getByRole("button", { name: "サインインを続ける" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "このメールアドレスのアカウントが見つかりません。サインアップに切り替えてください。",
    );
  });
});

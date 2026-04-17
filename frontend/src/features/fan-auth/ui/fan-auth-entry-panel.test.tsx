import userEvent from "@testing-library/user-event";
import { render, screen } from "@testing-library/react";
import { useState } from "react";

import {
  FanAuthEntryPanel,
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

function FanAuthEntryPanelHarness({
  initialMode = "sign-in",
  onResend = vi.fn(),
}: {
  initialMode?: FanAuthMode;
  onResend?: () => void | Promise<void>;
}) {
  const [confirmationCode, setConfirmationCode] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [email, setEmail] = useState("");
  const [mode, setMode] = useState<FanAuthMode>(initialMode);
  const [newPassword, setNewPassword] = useState("");
  const [password, setPassword] = useState("");

  return (
    <FanAuthEntryPanel
      avatar={emptyAvatarField}
      avatarInputKey={0}
      canResend
      clearAvatarSelection={vi.fn()}
      confirmationCode={confirmationCode}
      deliveryDestinationHint="f***@example.com"
      displayName={displayName}
      email={email}
      errorMessage={null}
      handle="@mina"
      hasConfirmedSignUp={false}
      infoMessage={null}
      isSubmitting={false}
      mode={mode}
      newPassword={newPassword}
      onAvatarSelect={vi.fn()}
      onConfirmationCodeChange={setConfirmationCode}
      onDisplayNameChange={setDisplayName}
      onEmailChange={setEmail}
      onHandleChange={vi.fn()}
      onModeChange={setMode}
      onNewPasswordChange={setNewPassword}
      onPasswordChange={setPassword}
      onResend={onResend}
      onSubmit={vi.fn()}
      password={password}
    />
  );
}

describe("FanAuthEntryPanel", () => {
  it("preserves email when switching from sign-in to sign-up", async () => {
    const user = userEvent.setup();

    render(<FanAuthEntryPanelHarness />);

    await user.type(screen.getByRole("textbox", { name: "Email" }), "fan@example.com");
    await user.click(screen.getByRole("button", { name: "サインアップへ" }));

    expect(screen.getByRole("button", { name: "確認コードを送る" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Email" })).toHaveValue("fan@example.com");
    expect(screen.getByRole("textbox", { name: "Display name" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Handle" })).toBeInTheDocument();
  });

  it("renders confirmation controls and allows resending sign-up codes", async () => {
    const user = userEvent.setup();
    const onResend = vi.fn();

    render(<FanAuthEntryPanelHarness initialMode="confirm-sign-up" onResend={onResend} />);

    expect(screen.getByRole("textbox", { name: "Confirmation code" })).toBeInTheDocument();
    expect(screen.getByText("確認コードを f***@example.com に送りました。メールを確認してください。")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "コードを再送" }));

    expect(onResend).toHaveBeenCalledTimes(1);
  });

  it("renders the new password field during password reset confirmation", () => {
    render(<FanAuthEntryPanelHarness initialMode="confirm-password-reset" />);

    expect(screen.getByRole("textbox", { name: "Confirmation code" })).toBeInTheDocument();
    expect(screen.getByLabelText("New password")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "パスワードを更新する" })).toBeInTheDocument();
  });

  it("hides the email field during re-auth and explains the action", () => {
    render(<FanAuthEntryPanelHarness initialMode="re-auth" />);

    expect(screen.queryByRole("textbox", { name: "Email" })).not.toBeInTheDocument();
    expect(screen.getByLabelText("Password")).toBeInTheDocument();
    expect(screen.getByText("現在の fan session を維持したまま、必要な操作だけを続行します。")).toBeInTheDocument();
  });
});

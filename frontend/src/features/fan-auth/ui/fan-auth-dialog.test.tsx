import userEvent from "@testing-library/user-event";
import { fireEvent, render, screen } from "@testing-library/react";

import { FanAuthDialog } from "@/features/fan-auth";

const signInFanMock = vi.hoisted(() => vi.fn());
const startFanPasswordResetMock = vi.hoisted(() => vi.fn());

vi.mock("@/features/fan-auth/api/request-fan-auth", () => ({
  confirmFanPasswordReset: vi.fn(),
  confirmFanSignUp: vi.fn(),
  reAuthenticateFan: vi.fn(),
  signInFan: signInFanMock,
  signUpFan: vi.fn(),
  startFanPasswordReset: startFanPasswordResetMock,
}));

describe("FanAuthDialog", () => {
  beforeEach(() => {
    signInFanMock.mockReset();
    startFanPasswordResetMock.mockReset();
  });

  it("renders a viewport-bounded scrollable shell for long sign-up content", () => {
    render(
      <FanAuthDialog
        onAuthenticated={vi.fn()}
        onOpenChange={vi.fn()}
        open
        sessionKey={0}
      />,
    );

    const dialog = screen.getByRole("dialog");
    const scrollShell = dialog.querySelector("div.max-h-\\[90svh\\]");

    expect(dialog).toHaveClass(
      "inset-x-0",
      "bottom-0",
      "left-1/2",
      "max-w-[408px]",
      "-translate-x-1/2",
    );
    expect(scrollShell).not.toBeNull();
    expect(scrollShell).toHaveClass(
      "max-h-[90svh]",
      "overflow-y-auto",
      "overscroll-contain",
      "rounded-t-[32px]",
    );
  });

  it("disables closing affordances while auth submission is in flight", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();

    signInFanMock.mockImplementation(
      () =>
        new Promise(() => {
          // keep pending to verify the dialog stays non-dismissible
        }),
    );

    render(
      <FanAuthDialog
        onAuthenticated={vi.fn()}
        onOpenChange={onOpenChange}
        open
        sessionKey={0}
      />,
    );

    await user.type(screen.getByRole("textbox", { name: "Email" }), "fan@example.com");
    await user.type(screen.getByLabelText("Password"), "VeryStrongPass123!");
    fireEvent.click(screen.getByRole("button", { name: "サインインを続ける" }));
    fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });

    expect(screen.getByRole("button", { name: "閉じる" })).toBeDisabled();
    expect(onOpenChange).not.toHaveBeenCalled();
  });

  it("disables closing affordances while confirmation resend is in flight", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();

    startFanPasswordResetMock
      .mockResolvedValueOnce({
        deliveryDestinationHint: "f***@example.com",
        nextStep: "confirm_password_reset",
      })
      .mockImplementationOnce(
        () =>
          new Promise(() => {
            // keep pending to verify the dialog stays non-dismissible
          }),
      );

    render(
      <FanAuthDialog
        initialMode="password-reset-request"
        onAuthenticated={vi.fn()}
        onOpenChange={onOpenChange}
        open
        sessionKey={0}
      />,
    );

    await user.type(screen.getByRole("textbox", { name: "Email" }), "fan@example.com");
    fireEvent.click(screen.getByRole("button", { name: "確認コードを送る" }));

    expect(await screen.findByRole("button", { name: "コードを再送" })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "コードを再送" }));
    fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });

    expect(screen.getByRole("button", { name: "閉じる" })).toBeDisabled();
    expect(onOpenChange).not.toHaveBeenCalled();
  });
});

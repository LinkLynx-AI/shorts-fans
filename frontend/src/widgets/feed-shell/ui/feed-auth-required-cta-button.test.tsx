import userEvent from "@testing-library/user-event";
import { render, screen } from "@testing-library/react";

import { useFanAuthDialogControls } from "@/features/fan-auth";

import { FeedAuthRequiredCtaButton } from "./feed-auth-required-cta-button";

vi.mock("@/features/fan-auth", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/features/fan-auth")>();

  return {
    ...actual,
    useFanAuthDialogControls: vi.fn(),
  };
});

const mockedUseFanAuthDialogControls = vi.mocked(useFanAuthDialogControls);
const openFanAuthDialog = vi.fn();

describe("FeedAuthRequiredCtaButton", () => {
  beforeEach(() => {
    openFanAuthDialog.mockReset();
    mockedUseFanAuthDialogControls.mockReturnValue({
      closeFanAuthDialog: vi.fn(),
      openFanAuthDialog,
    });
  });

  it("opens the shared auth dialog with default post-auth refresh behavior", async () => {
    const user = userEvent.setup();

    render(<FeedAuthRequiredCtaButton />);

    await user.click(screen.getByRole("button", { name: "ログインして続ける" }));

    expect(openFanAuthDialog).toHaveBeenCalledWith();
  });
});

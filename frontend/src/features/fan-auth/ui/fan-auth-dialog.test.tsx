import userEvent from "@testing-library/user-event";
import { render, screen } from "@testing-library/react";

import { FanAuthDialog } from "@/features/fan-auth";

const authenticateFanWithEmailMock = vi.hoisted(() => vi.fn());

vi.mock("@/features/fan-auth/api/request-fan-auth", () => ({
  authenticateFanWithEmail: authenticateFanWithEmailMock,
}));

describe("FanAuthDialog", () => {
  beforeEach(() => {
    authenticateFanWithEmailMock.mockReset();
  });

  it("disables closing affordances while auth submission is in flight", async () => {
    const user = userEvent.setup();

    authenticateFanWithEmailMock.mockImplementation(
      () =>
        new Promise(() => {
          // keep pending to verify the dialog stays non-dismissible
        }),
    );

    render(<FanAuthDialog onAuthenticated={vi.fn()} onOpenChange={vi.fn()} open />);

    await user.type(screen.getByRole("textbox", { name: "Email" }), "fan@example.com");
    await user.click(screen.getByRole("button", { name: "サインインを続ける" }));

    expect(screen.getByRole("button", { name: "閉じる" })).toBeDisabled();
  });
});

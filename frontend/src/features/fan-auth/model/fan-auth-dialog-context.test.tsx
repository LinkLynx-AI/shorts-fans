import userEvent from "@testing-library/user-event";
import { render, screen, waitFor } from "@testing-library/react";

import {
  CurrentViewerProvider,
  ViewerSessionProvider,
} from "@/entities/viewer";
import {
  FanAuthDialogProvider,
  useFanAuthDialog,
} from "@/features/fan-auth";

const authenticateFanWithEmailMock = vi.hoisted(() => vi.fn());
const getCurrentViewerBootstrapMock = vi.hoisted(() => vi.fn());

vi.mock("@/features/fan-auth/api/request-fan-auth", () => ({
  authenticateFanWithEmail: authenticateFanWithEmailMock,
}));

vi.mock("@/entities/viewer", async () => {
  const actual = await vi.importActual<typeof import("@/entities/viewer")>("@/entities/viewer");

  return {
    ...actual,
    getCurrentViewerBootstrap: getCurrentViewerBootstrapMock,
  };
});

function FanAuthDialogTrigger() {
  const { openFanAuthDialog } = useFanAuthDialog();

  return (
    <button
      onClick={() =>
        openFanAuthDialog({
          afterAuthenticatedHref: "/fan",
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
    authenticateFanWithEmailMock.mockReset();
    getCurrentViewerBootstrapMock.mockReset();
  });

  it("keeps the modal open and shows a recoverable error when bootstrap resolves null after auth success", async () => {
    const user = userEvent.setup();

    authenticateFanWithEmailMock.mockResolvedValue(undefined);
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
    expect(screen.getByRole("dialog", { name: "続けるにはログインが必要です" })).toBeInTheDocument();
  });
});

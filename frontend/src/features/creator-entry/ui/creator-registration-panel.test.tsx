import userEvent from "@testing-library/user-event";
import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";

import {
  CurrentViewerProvider,
  ViewerSessionProvider,
  getCurrentViewerBootstrap,
} from "@/entities/viewer";
import { ApiError } from "@/shared/api";
import {
  CreatorRegistrationPanel,
  registerCreator,
} from "@/features/creator-entry";

const mockedRouter = vi.hoisted(() => ({
  back: vi.fn(),
  forward: vi.fn(),
  prefetch: vi.fn(),
  push: vi.fn(),
  refresh: vi.fn(),
  replace: vi.fn(),
}));

vi.mock("next/navigation", async () => {
  const actual = await vi.importActual<typeof import("next/navigation")>("next/navigation");

  return {
    ...actual,
    useRouter: () => mockedRouter,
  };
});

vi.mock("@/entities/viewer", async () => {
  const actual = await vi.importActual<typeof import("@/entities/viewer")>("@/entities/viewer");

  return {
    ...actual,
    getCurrentViewerBootstrap: vi.fn(),
  };
});

vi.mock("@/features/creator-entry/api/register-creator", () => ({
  registerCreator: vi.fn(),
}));

function renderPanel() {
  return render(
    <ViewerSessionProvider hasSession>
      <CurrentViewerProvider
        currentViewer={{
          activeMode: "fan",
          canAccessCreatorMode: false,
          id: "viewer_123",
        }}
      >
        <CreatorRegistrationPanel />
      </CurrentViewerProvider>
    </ViewerSessionProvider>,
  );
}

describe("CreatorRegistrationPanel", () => {
  beforeEach(() => {
    mockedRouter.back.mockReset();
    mockedRouter.forward.mockReset();
    mockedRouter.prefetch.mockReset();
    mockedRouter.push.mockReset();
    mockedRouter.refresh.mockReset();
    mockedRouter.replace.mockReset();
    vi.mocked(getCurrentViewerBootstrap).mockReset();
    vi.mocked(registerCreator).mockReset();
  });

  it("renders a confirm-only registration surface", () => {
    renderPanel();

    expect(screen.getByRole("heading", { name: "Creator登録を始める" })).toBeInTheDocument();
    expect(screen.getByText(/signup と profile settings で fan \/ creator 共通/)).toBeInTheDocument();
    expect(screen.queryByRole("textbox", { name: "Display name" })).not.toBeInTheDocument();
    expect(screen.queryByRole("textbox", { name: "Handle" })).not.toBeInTheDocument();
  });

  it("submits registration and routes to the success screen after bootstrap sync", async () => {
    const user = userEvent.setup();

    vi.mocked(registerCreator).mockResolvedValue(undefined);
    vi.mocked(getCurrentViewerBootstrap).mockResolvedValue({
      activeMode: "fan",
      canAccessCreatorMode: true,
      id: "viewer_123",
    });

    renderPanel();

    await user.click(screen.getByRole("button", { name: "申し込む" }));

    await waitFor(() => {
      expect(registerCreator).toHaveBeenCalledWith();
      expect(getCurrentViewerBootstrap).toHaveBeenCalledWith({
        credentials: "include",
      });
      expect(mockedRouter.push).toHaveBeenCalledWith("/fan/creator/success");
    });
  });

  it("shows the mapped API error returned by registration", async () => {
    const user = userEvent.setup();

    vi.mocked(registerCreator).mockRejectedValue(
      new ApiError("API request failed with a non-success status.", {
        code: "http",
        details: JSON.stringify({
          error: {
            code: "handle_already_taken",
            message: "handle is already taken",
          },
          meta: {
            requestId: "req_handle_taken",
          },
        }),
        status: 409,
      }),
    );

    renderPanel();

    await user.click(screen.getByRole("button", { name: "申し込む" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "そのhandleは既に使われています。別のhandleを入力してください。",
    );
  });
});

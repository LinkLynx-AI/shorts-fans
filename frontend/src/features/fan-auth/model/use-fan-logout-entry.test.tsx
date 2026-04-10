import userEvent from "@testing-library/user-event";
import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";

import {
  CurrentViewerProvider,
  useCurrentViewer,
  useHasViewerSession,
  ViewerSessionProvider,
} from "@/entities/viewer";
import { ApiError } from "@/shared/api";

import { useFanLogoutEntry } from "@/features/fan-auth";
import { logoutFanSession } from "../api/logout-fan-session";

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

vi.mock("../api/logout-fan-session", () => ({
  logoutFanSession: vi.fn(),
}));

function createDeferredPromise<TResult = void>() {
  let resolvePromise: (value: TResult | PromiseLike<TResult>) => void = () => {};
  let rejectPromise: (reason?: unknown) => void = () => {};
  const promise = new Promise<TResult>((resolve, reject) => {
    resolvePromise = resolve;
    rejectPromise = reject;
  });

  return {
    promise,
    reject: rejectPromise,
    resolve: resolvePromise,
  };
}

function FanLogoutEntryConsumer() {
  const { clearError, errorMessage, isSubmitting, logout } = useFanLogoutEntry();
  const currentViewer = useCurrentViewer();
  const hasViewerSession = useHasViewerSession();

  return (
    <div>
      <button disabled={isSubmitting} onClick={() => void logout()} type="button">
        {isSubmitting ? "logging-out" : "logout"}
      </button>
      <button onClick={clearError} type="button">
        clear
      </button>
      <p>{currentViewer ? currentViewer.id : "viewer-missing"}</p>
      <p>{hasViewerSession ? "session-present" : "session-missing"}</p>
      {errorMessage ? <p role="alert">{errorMessage}</p> : null}
    </div>
  );
}

describe("useFanLogoutEntry", () => {
  beforeEach(() => {
    mockedRouter.back.mockReset();
    mockedRouter.forward.mockReset();
    mockedRouter.prefetch.mockReset();
    mockedRouter.push.mockReset();
    mockedRouter.refresh.mockReset();
    mockedRouter.replace.mockReset();
    vi.mocked(logoutFanSession).mockReset();
  });

  it("clears viewer state and redirects to the public feed when logout succeeds", async () => {
    const user = userEvent.setup();

    vi.mocked(logoutFanSession).mockResolvedValue(undefined);

    render(
      <ViewerSessionProvider hasSession>
        <CurrentViewerProvider
          currentViewer={{
            activeMode: "fan",
            canAccessCreatorMode: true,
            id: "viewer_123",
          }}
        >
          <FanLogoutEntryConsumer />
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    expect(screen.getByText("viewer_123")).toBeInTheDocument();
    expect(screen.getByText("session-present")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "logout" }));

    await waitFor(() => {
      expect(logoutFanSession).toHaveBeenCalledTimes(1);
      expect(mockedRouter.push).toHaveBeenCalledWith("/");
    });

    expect(screen.getByText("viewer-missing")).toBeInTheDocument();
    expect(screen.getByText("session-missing")).toBeInTheDocument();
  });

  it("keeps the menu retryable when logout fails", async () => {
    const user = userEvent.setup();

    vi.mocked(logoutFanSession).mockRejectedValue(
      new ApiError("API request failed before a response was received.", {
        code: "network",
      }),
    );

    render(
      <ViewerSessionProvider hasSession>
        <CurrentViewerProvider
          currentViewer={{
            activeMode: "fan",
            canAccessCreatorMode: false,
            id: "viewer_123",
          }}
        >
          <FanLogoutEntryConsumer />
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    await user.click(screen.getByRole("button", { name: "logout" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "ログアウトできませんでした。通信状態を確認してから再度お試しください。",
    );
    expect(screen.getByText("viewer_123")).toBeInTheDocument();
    expect(screen.getByText("session-present")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "clear" }));

    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("prevents duplicate logout submits while the request is pending", async () => {
    const user = userEvent.setup();
    const deferred = createDeferredPromise();

    vi.mocked(logoutFanSession).mockReturnValue(deferred.promise);

    render(
      <ViewerSessionProvider hasSession>
        <CurrentViewerProvider
          currentViewer={{
            activeMode: "fan",
            canAccessCreatorMode: false,
            id: "viewer_123",
          }}
        >
          <FanLogoutEntryConsumer />
        </CurrentViewerProvider>
      </ViewerSessionProvider>,
    );

    await user.click(screen.getByRole("button", { name: "logout" }));

    expect(screen.getByRole("button", { name: "logging-out" })).toBeDisabled();
    expect(logoutFanSession).toHaveBeenCalledTimes(1);

    deferred.resolve();

    await waitFor(() => {
      expect(mockedRouter.push).toHaveBeenCalledWith("/");
    });
  });
});

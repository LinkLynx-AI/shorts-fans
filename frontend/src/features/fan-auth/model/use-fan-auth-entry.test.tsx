import userEvent from "@testing-library/user-event";
import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";

import { useFanAuthEntry } from "@/features/fan-auth";

const mockedRouter = vi.hoisted(() => ({
  back: vi.fn(),
  forward: vi.fn(),
  prefetch: vi.fn(),
  push: vi.fn(),
  refresh: vi.fn(),
  replace: vi.fn(),
}));

const authenticateFanWithEmailMock = vi.hoisted(() => vi.fn());
const clearAvatarSelectionMock = vi.hoisted(() => vi.fn());
const getAvatarSubmissionErrorMock = vi.hoisted(() => vi.fn(() => null));
const getProfileValidationErrorMock = vi.hoisted(() => vi.fn(() => null));
const selectAvatarFileMock = vi.hoisted(() => vi.fn());
const setDisplayNameMock = vi.hoisted(() => vi.fn());
const setHandleMock = vi.hoisted(() => vi.fn());
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
  authenticateFanWithEmail: authenticateFanWithEmailMock,
}));

vi.mock("@/features/viewer-profile", () => ({
  updateViewerProfile: updateViewerProfileMock,
  useViewerProfileDraft: () => ({
    avatar: null,
    avatarInputKey: 0,
    clearAvatarSelection: clearAvatarSelectionMock,
    displayName: "Mina",
    getAvatarSubmissionError: getAvatarSubmissionErrorMock,
    getProfileValidationError: getProfileValidationErrorMock,
    handle: "@mina",
    selectAvatarFile: selectAvatarFileMock,
    setDisplayName: setDisplayNameMock,
    setHandle: setHandleMock,
    uploadAvatarIfNeeded: uploadAvatarIfNeededMock,
  }),
}));

function FanAuthEntryConsumer(props: {
  onAuthenticated?: () => Promise<string | null> | string | null;
}) {
  const {
    email,
    errorMessage,
    mode,
    setEmail,
    submit,
    switchMode,
  } = useFanAuthEntry(props);

  return (
    <div>
      <p>{mode}</p>
      <p>{email}</p>
      <button onClick={() => setEmail("fan@example.com")} type="button">
        set-email
      </button>
      <button onClick={switchMode} type="button">
        switch-mode
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
    authenticateFanWithEmailMock.mockReset();
    clearAvatarSelectionMock.mockReset();
    getAvatarSubmissionErrorMock.mockReset();
    getProfileValidationErrorMock.mockReset();
    selectAvatarFileMock.mockReset();
    setDisplayNameMock.mockReset();
    setHandleMock.mockReset();
    updateViewerProfileMock.mockReset();
    uploadAvatarIfNeededMock.mockReset();

    getAvatarSubmissionErrorMock.mockReturnValue(null);
    getProfileValidationErrorMock.mockReturnValue(null);
  });

  it("refreshes the route after sign-up even when optional avatar save fails", async () => {
    const user = userEvent.setup();

    authenticateFanWithEmailMock.mockResolvedValue(undefined);
    uploadAvatarIfNeededMock.mockResolvedValue("avatar-token");
    updateViewerProfileMock.mockRejectedValue(new Error("avatar save failed"));

    render(<FanAuthEntryConsumer />);

    await user.click(screen.getByRole("button", { name: "set-email" }));
    await user.click(screen.getByRole("button", { name: "switch-mode" }));
    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(authenticateFanWithEmailMock).toHaveBeenCalledWith({
        displayName: "Mina",
        email: "fan@example.com",
        handle: "@mina",
        mode: "sign-up",
      });
      expect(updateViewerProfileMock).toHaveBeenCalledWith({
        avatarUploadToken: "avatar-token",
        displayName: "Mina",
        handle: "@mina",
      });
      expect(mockedRouter.refresh).toHaveBeenCalledTimes(1);
    });
  });

  it("still notifies the caller after sign-up when optional avatar save fails", async () => {
    const user = userEvent.setup();
    const onAuthenticated = vi.fn().mockResolvedValue(null);

    authenticateFanWithEmailMock.mockResolvedValue(undefined);
    uploadAvatarIfNeededMock.mockResolvedValue("avatar-token");
    updateViewerProfileMock.mockRejectedValue(new Error("avatar save failed"));

    render(<FanAuthEntryConsumer onAuthenticated={onAuthenticated} />);

    await user.click(screen.getByRole("button", { name: "set-email" }));
    await user.click(screen.getByRole("button", { name: "switch-mode" }));
    await user.click(screen.getByRole("button", { name: "submit" }));

    await waitFor(() => {
      expect(onAuthenticated).toHaveBeenCalledTimes(1);
    });

    expect(mockedRouter.refresh).not.toHaveBeenCalled();
  });
});

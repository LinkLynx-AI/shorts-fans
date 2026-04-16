import userEvent from "@testing-library/user-event";
import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";

import { ApiError } from "@/shared/api";

import {
  completeViewerProfileAvatarUpload,
  createViewerProfileAvatarUpload,
  uploadViewerProfileAvatarTarget,
} from "../api/avatar-upload";
import { updateViewerProfile } from "../api/update-viewer-profile";
import { ViewerProfileSettingsPanel } from "./viewer-profile-settings-panel";

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

vi.mock("../api/avatar-upload", () => ({
  completeViewerProfileAvatarUpload: vi.fn(),
  createViewerProfileAvatarUpload: vi.fn(),
  uploadViewerProfileAvatarTarget: vi.fn(),
}));

vi.mock("../api/update-viewer-profile", () => ({
  updateViewerProfile: vi.fn(),
}));

function createImageFile(name: string, type = "image/png"): File {
  return new File(["avatar"], name, { type });
}

function createOversizeImageFile(name: string): File {
  return new File([new Uint8Array(5_242_881)], name, { type: "image/png" });
}

describe("ViewerProfileSettingsPanel", () => {
  beforeEach(() => {
    mockedRouter.back.mockReset();
    mockedRouter.forward.mockReset();
    mockedRouter.prefetch.mockReset();
    mockedRouter.push.mockReset();
    mockedRouter.refresh.mockReset();
    mockedRouter.replace.mockReset();
    Object.defineProperty(URL, "createObjectURL", {
      configurable: true,
      value: vi.fn(() => "blob:viewer-profile-preview"),
      writable: true,
    });
    Object.defineProperty(URL, "revokeObjectURL", {
      configurable: true,
      value: vi.fn(),
      writable: true,
    });
    vi.mocked(createViewerProfileAvatarUpload).mockReset();
    vi.mocked(uploadViewerProfileAvatarTarget).mockReset();
    vi.mocked(completeViewerProfileAvatarUpload).mockReset();
    vi.mocked(updateViewerProfile).mockReset();
    vi.mocked(createViewerProfileAvatarUpload).mockResolvedValue({
      avatarUploadToken: "vcupl_test_token",
      expiresAt: "2026-04-16T12:00:00Z",
      uploadTarget: {
        fileName: "avatar.png",
        mimeType: "image/png",
        upload: {
          headers: {
            "Content-Type": "image/png",
          },
          method: "PUT",
          url: "https://raw-bucket.example.com/avatar.png",
        },
      },
    });
    vi.mocked(uploadViewerProfileAvatarTarget).mockResolvedValue(undefined);
    vi.mocked(completeViewerProfileAvatarUpload).mockResolvedValue({
      avatar: {
        durationSeconds: null,
        id: "asset_viewer_avatar_uploaded",
        kind: "image",
        posterUrl: null,
        url: "https://cdn.example.com/viewer/uploaded/avatar.png",
      },
      avatarUploadToken: "vcupl_test_token",
    });
    vi.mocked(updateViewerProfile).mockResolvedValue(undefined);
  });

  it("renders the fan profile editor surface with the reference-first structure", () => {
    render(
      <ViewerProfileSettingsPanel
        initialValues={{
          avatarUrl: "https://cdn.example.com/viewer/mina/avatar.jpg",
          displayName: "Mina Rei",
          handle: "@minarei",
        }}
      />,
    );

    expect(screen.getByRole("link", { name: "fan hub に戻る" })).toHaveAttribute("href", "/fan");
    expect(screen.getByRole("heading", { name: "プロフィールを編集" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: /display name/i })).toHaveValue("Mina Rei");
    expect(screen.getByRole("textbox", { name: /handle/i })).toHaveValue("@minarei");
    expect(screen.getByRole("button", { name: "保存する" })).toBeInTheDocument();
  });

  it("submits the updated shared profile and routes back to fan hub", async () => {
    const user = userEvent.setup();

    render(
      <ViewerProfileSettingsPanel
        initialValues={{
          avatarUrl: null,
          displayName: "Mina Rei",
          handle: "@minarei",
        }}
      />,
    );

    await user.clear(screen.getByRole("textbox", { name: /display name/i }));
    await user.type(screen.getByRole("textbox", { name: /display name/i }), "sabe");
    await user.clear(screen.getByRole("textbox", { name: /handle/i }));
    await user.type(screen.getByRole("textbox", { name: /handle/i }), "@sabe_123");
    await user.click(screen.getByRole("button", { name: "保存する" }));

    await waitFor(() => {
      expect(updateViewerProfile).toHaveBeenCalledWith({
        displayName: "sabe",
        handle: "@sabe_123",
      });
      expect(mockedRouter.replace).toHaveBeenCalledWith("/fan");
    });
  });

  it("keeps the existing avatar and hides clear controls after clearing a selected file", async () => {
    const user = userEvent.setup();

    render(
      <ViewerProfileSettingsPanel
        initialValues={{
          avatarUrl: "https://cdn.example.com/viewer/mina/avatar.jpg",
          displayName: "Mina Rei",
          handle: "@minarei",
        }}
      />,
    );

    await user.upload(screen.getByLabelText("avatar を変更"), createImageFile("avatar.png"));

    expect(screen.getByText("保存すると新しい画像に切り替わります。")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "選択を外す" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "選択を外す" }));
    await user.click(screen.getByRole("button", { name: "保存する" }));

    await waitFor(() => {
      expect(screen.queryByText("保存すると新しい画像に切り替わります。")).not.toBeInTheDocument();
      expect(screen.queryByRole("button", { name: "選択を外す" })).not.toBeInTheDocument();
      expect(createViewerProfileAvatarUpload).not.toHaveBeenCalled();
      expect(uploadViewerProfileAvatarTarget).not.toHaveBeenCalled();
      expect(completeViewerProfileAvatarUpload).not.toHaveBeenCalled();
      expect(updateViewerProfile).toHaveBeenCalledWith({
        displayName: "Mina Rei",
        handle: "@minarei",
      });
    });
  });

  it("shows and clears the invalid avatar state without dropping the persisted avatar", async () => {
    const user = userEvent.setup();

    render(
      <ViewerProfileSettingsPanel
        initialValues={{
          avatarUrl: "https://cdn.example.com/viewer/mina/avatar.jpg",
          displayName: "Mina Rei",
          handle: "@minarei",
        }}
      />,
    );

    await user.upload(screen.getByLabelText("avatar を変更"), createOversizeImageFile("avatar.png"));

    expect(screen.getByText("avatar は 5MB 以下の画像を選択してください。")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "選択を外す" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "選択を外す" }));
    await user.click(screen.getByRole("button", { name: "保存する" }));

    await waitFor(() => {
      expect(screen.queryByText("avatar は 5MB 以下の画像を選択してください。")).not.toBeInTheDocument();
      expect(screen.queryByRole("button", { name: "選択を外す" })).not.toBeInTheDocument();
      expect(createViewerProfileAvatarUpload).not.toHaveBeenCalled();
      expect(uploadViewerProfileAvatarTarget).not.toHaveBeenCalled();
      expect(completeViewerProfileAvatarUpload).not.toHaveBeenCalled();
      expect(updateViewerProfile).toHaveBeenCalledWith({
        displayName: "Mina Rei",
        handle: "@minarei",
      });
    });
  });

  it("allows clearing the avatar selection after upload completion state is reached", async () => {
    const user = userEvent.setup();
    vi.mocked(updateViewerProfile)
      .mockRejectedValueOnce(
        new ApiError("API request failed with a non-success status.", {
          code: "http",
          details: JSON.stringify({
            error: {
              code: "handle_already_taken",
              message: "handle is already taken",
            },
            meta: {
              requestId: "req_handle_taken_after_avatar_upload",
            },
          }),
          status: 409,
        }),
      )
      .mockResolvedValueOnce(undefined);

    render(
      <ViewerProfileSettingsPanel
        initialValues={{
          avatarUrl: "https://cdn.example.com/viewer/mina/avatar.jpg",
          displayName: "Mina Rei",
          handle: "@minarei",
        }}
      />,
    );

    await user.upload(screen.getByLabelText("avatar を変更"), createImageFile("avatar.png"));
    await user.click(screen.getByRole("button", { name: "保存する" }));

    await waitFor(() => {
      expect(createViewerProfileAvatarUpload).toHaveBeenCalledTimes(1);
      expect(uploadViewerProfileAvatarTarget).toHaveBeenCalledTimes(1);
      expect(completeViewerProfileAvatarUpload).toHaveBeenCalledTimes(1);
      expect(updateViewerProfile).toHaveBeenCalledWith({
        avatarUploadToken: "vcupl_test_token",
        displayName: "Mina Rei",
        handle: "@minarei",
      });
    });

    expect(screen.getByRole("button", { name: "選択を外す" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "選択を外す" }));
    await user.click(screen.getByRole("button", { name: "保存する" }));

    await waitFor(() => {
      expect(screen.queryByRole("button", { name: "選択を外す" })).not.toBeInTheDocument();
      expect(createViewerProfileAvatarUpload).toHaveBeenCalledTimes(1);
      expect(uploadViewerProfileAvatarTarget).toHaveBeenCalledTimes(1);
      expect(completeViewerProfileAvatarUpload).toHaveBeenCalledTimes(1);
      expect(updateViewerProfile).toHaveBeenNthCalledWith(2, {
        displayName: "Mina Rei",
        handle: "@minarei",
      });
    });
  });

  it("shows the mapped save error when the API rejects the update", async () => {
    const user = userEvent.setup();

    vi.mocked(updateViewerProfile).mockRejectedValue(
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

    render(
      <ViewerProfileSettingsPanel
        initialValues={{
          avatarUrl: null,
          displayName: "Mina Rei",
          handle: "@minarei",
        }}
      />,
    );

    await user.click(screen.getByRole("button", { name: "保存する" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "そのhandleは既に使われています。別のhandleを入力してください。",
    );
    expect(mockedRouter.replace).not.toHaveBeenCalled();
  });
});

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
import {
  completeCreatorRegistrationAvatarUpload,
  createCreatorRegistrationAvatarUpload,
  uploadCreatorRegistrationAvatarTarget,
} from "@/features/creator-entry/api/avatar-upload";

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

vi.mock("@/features/creator-entry/api/avatar-upload", () => ({
  completeCreatorRegistrationAvatarUpload: vi.fn(),
  createCreatorRegistrationAvatarUpload: vi.fn(),
  uploadCreatorRegistrationAvatarTarget: vi.fn(),
}));

function createAvatarFile(name: string, type = "image/png", body = "avatar"): File {
  return new File([body], name, { type });
}

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
    vi.mocked(createCreatorRegistrationAvatarUpload).mockReset();
    vi.mocked(uploadCreatorRegistrationAvatarTarget).mockReset();
    vi.mocked(completeCreatorRegistrationAvatarUpload).mockReset();
  });

  it("validates display name before sending the request", async () => {
    const user = userEvent.setup();

    renderPanel();

    await user.click(screen.getByRole("button", { name: "申し込む" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("表示名を入力してください。");
    expect(registerCreator).not.toHaveBeenCalled();
  });

  it("validates handle before sending the request", async () => {
    const user = userEvent.setup();

    renderPanel();

    await user.type(screen.getByRole("textbox", { name: "Display name" }), "Mina Rei");
    await user.click(screen.getByRole("button", { name: "申し込む" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("handleを入力してください。");
    expect(registerCreator).not.toHaveBeenCalled();
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

    await user.type(screen.getByRole("textbox", { name: "Display name" }), "Mina Rei");
    await user.type(screen.getByRole("textbox", { name: "Handle" }), "@mina.rei");
    await user.type(screen.getByRole("textbox", { name: "Bio" }), "quiet rooftop の continuation を中心に投稿します。");
    await user.click(screen.getByRole("button", { name: "申し込む" }));

    await waitFor(() => {
      expect(registerCreator).toHaveBeenCalledWith({
        bio: "quiet rooftop の continuation を中心に投稿します。",
        displayName: "Mina Rei",
        handle: "@mina.rei",
      });
      expect(getCurrentViewerBootstrap).toHaveBeenCalledWith({
        credentials: "include",
      });
      expect(mockedRouter.push).toHaveBeenCalledWith("/fan/creator/success");
    });
    expect(createCreatorRegistrationAvatarUpload).not.toHaveBeenCalled();
  });

  it("uploads the selected avatar before registration", async () => {
    const user = userEvent.setup();

    vi.mocked(createCreatorRegistrationAvatarUpload).mockResolvedValue({
      avatarUploadToken: "vcupl_token",
      expiresAt: "2026-04-10T12:15:00Z",
      uploadTarget: {
        fileName: "avatar.png",
        mimeType: "image/png",
        upload: {
          headers: {
            "Content-Type": "image/png",
          },
          method: "PUT",
          url: "https://raw-bucket.example.com/avatar",
        },
      },
    });
    vi.mocked(uploadCreatorRegistrationAvatarTarget).mockResolvedValue(undefined);
    vi.mocked(completeCreatorRegistrationAvatarUpload).mockResolvedValue({
      avatar: {
        durationSeconds: null,
        id: "asset_creator_registration_avatar_fixed",
        kind: "image",
        posterUrl: null,
        url: "https://cdn.example.com/creator-avatar/avatar.png",
      },
      avatarUploadToken: "vcupl_token",
    });
    vi.mocked(registerCreator).mockResolvedValue(undefined);
    vi.mocked(getCurrentViewerBootstrap).mockResolvedValue({
      activeMode: "fan",
      canAccessCreatorMode: true,
      id: "viewer_123",
    });

    renderPanel();

    await user.upload(screen.getByLabelText("Avatar image"), createAvatarFile("avatar.png"));
    await user.type(screen.getByRole("textbox", { name: "Display name" }), "Mina Rei");
    await user.type(screen.getByRole("textbox", { name: "Handle" }), "@mina.rei");
    await user.click(screen.getByRole("button", { name: "申し込む" }));

    await waitFor(() => {
      expect(createCreatorRegistrationAvatarUpload).toHaveBeenCalledTimes(1);
      expect(uploadCreatorRegistrationAvatarTarget).toHaveBeenCalledTimes(1);
      expect(completeCreatorRegistrationAvatarUpload).toHaveBeenCalledWith("vcupl_token");
      expect(registerCreator).toHaveBeenCalledWith({
        avatarUploadToken: "vcupl_token",
        bio: "",
        displayName: "Mina Rei",
        handle: "@mina.rei",
      });
    });
  });

  it("rejects unsupported avatar files before submit", async () => {
    const user = userEvent.setup({ applyAccept: false });

    renderPanel();

    await user.upload(screen.getByLabelText("Avatar image"), createAvatarFile("avatar.gif", "image/gif"));
    await user.type(screen.getByRole("textbox", { name: "Display name" }), "Mina Rei");
    await user.type(screen.getByRole("textbox", { name: "Handle" }), "@mina.rei");
    await user.click(screen.getByRole("button", { name: "申し込む" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("avatar は JPEG / PNG / WebP のみ選択できます。");
    expect(createCreatorRegistrationAvatarUpload).not.toHaveBeenCalled();
    expect(registerCreator).not.toHaveBeenCalled();
  });

  it("reuses the completed avatar token when registration fails for a non-avatar reason", async () => {
    const user = userEvent.setup();

    vi.mocked(createCreatorRegistrationAvatarUpload).mockResolvedValue({
      avatarUploadToken: "vcupl_token",
      expiresAt: "2026-04-10T12:15:00Z",
      uploadTarget: {
        fileName: "avatar.png",
        mimeType: "image/png",
        upload: {
          headers: {
            "Content-Type": "image/png",
          },
          method: "PUT",
          url: "https://raw-bucket.example.com/avatar",
        },
      },
    });
    vi.mocked(uploadCreatorRegistrationAvatarTarget).mockResolvedValue(undefined);
    vi.mocked(completeCreatorRegistrationAvatarUpload).mockResolvedValue({
      avatar: {
        durationSeconds: null,
        id: "asset_creator_registration_avatar_fixed",
        kind: "image",
        posterUrl: null,
        url: "https://cdn.example.com/creator-avatar/avatar.png",
      },
      avatarUploadToken: "vcupl_token",
    });
    vi.mocked(registerCreator)
      .mockRejectedValueOnce(
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
      )
      .mockResolvedValueOnce(undefined);
    vi.mocked(getCurrentViewerBootstrap).mockResolvedValue({
      activeMode: "fan",
      canAccessCreatorMode: true,
      id: "viewer_123",
    });

    renderPanel();

    await user.upload(screen.getByLabelText("Avatar image"), createAvatarFile("avatar.png"));
    await user.type(screen.getByRole("textbox", { name: "Display name" }), "Mina Rei");
    await user.type(screen.getByRole("textbox", { name: "Handle" }), "@mina.rei");
    await user.click(screen.getByRole("button", { name: "申し込む" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "そのhandleは既に使われています。別のhandleを入力してください。",
    );

    await user.clear(screen.getByRole("textbox", { name: "Handle" }));
    await user.type(screen.getByRole("textbox", { name: "Handle" }), "@mina_rei");
    await user.click(screen.getByRole("button", { name: "申し込む" }));

    await waitFor(() => {
      expect(createCreatorRegistrationAvatarUpload).toHaveBeenCalledTimes(1);
      expect(uploadCreatorRegistrationAvatarTarget).toHaveBeenCalledTimes(1);
      expect(completeCreatorRegistrationAvatarUpload).toHaveBeenCalledTimes(1);
      expect(registerCreator).toHaveBeenNthCalledWith(1, {
        avatarUploadToken: "vcupl_token",
        bio: "",
        displayName: "Mina Rei",
        handle: "@mina.rei",
      });
      expect(registerCreator).toHaveBeenNthCalledWith(2, {
        avatarUploadToken: "vcupl_token",
        bio: "",
        displayName: "Mina Rei",
        handle: "@mina_rei",
      });
    });
  });

  it("re-uploads the avatar when registration reports an invalid avatar token", async () => {
    const user = userEvent.setup();

    vi.mocked(createCreatorRegistrationAvatarUpload)
      .mockResolvedValueOnce({
        avatarUploadToken: "vcupl_token_1",
        expiresAt: "2026-04-10T12:15:00Z",
        uploadTarget: {
          fileName: "avatar.png",
          mimeType: "image/png",
          upload: {
            headers: {
              "Content-Type": "image/png",
            },
            method: "PUT",
            url: "https://raw-bucket.example.com/avatar-1",
          },
        },
      })
      .mockResolvedValueOnce({
        avatarUploadToken: "vcupl_token_2",
        expiresAt: "2026-04-10T12:20:00Z",
        uploadTarget: {
          fileName: "avatar.png",
          mimeType: "image/png",
          upload: {
            headers: {
              "Content-Type": "image/png",
            },
            method: "PUT",
            url: "https://raw-bucket.example.com/avatar-2",
          },
        },
      });
    vi.mocked(uploadCreatorRegistrationAvatarTarget).mockResolvedValue(undefined);
    vi.mocked(completeCreatorRegistrationAvatarUpload)
      .mockResolvedValueOnce({
        avatar: {
          durationSeconds: null,
          id: "asset_creator_registration_avatar_1",
          kind: "image",
          posterUrl: null,
          url: "https://cdn.example.com/creator-avatar/avatar-1.png",
        },
        avatarUploadToken: "vcupl_token_1",
      })
      .mockResolvedValueOnce({
        avatar: {
          durationSeconds: null,
          id: "asset_creator_registration_avatar_2",
          kind: "image",
          posterUrl: null,
          url: "https://cdn.example.com/creator-avatar/avatar-2.png",
        },
        avatarUploadToken: "vcupl_token_2",
      });
    vi.mocked(registerCreator)
      .mockRejectedValueOnce(
        new ApiError("API request failed with a non-success status.", {
          code: "http",
          details: JSON.stringify({
            error: {
              code: "invalid_avatar_upload_token",
              message: "avatar upload token is invalid",
            },
            meta: {
              requestId: "req_invalid_avatar_upload_token",
            },
          }),
          status: 400,
        }),
      )
      .mockResolvedValueOnce(undefined);
    vi.mocked(getCurrentViewerBootstrap).mockResolvedValue({
      activeMode: "fan",
      canAccessCreatorMode: true,
      id: "viewer_123",
    });

    renderPanel();

    await user.upload(screen.getByLabelText("Avatar image"), createAvatarFile("avatar.png"));
    await user.type(screen.getByRole("textbox", { name: "Display name" }), "Mina Rei");
    await user.type(screen.getByRole("textbox", { name: "Handle" }), "@mina.rei");
    await user.click(screen.getByRole("button", { name: "申し込む" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "avatar upload の有効期限が切れました。もう一度申し込んでください。",
    );

    await user.click(screen.getByRole("button", { name: "申し込む" }));

    await waitFor(() => {
      expect(createCreatorRegistrationAvatarUpload).toHaveBeenCalledTimes(2);
      expect(uploadCreatorRegistrationAvatarTarget).toHaveBeenCalledTimes(2);
      expect(completeCreatorRegistrationAvatarUpload).toHaveBeenNthCalledWith(1, "vcupl_token_1");
      expect(completeCreatorRegistrationAvatarUpload).toHaveBeenNthCalledWith(2, "vcupl_token_2");
      expect(registerCreator).toHaveBeenNthCalledWith(1, {
        avatarUploadToken: "vcupl_token_1",
        bio: "",
        displayName: "Mina Rei",
        handle: "@mina.rei",
      });
      expect(registerCreator).toHaveBeenNthCalledWith(2, {
        avatarUploadToken: "vcupl_token_2",
        bio: "",
        displayName: "Mina Rei",
        handle: "@mina.rei",
      });
    });
  });

  it("shows the invalid handle message returned by the API", async () => {
    const user = userEvent.setup();

    vi.mocked(registerCreator).mockRejectedValue(
      new ApiError("API request failed with a non-success status.", {
        code: "http",
        details: JSON.stringify({
          error: {
            code: "invalid_handle",
            message: "handle is invalid",
          },
          meta: {
            requestId: "req_invalid_handle",
          },
        }),
        status: 400,
      }),
    );

    renderPanel();

    await user.type(screen.getByRole("textbox", { name: "Display name" }), "Mina Rei");
    await user.type(screen.getByRole("textbox", { name: "Handle" }), "@");
    await user.click(screen.getByRole("button", { name: "申し込む" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "handleは英数字・`.`・`_`のみ使えます。`@`は先頭に付けても構いません。",
    );
  });

  it("shows the duplicate handle message returned by the API", async () => {
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

    await user.type(screen.getByRole("textbox", { name: "Display name" }), "Mina Rei");
    await user.type(screen.getByRole("textbox", { name: "Handle" }), "mina");
    await user.click(screen.getByRole("button", { name: "申し込む" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "そのhandleは既に使われています。別のhandleを入力してください。",
    );
  });
});

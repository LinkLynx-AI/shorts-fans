import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  completeCreatorUploadPackage,
  createCreatorUploadPackage,
  CreatorUploadApiError,
  uploadCreatorUploadTarget,
} from "../api";
import { CreatorUploadForm } from "./creator-upload-form";

vi.mock("../api", async () => {
  const actual = await vi.importActual<typeof import("../api")>("../api");

  return {
    ...actual,
    completeCreatorUploadPackage: vi.fn(),
    createCreatorUploadPackage: vi.fn(),
    uploadCreatorUploadTarget: vi.fn(),
  };
});

function createVideoFile(name: string): File {
  return new File(["video"], name, { type: "video/mp4" });
}

describe("CreatorUploadForm", () => {
  beforeEach(() => {
    vi.mocked(createCreatorUploadPackage).mockReset();
    vi.mocked(uploadCreatorUploadTarget).mockReset();
    vi.mocked(completeCreatorUploadPackage).mockReset();
  });

  it("keeps completion disabled until both a main and a short are selected", async () => {
    const user = userEvent.setup();

    render(<CreatorUploadForm />);

    const submitButton = screen.getByRole("button", { name: "アップロード" });

    expect(submitButton).toBeDisabled();
    expect(screen.getByText("本編動画を追加してください")).toBeInTheDocument();
    expect(screen.getByText("ショート動画を追加してください")).toBeInTheDocument();

    await user.upload(screen.getByLabelText("本編動画ファイル"), createVideoFile("main.mp4"));

    expect(screen.getByText("main.mp4")).toBeInTheDocument();
    expect(submitButton).toBeDisabled();

    await user.upload(screen.getByLabelText("ショート動画 1 ファイル"), createVideoFile("short-1.mp4"));

    expect(screen.getByText("short-1.mp4")).toBeInTheDocument();
    expect(submitButton).toBeEnabled();
    expect(screen.getByText("1本")).toBeInTheDocument();
  });

  it("adds and removes short slots while preserving selected files", async () => {
    const user = userEvent.setup();

    render(<CreatorUploadForm />);

    await user.click(screen.getByRole("button", { name: "ショート欄を追加" }));
    expect(screen.getByLabelText("ショート動画 2 ファイル")).toBeInTheDocument();

    await user.upload(screen.getByLabelText("ショート動画 2 ファイル"), createVideoFile("short-2.mp4"));
    expect(screen.getByText("short-2.mp4")).toBeInTheDocument();

    const [firstRemoveButton] = screen.getAllByRole("button", { name: "ショート欄を削除" });

    if (!firstRemoveButton) {
      throw new Error("short remove button is missing");
    }

    await user.click(firstRemoveButton);

    expect(screen.queryByLabelText("ショート動画 2 ファイル")).not.toBeInTheDocument();
    expect(screen.getByText("short-2.mp4")).toBeInTheDocument();
    expect(screen.getByText("1本")).toBeInTheDocument();
  });

  it("renders the connected upload success state after submit", async () => {
    const user = userEvent.setup();

    vi.mocked(createCreatorUploadPackage).mockResolvedValue({
      expiresAt: "2026-04-08T12:15:00Z",
      packageToken: "cupkg_123",
      uploadTargets: {
        main: {
          fileName: "main.mp4",
          mimeType: "video/mp4",
          role: "main",
          upload: {
            headers: {
              "Content-Type": "video/mp4",
            },
            method: "PUT",
            url: "https://raw-bucket.example.com/main",
          },
          uploadEntryId: "main-entry",
        },
        shorts: [
          {
            fileName: "short-1.mp4",
            mimeType: "video/mp4",
            role: "short",
            upload: {
              headers: {
                "Content-Type": "video/mp4",
              },
              method: "PUT",
              url: "https://raw-bucket.example.com/short-1",
            },
            uploadEntryId: "short-entry-1",
          },
        ],
      },
    });
    vi.mocked(uploadCreatorUploadTarget).mockResolvedValue(undefined);
    vi.mocked(completeCreatorUploadPackage).mockResolvedValue({
      main: {
        id: "main_001",
        mediaAsset: {
          id: "asset_main_001",
          mimeType: "video/mp4",
          processingState: "uploaded",
        },
        state: "draft",
      },
      shorts: [
        {
          canonicalMainId: "main_001",
          id: "short_001",
          mediaAsset: {
            id: "asset_short_001",
            mimeType: "video/mp4",
            processingState: "uploaded",
          },
          state: "draft",
        },
      ],
    });

    render(<CreatorUploadForm />);

    await user.upload(screen.getByLabelText("本編動画ファイル"), createVideoFile("main.mp4"));
    await user.upload(screen.getByLabelText("ショート動画 1 ファイル"), createVideoFile("short-1.mp4"));
    await user.click(screen.getByRole("button", { name: "アップロード" }));

    await waitFor(() => {
      expect(createCreatorUploadPackage).toHaveBeenCalledTimes(1);
      expect(uploadCreatorUploadTarget).toHaveBeenCalledTimes(2);
      expect(completeCreatorUploadPackage).toHaveBeenCalledTimes(1);
    });

    expect(await screen.findByText("ドラフトの作成まで完了しました。")).toBeInTheDocument();
    expect(screen.getByText("main 1本 / short 1本 を保存しました。公開や審査提出はまだ行われていません。")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "アップロード完了" })).toBeDisabled();
  });

  it("shows a recoverable upload error and allows retry", async () => {
    const user = userEvent.setup();

    vi.mocked(createCreatorUploadPackage).mockResolvedValue({
      expiresAt: "2026-04-08T12:15:00Z",
      packageToken: "cupkg_123",
      uploadTargets: {
        main: {
          fileName: "main.mp4",
          mimeType: "video/mp4",
          role: "main",
          upload: {
            headers: {
              "Content-Type": "video/mp4",
            },
            method: "PUT",
            url: "https://raw-bucket.example.com/main",
          },
          uploadEntryId: "main-entry",
        },
        shorts: [
          {
            fileName: "short-1.mp4",
            mimeType: "video/mp4",
            role: "short",
            upload: {
              headers: {
                "Content-Type": "video/mp4",
              },
              method: "PUT",
              url: "https://raw-bucket.example.com/short-1",
            },
            uploadEntryId: "short-entry-1",
          },
        ],
      },
    });
    vi.mocked(uploadCreatorUploadTarget)
      .mockResolvedValueOnce(undefined)
      .mockRejectedValueOnce(new Error("s3 down"))
      .mockResolvedValueOnce(undefined)
      .mockResolvedValueOnce(undefined);
    vi.mocked(completeCreatorUploadPackage).mockResolvedValue({
      main: {
        id: "main_001",
        mediaAsset: {
          id: "asset_main_001",
          mimeType: "video/mp4",
          processingState: "uploaded",
        },
        state: "draft",
      },
      shorts: [
        {
          canonicalMainId: "main_001",
          id: "short_001",
          mediaAsset: {
            id: "asset_short_001",
            mimeType: "video/mp4",
            processingState: "uploaded",
          },
          state: "draft",
        },
      ],
    });

    render(<CreatorUploadForm />);

    await user.upload(screen.getByLabelText("本編動画ファイル"), createVideoFile("main.mp4"));
    await user.upload(screen.getByLabelText("ショート動画 1 ファイル"), createVideoFile("short-1.mp4"));
    await user.click(screen.getByRole("button", { name: "アップロード" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "ショート動画 1のアップロードに失敗しました。再試行してください。 再試行するか、ファイルを選び直してください。",
    );
    expect(screen.getByRole("button", { name: "再試行" })).toBeEnabled();
    expect(screen.getByText("失敗")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "再試行" }));

    expect(await screen.findByText("ドラフトの作成まで完了しました。")).toBeInTheDocument();
    expect(uploadCreatorUploadTarget).toHaveBeenCalledTimes(4);
    expect(completeCreatorUploadPackage).toHaveBeenCalledTimes(1);
  });

  it("shows a backend completion error and keeps retry available", async () => {
    const user = userEvent.setup();

    vi.mocked(createCreatorUploadPackage).mockResolvedValue({
      expiresAt: "2026-04-08T12:15:00Z",
      packageToken: "cupkg_123",
      uploadTargets: {
        main: {
          fileName: "main.mp4",
          mimeType: "video/mp4",
          role: "main",
          upload: {
            headers: {
              "Content-Type": "video/mp4",
            },
            method: "PUT",
            url: "https://raw-bucket.example.com/main",
          },
          uploadEntryId: "main-entry",
        },
        shorts: [
          {
            fileName: "short-1.mp4",
            mimeType: "video/mp4",
            role: "short",
            upload: {
              headers: {
                "Content-Type": "video/mp4",
              },
              method: "PUT",
              url: "https://raw-bucket.example.com/short-1",
            },
            uploadEntryId: "short-entry-1",
          },
        ],
      },
    });
    vi.mocked(uploadCreatorUploadTarget).mockResolvedValue(undefined);
    vi.mocked(completeCreatorUploadPackage).mockRejectedValue(
      new CreatorUploadApiError("upload_expired", "creator upload package has expired", {
        requestId: "req_creator_upload_packages_complete_expired_001",
        status: 409,
      }),
    );

    render(<CreatorUploadForm />);

    await user.upload(screen.getByLabelText("本編動画ファイル"), createVideoFile("main.mp4"));
    await user.upload(screen.getByLabelText("ショート動画 1 ファイル"), createVideoFile("short-1.mp4"));
    await user.click(screen.getByRole("button", { name: "アップロード" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "アップロード準備の有効期限が切れました。再試行してください。 再試行するか、ファイルを選び直してください。",
    );
    expect(screen.getByRole("button", { name: "再試行" })).toBeEnabled();
  });
});

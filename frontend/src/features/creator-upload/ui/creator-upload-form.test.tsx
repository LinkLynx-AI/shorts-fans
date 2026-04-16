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

async function fillRequiredMetadata(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText("価格（円）"), "1800");
  await user.click(screen.getByLabelText("本編の権利確認"));
  await user.click(screen.getByLabelText("本編の同意確認"));
}

function createDeferredPromise<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void;
  let reject!: (reason?: unknown) => void;

  const promise = new Promise<T>((innerResolve, innerReject) => {
    resolve = innerResolve;
    reject = innerReject;
  });

  return {
    promise,
    reject,
    resolve,
  };
}

describe("CreatorUploadForm", () => {
  beforeEach(() => {
    vi.mocked(createCreatorUploadPackage).mockReset();
    vi.mocked(uploadCreatorUploadTarget).mockReset();
    vi.mocked(completeCreatorUploadPackage).mockReset();
  });

  it("keeps completion disabled until files, price, and confirmations are filled", async () => {
    const user = userEvent.setup();

    render(<CreatorUploadForm />);

    const submitButton = screen.getByRole("button", { name: "保存してアップロード" });

    expect(submitButton).toBeDisabled();
    expect(screen.getByText("本編動画を追加してください")).toBeInTheDocument();
    expect(screen.getByText("ショート動画を追加してください")).toBeInTheDocument();

    await user.upload(screen.getByLabelText("本編動画ファイル"), createVideoFile("main.mp4"));

    expect(screen.getByText("main.mp4")).toBeInTheDocument();
    expect(submitButton).toBeDisabled();

    await user.upload(screen.getByLabelText("ショート動画 1 ファイル"), createVideoFile("short-1.mp4"));

    expect(screen.getByText("short-1.mp4")).toBeInTheDocument();
    expect(submitButton).toBeDisabled();
    expect(screen.getByText("1 videos")).toBeInTheDocument();

    await user.type(screen.getByLabelText("価格（円）"), "1800");
    expect(submitButton).toBeDisabled();

    await user.click(screen.getByLabelText("本編の権利確認"));
    expect(submitButton).toBeDisabled();

    await user.click(screen.getByLabelText("本編の同意確認"));
    expect(submitButton).toBeEnabled();
  });

  it("adds and removes short slots while preserving selected files", async () => {
    const user = userEvent.setup();

    render(<CreatorUploadForm />);

    await user.click(screen.getByRole("button", { name: "ショート欄を追加" }));
    expect(screen.getByLabelText("ショート動画 2 ファイル")).toBeInTheDocument();
    expect(screen.getByText("0 videos")).toBeInTheDocument();

    await user.upload(screen.getByLabelText("ショート動画 2 ファイル"), createVideoFile("short-2.mp4"));
    expect(screen.getByText("short-2.mp4")).toBeInTheDocument();
    expect(screen.getByText("1 videos")).toBeInTheDocument();

    const [firstRemoveButton] = screen.getAllByRole("button", { name: "ショート欄を削除" });

    if (!firstRemoveButton) {
      throw new Error("short remove button is missing");
    }

    await user.click(firstRemoveButton);

    expect(screen.queryByLabelText("ショート動画 2 ファイル")).not.toBeInTheDocument();
    expect(screen.getByText("short-2.mp4")).toBeInTheDocument();
    expect(screen.getByText("1 videos")).toBeInTheDocument();
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
    await fillRequiredMetadata(user);
    await user.type(screen.getByLabelText("ショート動画 1 の caption"), "quiet rooftop preview。");
    await user.click(screen.getByRole("button", { name: "保存してアップロード" }));

    await waitFor(() => {
      expect(createCreatorUploadPackage).toHaveBeenCalledTimes(1);
      expect(uploadCreatorUploadTarget).toHaveBeenCalledTimes(2);
      expect(completeCreatorUploadPackage).toHaveBeenCalledTimes(1);
    });

    expect(completeCreatorUploadPackage).toHaveBeenCalledWith({
      consentConfirmed: true,
      mainUploadEntryId: "main-entry",
      ownershipConfirmed: true,
      packageToken: "cupkg_123",
      priceJpy: 1800,
      shorts: [
        {
          caption: "quiet rooftop preview。",
          uploadEntryId: "short-entry-1",
        },
      ],
    });
    expect(await screen.findByText("処理開始を受け付けました。")).toBeInTheDocument();
    expect(screen.getByText("main 1本 / short 1本 の処理を開始しました。公開や審査提出はまだ行われていません。")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "処理開始完了" })).toBeDisabled();
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
    await fillRequiredMetadata(user);
    await user.click(screen.getByRole("button", { name: "保存してアップロード" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "ショート動画 1のアップロードに失敗しました。再試行してください。 再試行するか、ファイルを選び直してください。",
    );
    expect(screen.getByRole("button", { name: "再試行" })).toBeEnabled();
    expect(screen.getByText("失敗")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "再試行" }));

    expect(await screen.findByText("処理開始を受け付けました。")).toBeInTheDocument();
    expect(uploadCreatorUploadTarget).toHaveBeenCalledTimes(4);
    expect(completeCreatorUploadPackage).toHaveBeenCalledTimes(1);
  });

  it("prevents submit re-entry while an upload request is in flight", async () => {
    const user = userEvent.setup();
    const createDeferred = createDeferredPromise<Awaited<ReturnType<typeof createCreatorUploadPackage>>>();

    vi.mocked(createCreatorUploadPackage).mockReturnValue(createDeferred.promise);
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
    await fillRequiredMetadata(user);

    const submitButton = screen.getByRole("button", { name: "保存してアップロード" });

    await Promise.all([
      user.click(submitButton),
      user.click(submitButton),
    ]);

    await waitFor(() => {
      expect(createCreatorUploadPackage).toHaveBeenCalledTimes(1);
    });

    createDeferred.resolve({
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

    expect(await screen.findByText("処理開始を受け付けました。")).toBeInTheDocument();
    expect(uploadCreatorUploadTarget).toHaveBeenCalledTimes(2);
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
    await fillRequiredMetadata(user);
    await user.click(screen.getByRole("button", { name: "保存してアップロード" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "アップロード準備の有効期限が切れました。再試行してください。 再試行するか、ファイルを選び直してください。",
    );
    expect(screen.getByRole("button", { name: "再試行" })).toBeEnabled();
  });

  it("allows optional caption to remain empty and sends null on completion", async () => {
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
    await fillRequiredMetadata(user);
    await user.click(screen.getByRole("button", { name: "保存してアップロード" }));

    await waitFor(() => {
      expect(completeCreatorUploadPackage).toHaveBeenCalledWith({
        consentConfirmed: true,
        mainUploadEntryId: "main-entry",
        ownershipConfirmed: true,
        packageToken: "cupkg_123",
        priceJpy: 1800,
        shorts: [
          {
            caption: null,
            uploadEntryId: "short-entry-1",
          },
        ],
      });
    });
  });
});

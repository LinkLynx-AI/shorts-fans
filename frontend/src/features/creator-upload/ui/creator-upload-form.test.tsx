import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { CreatorUploadForm } from "./creator-upload-form";

function createVideoFile(name: string): File {
  return new File(["video"], name, { type: "video/mp4" });
}

describe("CreatorUploadForm", () => {
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

  it("shows local loading and error states after submit without backend connectivity", async () => {
    const user = userEvent.setup();

    render(<CreatorUploadForm />);

    await user.upload(screen.getByLabelText("本編動画ファイル"), createVideoFile("main.mp4"));
    await user.upload(screen.getByLabelText("ショート動画 1 ファイル"), createVideoFile("short-1.mp4"));
    await user.click(screen.getByRole("button", { name: "アップロード" }));

    expect(screen.getByRole("button", { name: "接続準備中..." })).toBeDisabled();
    expect(screen.getByText("upload package を準備しています...")).toBeInTheDocument();

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "アップロード API はまだ接続されていません。UI shell のみ先に実装されています。",
    );
  });
});

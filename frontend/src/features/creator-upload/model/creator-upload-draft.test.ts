import {
  addCreatorUploadShortSlot,
  createInitialCreatorUploadDraft,
  getCreatorUploadPendingMessage,
  getCreatorUploadSelectedShortCount,
  getCreatorUploadSelectedShorts,
  getCreatorUploadSubmitLabel,
  isCreatorUploadReady,
  isCreatorUploadSubmitting,
  removeCreatorUploadShortSlot,
  setCreatorUploadMainFile,
  setCreatorUploadMainTransferState,
  setCreatorUploadShortFile,
  setCreatorUploadSubmissionError,
  setCreatorUploadSubmissionSuccess,
  startCreatorUploadCompletion,
  startCreatorUploadInitiation,
  startCreatorUploadTransfer,
} from "./creator-upload-draft";

function createVideoFile(name: string): File {
  return new File(["video"], name, { type: "video/mp4" });
}

describe("creator upload draft helpers", () => {
  it("starts with one empty short slot and disabled completion", () => {
    const draft = createInitialCreatorUploadDraft();

    expect(draft.mainFile).toBeNull();
    expect(draft.shortSlots).toHaveLength(1);
    expect(draft.shortSlots[0]?.file).toBeNull();
    expect(getCreatorUploadSelectedShortCount(draft)).toBe(0);
    expect(isCreatorUploadReady(draft)).toBe(false);
    expect(getCreatorUploadSubmitLabel(draft)).toBe("アップロード");
  });

  it("requires both a main and at least one short before completion is ready", () => {
    const mainOnlyDraft = setCreatorUploadMainFile(createInitialCreatorUploadDraft(), createVideoFile("main.mp4"));
    const readyDraft = setCreatorUploadShortFile(mainOnlyDraft, 0, createVideoFile("short-1.mp4"));

    expect(isCreatorUploadReady(mainOnlyDraft)).toBe(false);
    expect(isCreatorUploadReady(readyDraft)).toBe(true);
    expect(getCreatorUploadSelectedShorts(readyDraft)).toEqual([
      {
        file: readyDraft.shortSlots[0]?.file,
        slotIndex: 0,
      },
    ]);
  });

  it("adds and removes short slots while preserving selected files", () => {
    const draft = addCreatorUploadShortSlot(createInitialCreatorUploadDraft());
    const selectedDraft = setCreatorUploadShortFile(draft, 1, createVideoFile("short-2.mp4"));
    const reducedDraft = removeCreatorUploadShortSlot(selectedDraft, 0);

    expect(draft.shortSlots).toHaveLength(2);
    expect(getCreatorUploadSelectedShortCount(selectedDraft)).toBe(1);
    expect(reducedDraft.shortSlots).toHaveLength(1);
    expect(reducedDraft.shortSlots[0]?.file?.name).toBe("short-2.mp4");
  });

  it("resets transfer progress when files change after an outcome", () => {
    const readyDraft = setCreatorUploadShortFile(
      setCreatorUploadMainFile(createInitialCreatorUploadDraft(), createVideoFile("main.mp4")),
      0,
      createVideoFile("short-1.mp4"),
    );
    const succeededDraft = setCreatorUploadSubmissionSuccess(
      setCreatorUploadMainTransferState(readyDraft, { kind: "uploaded" }),
      "main_001",
      ["short_001"],
    );
    const updatedDraft = setCreatorUploadMainFile(succeededDraft, createVideoFile("main-v2.mp4"));

    expect(succeededDraft.submissionState).toEqual({
      kind: "success",
      mainId: "main_001",
      shortIds: ["short_001"],
    });
    expect(updatedDraft.submissionState).toEqual({ kind: "idle" });
    expect(updatedDraft.mainTransferState).toEqual({ kind: "idle" });
  });

  it("ignores invalid short indexes", () => {
    const draft = createInitialCreatorUploadDraft();

    expect(setCreatorUploadShortFile(draft, 9, createVideoFile("ignored.mp4"))).toBe(draft);
    expect(removeCreatorUploadShortSlot(draft, 9)).toBe(draft);
  });

  it("tracks pending submission labels and messages separately from ready validation", () => {
    const readyDraft = setCreatorUploadShortFile(
      setCreatorUploadMainFile(createInitialCreatorUploadDraft(), createVideoFile("main.mp4")),
      0,
      createVideoFile("short-1.mp4"),
    );

    const initiatingDraft = startCreatorUploadInitiation(readyDraft);
    const uploadingDraft = startCreatorUploadTransfer(initiatingDraft);
    const completingDraft = startCreatorUploadCompletion(uploadingDraft);
    const erroredDraft = setCreatorUploadSubmissionError(completingDraft, "failed");

    expect(isCreatorUploadSubmitting(initiatingDraft)).toBe(true);
    expect(getCreatorUploadPendingMessage(initiatingDraft)).toBe("upload package を準備しています...");
    expect(getCreatorUploadSubmitLabel(uploadingDraft)).toBe("アップロード中...");
    expect(getCreatorUploadPendingMessage(completingDraft)).toBe("アップロード結果を保存しています...");
    expect(isCreatorUploadSubmitting(erroredDraft)).toBe(false);
    expect(getCreatorUploadSubmitLabel(erroredDraft)).toBe("再試行");
  });
});

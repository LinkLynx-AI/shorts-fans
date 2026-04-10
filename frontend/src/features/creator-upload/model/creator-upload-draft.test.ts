import {
  addCreatorUploadShortSlot,
  createInitialCreatorUploadDraft,
  getCreatorUploadMainPriceJpy,
  getCreatorUploadPendingMessage,
  getCreatorUploadSelectedShortCount,
  getCreatorUploadSelectedShorts,
  getCreatorUploadSubmitLabel,
  isCreatorUploadReady,
  isCreatorUploadSubmitting,
  removeCreatorUploadShortSlot,
  setCreatorUploadMainConsentConfirmed,
  setCreatorUploadMainFile,
  setCreatorUploadMainOwnershipConfirmed,
  setCreatorUploadMainPriceJpyInput,
  setCreatorUploadMainTransferState,
  setCreatorUploadShortCaption,
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
    expect(draft.mainPriceJpyInput).toBe("");
    expect(draft.shortSlots).toHaveLength(1);
    expect(draft.shortSlots[0]?.file).toBeNull();
    expect(draft.shortSlots[0]?.caption).toBe("");
    expect(getCreatorUploadSelectedShortCount(draft)).toBe(0);
    expect(isCreatorUploadReady(draft)).toBe(false);
    expect(getCreatorUploadSubmitLabel(draft)).toBe("アップロード");
  });

  it("requires file selection, price, and confirmations before completion is ready", () => {
    const mainOnlyDraft = setCreatorUploadMainFile(createInitialCreatorUploadDraft(), createVideoFile("main.mp4"));
    const shortSelectedDraft = setCreatorUploadShortFile(mainOnlyDraft, 0, createVideoFile("short-1.mp4"));
    const pricedDraft = setCreatorUploadMainPriceJpyInput(shortSelectedDraft, "1800");
    const ownershipConfirmedDraft = setCreatorUploadMainOwnershipConfirmed(pricedDraft, true);
    const readyDraft = setCreatorUploadMainConsentConfirmed(ownershipConfirmedDraft, true);

    expect(isCreatorUploadReady(mainOnlyDraft)).toBe(false);
    expect(isCreatorUploadReady(shortSelectedDraft)).toBe(false);
    expect(isCreatorUploadReady(pricedDraft)).toBe(false);
    expect(isCreatorUploadReady(readyDraft)).toBe(true);
    expect(getCreatorUploadMainPriceJpy(readyDraft)).toBe(1800);
    expect(getCreatorUploadSelectedShorts(readyDraft)).toEqual([
      {
        caption: null,
        file: readyDraft.shortSlots[0]?.file,
        slotIndex: 0,
      },
    ]);
  });

  it("adds and removes short slots while preserving selected files", () => {
    const draft = addCreatorUploadShortSlot(createInitialCreatorUploadDraft());
    const selectedDraft = setCreatorUploadShortCaption(
      setCreatorUploadShortFile(draft, 1, createVideoFile("short-2.mp4")),
      1,
      "short caption",
    );
    const reducedDraft = removeCreatorUploadShortSlot(selectedDraft, 0);

    expect(draft.shortSlots).toHaveLength(2);
    expect(getCreatorUploadSelectedShortCount(selectedDraft)).toBe(1);
    expect(reducedDraft.shortSlots).toHaveLength(1);
    expect(reducedDraft.shortSlots[0]?.file?.name).toBe("short-2.mp4");
    expect(reducedDraft.shortSlots[0]?.caption).toBe("short caption");
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
    expect(setCreatorUploadShortCaption(draft, 9, "ignored")).toBe(draft);
    expect(removeCreatorUploadShortSlot(draft, 9)).toBe(draft);
  });

  it("tracks pending submission labels and messages separately from ready validation", () => {
    const readyDraft = setCreatorUploadMainConsentConfirmed(
      setCreatorUploadMainOwnershipConfirmed(
        setCreatorUploadMainPriceJpyInput(
          setCreatorUploadShortFile(
            setCreatorUploadMainFile(createInitialCreatorUploadDraft(), createVideoFile("main.mp4")),
            0,
            createVideoFile("short-1.mp4"),
          ),
          "1800",
        ),
        true,
      ),
      true,
    );

    const initiatingDraft = startCreatorUploadInitiation(readyDraft);
    const uploadingDraft = startCreatorUploadTransfer(initiatingDraft);
    const completingDraft = startCreatorUploadCompletion(uploadingDraft);
    const erroredDraft = setCreatorUploadSubmissionError(completingDraft, "failed");

    expect(isCreatorUploadSubmitting(initiatingDraft)).toBe(true);
    expect(getCreatorUploadPendingMessage(initiatingDraft)).toBe("upload package を準備しています...");
    expect(getCreatorUploadSubmitLabel(uploadingDraft)).toBe("アップロード中...");
    expect(getCreatorUploadPendingMessage(completingDraft)).toBe("処理開始を受け付けています...");
    expect(isCreatorUploadSubmitting(erroredDraft)).toBe(false);
    expect(getCreatorUploadSubmitLabel(erroredDraft)).toBe("再試行");
  });

  it("normalizes blank captions to null and resets progress when metadata changes", () => {
    const readyDraft = setCreatorUploadMainConsentConfirmed(
      setCreatorUploadMainOwnershipConfirmed(
        setCreatorUploadMainPriceJpyInput(
          setCreatorUploadShortCaption(
            setCreatorUploadShortFile(
              setCreatorUploadMainFile(createInitialCreatorUploadDraft(), createVideoFile("main.mp4")),
              0,
              createVideoFile("short-1.mp4"),
            ),
            0,
            "  quiet rooftop preview。  ",
          ),
          "1800",
        ),
        true,
      ),
      true,
    );
    const succeededDraft = setCreatorUploadSubmissionSuccess(
      setCreatorUploadMainTransferState(readyDraft, { kind: "uploaded" }),
      "main_001",
      ["short_001"],
    );
    const updatedDraft = setCreatorUploadShortCaption(succeededDraft, 0, "   ");

    expect(getCreatorUploadSelectedShorts(readyDraft)[0]?.caption).toBe("quiet rooftop preview。");
    expect(getCreatorUploadSelectedShorts(updatedDraft)[0]?.caption).toBeNull();
    expect(updatedDraft.submissionState).toEqual({ kind: "idle" });
  });
});

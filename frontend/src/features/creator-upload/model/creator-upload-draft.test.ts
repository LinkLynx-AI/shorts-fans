import {
  addCreatorUploadShortSlot,
  createInitialCreatorUploadDraft,
  getCreatorUploadSelectedShortCount,
  isCreatorUploadReady,
  removeCreatorUploadShortSlot,
  setCreatorUploadMainFile,
  setCreatorUploadShortFile,
  setCreatorUploadSubmissionError,
  startCreatorUploadSubmission,
} from "./creator-upload-draft";

function createVideoFile(name: string): File {
  return new File(["video"], name, { type: "video/mp4" });
}

describe("creator upload draft helpers", () => {
  it("starts with one empty short slot and disabled completion", () => {
    const draft = createInitialCreatorUploadDraft();

    expect(draft.mainFile).toBeNull();
    expect(draft.shortFiles).toEqual([null]);
    expect(getCreatorUploadSelectedShortCount(draft)).toBe(0);
    expect(isCreatorUploadReady(draft)).toBe(false);
  });

  it("requires both a main and at least one short before completion is ready", () => {
    const mainOnlyDraft = setCreatorUploadMainFile(createInitialCreatorUploadDraft(), createVideoFile("main.mp4"));
    const readyDraft = setCreatorUploadShortFile(mainOnlyDraft, 0, createVideoFile("short-1.mp4"));

    expect(isCreatorUploadReady(mainOnlyDraft)).toBe(false);
    expect(isCreatorUploadReady(readyDraft)).toBe(true);
  });

  it("adds and removes short slots while clearing submission errors", () => {
    const erroredDraft = setCreatorUploadSubmissionError(createInitialCreatorUploadDraft(), "not connected");
    const extendedDraft = addCreatorUploadShortSlot(erroredDraft);
    const selectedDraft = setCreatorUploadShortFile(extendedDraft, 1, createVideoFile("short-2.mp4"));
    const reducedDraft = removeCreatorUploadShortSlot(selectedDraft, 0);

    expect(extendedDraft.shortFiles).toEqual([null, null]);
    expect(extendedDraft.submissionState).toEqual({ kind: "idle" });
    expect(getCreatorUploadSelectedShortCount(selectedDraft)).toBe(1);
    expect(reducedDraft.shortFiles).toHaveLength(1);
    expect(reducedDraft.shortFiles[0]?.name).toBe("short-2.mp4");
  });

  it("ignores invalid short indexes", () => {
    const draft = createInitialCreatorUploadDraft();

    expect(setCreatorUploadShortFile(draft, 9, createVideoFile("ignored.mp4"))).toBe(draft);
    expect(removeCreatorUploadShortSlot(draft, 9)).toBe(draft);
  });

  it("tracks submitting state separately from ready validation", () => {
    const readyDraft = setCreatorUploadShortFile(
      setCreatorUploadMainFile(createInitialCreatorUploadDraft(), createVideoFile("main.mp4")),
      0,
      createVideoFile("short-1.mp4"),
    );

    expect(startCreatorUploadSubmission(readyDraft).submissionState).toEqual({ kind: "submitting" });
  });
});

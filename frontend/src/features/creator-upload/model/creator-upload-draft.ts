export type CreatorUploadSubmissionState =
  | { kind: "idle" }
  | { kind: "submitting" }
  | { kind: "error"; message: string };

export type CreatorUploadDraft = {
  mainFile: File | null;
  shortFiles: readonly (File | null)[];
  submissionState: CreatorUploadSubmissionState;
};

function createIdleSubmissionState(): CreatorUploadSubmissionState {
  return { kind: "idle" };
}

/**
 * creator upload form の初期 draft state を返す。
 */
export function createInitialCreatorUploadDraft(): CreatorUploadDraft {
  return {
    mainFile: null,
    shortFiles: [null],
    submissionState: createIdleSubmissionState(),
  };
}

/**
 * 選択済み short 動画の本数を返す。
 */
export function getCreatorUploadSelectedShortCount(draft: CreatorUploadDraft): number {
  return draft.shortFiles.filter((file) => file !== null).length;
}

/**
 * upload 送信条件を満たしているかを判定する。
 */
export function isCreatorUploadReady(draft: CreatorUploadDraft): boolean {
  return draft.mainFile !== null && getCreatorUploadSelectedShortCount(draft) > 0;
}

/**
 * main 動画の選択状態を更新する。
 */
export function setCreatorUploadMainFile(
  draft: CreatorUploadDraft,
  file: File | null,
): CreatorUploadDraft {
  return {
    ...draft,
    mainFile: file,
    submissionState: createIdleSubmissionState(),
  };
}

/**
 * short 動画の選択状態を更新する。
 */
export function setCreatorUploadShortFile(
  draft: CreatorUploadDraft,
  index: number,
  file: File | null,
): CreatorUploadDraft {
  if (!Number.isInteger(index) || index < 0 || index >= draft.shortFiles.length) {
    return draft;
  }

  return {
    ...draft,
    shortFiles: draft.shortFiles.map((currentFile, currentIndex) => (currentIndex === index ? file : currentFile)),
    submissionState: createIdleSubmissionState(),
  };
}

/**
 * short 動画の入力欄を追加する。
 */
export function addCreatorUploadShortSlot(draft: CreatorUploadDraft): CreatorUploadDraft {
  return {
    ...draft,
    shortFiles: [...draft.shortFiles, null],
    submissionState: createIdleSubmissionState(),
  };
}

/**
 * short 動画の入力欄を削除する。
 */
export function removeCreatorUploadShortSlot(
  draft: CreatorUploadDraft,
  index: number,
): CreatorUploadDraft {
  if (!Number.isInteger(index) || index < 0 || index >= draft.shortFiles.length) {
    return draft;
  }

  return {
    ...draft,
    shortFiles: draft.shortFiles.filter((_, currentIndex) => currentIndex !== index),
    submissionState: createIdleSubmissionState(),
  };
}

/**
 * upload 送信中の状態へ遷移させる。
 */
export function startCreatorUploadSubmission(draft: CreatorUploadDraft): CreatorUploadDraft {
  return {
    ...draft,
    submissionState: { kind: "submitting" },
  };
}

/**
 * upload 送信エラーを state に反映する。
 */
export function setCreatorUploadSubmissionError(
  draft: CreatorUploadDraft,
  message: string,
): CreatorUploadDraft {
  return {
    ...draft,
    submissionState: { kind: "error", message },
  };
}

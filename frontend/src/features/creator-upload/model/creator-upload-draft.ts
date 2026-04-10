export type CreatorUploadTransferState =
  | { kind: "idle" }
  | { kind: "uploading" }
  | { kind: "uploaded" }
  | { kind: "failed"; message: string };

export type CreatorUploadSubmissionState =
  | { kind: "idle" }
  | { kind: "initiating" }
  | { kind: "uploading" }
  | { kind: "completing" }
  | { kind: "error"; message: string }
  | { kind: "success"; mainId: string; shortIds: readonly string[] };

export type CreatorUploadShortSlot = {
  caption: string;
  file: File | null;
  id: number;
  transferState: CreatorUploadTransferState;
};

export type CreatorUploadDraft = {
  mainFile: File | null;
  mainConsentConfirmed: boolean;
  mainOwnershipConfirmed: boolean;
  mainPriceJpyInput: string;
  mainTransferState: CreatorUploadTransferState;
  nextShortSlotId: number;
  shortSlots: readonly CreatorUploadShortSlot[];
  submissionState: CreatorUploadSubmissionState;
};

export type CreatorUploadSelectedShort = {
  caption: string | null;
  file: File;
  slotIndex: number;
};

function createIdleTransferState(): CreatorUploadTransferState {
  return { kind: "idle" };
}

function normalizeCaption(caption: string): string | null {
  const normalizedCaption = caption.trim();

  return normalizedCaption === "" ? null : normalizedCaption;
}

function resetTransferProgress(draft: CreatorUploadDraft): CreatorUploadDraft {
  return {
    ...draft,
    mainTransferState: createIdleTransferState(),
    shortSlots: draft.shortSlots.map((slot) => ({
      ...slot,
      transferState: createIdleTransferState(),
    })),
    submissionState: { kind: "idle" },
  };
}

/**
 * creator upload form の初期 draft state を返す。
 */
export function createInitialCreatorUploadDraft(): CreatorUploadDraft {
  return {
    mainFile: null,
    mainConsentConfirmed: false,
    mainOwnershipConfirmed: false,
    mainPriceJpyInput: "",
    mainTransferState: createIdleTransferState(),
    nextShortSlotId: 2,
    shortSlots: [
      {
        caption: "",
        file: null,
        id: 1,
        transferState: createIdleTransferState(),
      },
    ],
    submissionState: { kind: "idle" },
  };
}

/**
 * 選択済み short 動画の本数を返す。
 */
export function getCreatorUploadSelectedShortCount(draft: CreatorUploadDraft): number {
  return draft.shortSlots.filter((slot) => slot.file !== null).length;
}

/**
 * upload 送信に使う short 選択一覧を返す。
 */
export function getCreatorUploadSelectedShorts(draft: CreatorUploadDraft): readonly CreatorUploadSelectedShort[] {
  return draft.shortSlots.flatMap((slot, slotIndex) =>
    slot.file
      ? [
          {
            caption: normalizeCaption(slot.caption),
            file: slot.file,
            slotIndex,
          },
        ]
      : [],
  );
}

/**
 * 本編の価格入力値を数値化して返す。
 */
export function getCreatorUploadMainPriceJpy(draft: CreatorUploadDraft): number | null {
  const normalizedPrice = draft.mainPriceJpyInput.trim();

  if (!/^[1-9]\d*$/.test(normalizedPrice)) {
    return null;
  }

  return Number.parseInt(normalizedPrice, 10);
}

/**
 * upload 送信条件を満たしているかを判定する。
 */
export function isCreatorUploadReady(draft: CreatorUploadDraft): boolean {
  return (
    draft.mainFile !== null &&
    getCreatorUploadSelectedShortCount(draft) > 0 &&
    getCreatorUploadMainPriceJpy(draft) !== null &&
    draft.mainOwnershipConfirmed &&
    draft.mainConsentConfirmed
  );
}

/**
 * upload workflow が送信中かを返す。
 */
export function isCreatorUploadSubmitting(draft: CreatorUploadDraft): boolean {
  return (
    draft.submissionState.kind === "initiating" ||
    draft.submissionState.kind === "uploading" ||
    draft.submissionState.kind === "completing"
  );
}

/**
 * submit button の表示文言を返す。
 */
export function getCreatorUploadSubmitLabel(draft: CreatorUploadDraft): string {
  switch (draft.submissionState.kind) {
    case "initiating":
      return "準備中...";
    case "uploading":
      return "アップロード中...";
    case "completing":
      return "処理開始中...";
    case "success":
      return "処理開始完了";
    case "error":
      return "再試行";
    default:
      return "アップロード";
  }
}

/**
 * pending 状態に対応するメッセージを返す。
 */
export function getCreatorUploadPendingMessage(draft: CreatorUploadDraft): string | null {
  switch (draft.submissionState.kind) {
    case "initiating":
      return "upload package を準備しています...";
    case "uploading":
      return "動画ファイルをアップロードしています...";
    case "completing":
      return "処理開始を受け付けています...";
    default:
      return null;
  }
}

/**
 * main の価格入力値を更新する。
 */
export function setCreatorUploadMainPriceJpyInput(draft: CreatorUploadDraft, value: string): CreatorUploadDraft {
  return {
    ...resetTransferProgress(draft),
    mainPriceJpyInput: value,
  };
}

/**
 * main の権利確認状態を更新する。
 */
export function setCreatorUploadMainOwnershipConfirmed(draft: CreatorUploadDraft, checked: boolean): CreatorUploadDraft {
  return {
    ...resetTransferProgress(draft),
    mainOwnershipConfirmed: checked,
  };
}

/**
 * main の同意確認状態を更新する。
 */
export function setCreatorUploadMainConsentConfirmed(draft: CreatorUploadDraft, checked: boolean): CreatorUploadDraft {
  return {
    ...resetTransferProgress(draft),
    mainConsentConfirmed: checked,
  };
}

/**
 * main 動画の選択状態を更新する。
 */
export function setCreatorUploadMainFile(draft: CreatorUploadDraft, file: File | null): CreatorUploadDraft {
  return {
    ...resetTransferProgress(draft),
    mainFile: file,
  };
}

/**
 * short 動画の選択状態を更新する。
 */
export function setCreatorUploadShortFile(draft: CreatorUploadDraft, index: number, file: File | null): CreatorUploadDraft {
  if (!Number.isInteger(index) || index < 0 || index >= draft.shortSlots.length) {
    return draft;
  }

  const resetDraft = resetTransferProgress(draft);

  return {
    ...resetDraft,
    shortSlots: resetDraft.shortSlots.map((slot, currentIndex) =>
      currentIndex === index
        ? {
            ...slot,
            file,
          }
        : slot,
    ),
  };
}

/**
 * short caption を更新する。
 */
export function setCreatorUploadShortCaption(draft: CreatorUploadDraft, index: number, caption: string): CreatorUploadDraft {
  if (!Number.isInteger(index) || index < 0 || index >= draft.shortSlots.length) {
    return draft;
  }

  const resetDraft = resetTransferProgress(draft);

  return {
    ...resetDraft,
    shortSlots: resetDraft.shortSlots.map((slot, currentIndex) =>
      currentIndex === index
        ? {
            ...slot,
            caption,
          }
        : slot,
    ),
  };
}

/**
 * short 動画の入力欄を追加する。
 */
export function addCreatorUploadShortSlot(draft: CreatorUploadDraft): CreatorUploadDraft {
  const resetDraft = resetTransferProgress(draft);

  return {
    ...resetDraft,
    nextShortSlotId: resetDraft.nextShortSlotId + 1,
    shortSlots: [
      ...resetDraft.shortSlots,
      {
        caption: "",
        file: null,
        id: resetDraft.nextShortSlotId,
        transferState: createIdleTransferState(),
      },
    ],
  };
}

/**
 * short 動画の入力欄を削除する。
 */
export function removeCreatorUploadShortSlot(draft: CreatorUploadDraft, index: number): CreatorUploadDraft {
  if (!Number.isInteger(index) || index < 0 || index >= draft.shortSlots.length) {
    return draft;
  }

  const resetDraft = resetTransferProgress(draft);

  return {
    ...resetDraft,
    shortSlots: resetDraft.shortSlots.filter((_, currentIndex) => currentIndex !== index),
  };
}

/**
 * initiation 開始状態へ遷移させる。
 */
export function startCreatorUploadInitiation(draft: CreatorUploadDraft): CreatorUploadDraft {
  const resetDraft = resetTransferProgress(draft);

  return {
    ...resetDraft,
    submissionState: { kind: "initiating" },
  };
}

/**
 * direct upload 中状態へ遷移させる。
 */
export function startCreatorUploadTransfer(draft: CreatorUploadDraft): CreatorUploadDraft {
  return {
    ...draft,
    submissionState: { kind: "uploading" },
  };
}

/**
 * completion 中状態へ遷移させる。
 */
export function startCreatorUploadCompletion(draft: CreatorUploadDraft): CreatorUploadDraft {
  return {
    ...draft,
    submissionState: { kind: "completing" },
  };
}

/**
 * main 動画の transfer 状態を更新する。
 */
export function setCreatorUploadMainTransferState(
  draft: CreatorUploadDraft,
  transferState: CreatorUploadTransferState,
): CreatorUploadDraft {
  return {
    ...draft,
    mainTransferState: transferState,
  };
}

/**
 * short 動画の transfer 状態を更新する。
 */
export function setCreatorUploadShortTransferState(
  draft: CreatorUploadDraft,
  index: number,
  transferState: CreatorUploadTransferState,
): CreatorUploadDraft {
  if (!Number.isInteger(index) || index < 0 || index >= draft.shortSlots.length) {
    return draft;
  }

  return {
    ...draft,
    shortSlots: draft.shortSlots.map((slot, currentIndex) =>
      currentIndex === index
        ? {
            ...slot,
            transferState,
          }
        : slot,
    ),
  };
}

/**
 * upload 送信エラーを state に反映する。
 */
export function setCreatorUploadSubmissionError(draft: CreatorUploadDraft, message: string): CreatorUploadDraft {
  return {
    ...draft,
    submissionState: { kind: "error", message },
  };
}

/**
 * upload 完了成功を state に反映する。
 */
export function setCreatorUploadSubmissionSuccess(
  draft: CreatorUploadDraft,
  mainId: string,
  shortIds: readonly string[],
): CreatorUploadDraft {
  return {
    ...draft,
    submissionState: {
      kind: "success",
      mainId,
      shortIds,
    },
  };
}

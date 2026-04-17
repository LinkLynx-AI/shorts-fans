"use client";

import {
  useEffect,
  useMemo,
  useState,
} from "react";

import { ApiError } from "@/shared/api";

import {
  completeViewerProfileAvatarUpload,
  createViewerProfileAvatarUpload,
  uploadViewerProfileAvatarTarget,
} from "../api/avatar-upload";

const viewerProfileAvatarMimeTypes = ["image/jpeg", "image/png", "image/webp"] as const;
const viewerProfileAvatarAccept = viewerProfileAvatarMimeTypes.join(",");
const viewerProfileAvatarMimeTypeSet = new Set<string>(viewerProfileAvatarMimeTypes);
const viewerProfileAvatarMaxFileSizeBytes = 5_242_880;
const invalidCompletedAvatarMessage =
  "avatar upload の有効期限が切れました。もう一度画像を選択してください。";

type CompletedAvatarUpload = {
  avatarAssetID: string;
  avatarUploadToken: string;
  avatarURL: string;
};

type ViewerProfileDraftMode = "edit" | "sign-up";

type ViewerProfileAvatarState =
  | { kind: "empty" }
  | { fileName: string; kind: "invalid"; message: string }
  | { file: File; kind: "selected" }
  | { file: File; kind: "uploading" }
  | { completedUpload: CompletedAvatarUpload; file: File; kind: "completed" }
  | { file: File; kind: "failed"; message: string };

export type ViewerProfileAvatarField = {
  canClear: boolean;
  fileName: string | null;
  inputAccept: string;
  isError: boolean;
  kind: ViewerProfileAvatarState["kind"];
  message: string;
  previewUrl: string | null;
};

export type ViewerProfileDraftInitialValues = {
  avatarUrl: string | null;
  displayName: string;
  handle: string;
};

type UseViewerProfileDraftOptions = {
  initialValues?: ViewerProfileDraftInitialValues;
  mode?: ViewerProfileDraftMode;
  onDirty?: () => void;
};

const defaultInitialValues: ViewerProfileDraftInitialValues = {
  avatarUrl: null,
  displayName: "",
  handle: "",
};

function validateAvatarFile(file: File): string | null {
  if (file.size <= 0) {
    return "avatar file を読み取れませんでした。別の画像を選択してください。";
  }
  if (file.size > viewerProfileAvatarMaxFileSizeBytes) {
    return "avatar は 5MB 以下の画像を選択してください。";
  }
  if (!viewerProfileAvatarMimeTypeSet.has(file.type)) {
    return "avatar は JPEG / PNG / WebP のみ選択できます。";
  }

  return null;
}

function getAvatarUploadErrorMessage(error: unknown): string {
  if (error instanceof ApiError && error.code === "network") {
    return "avatar のアップロードに失敗しました。通信状態を確認してから再度お試しください。";
  }
  if (error instanceof ApiError) {
    return "avatar のアップロードに失敗しました。再度お試しください。";
  }

  return "avatar のアップロードに失敗しました。少し時間を置いてから再度お試しください。";
}

function getAvatarPreviewFile(state: ViewerProfileAvatarState): File | null {
  switch (state.kind) {
    case "selected":
    case "uploading":
    case "completed":
    case "failed":
      return state.file;
    case "empty":
    case "invalid":
      return null;
  }
}

function buildAvatarField(
  state: ViewerProfileAvatarState,
  previewUrl: string | null,
  {
    mode,
    persistedAvatarUrl,
  }: {
    mode: ViewerProfileDraftMode;
    persistedAvatarUrl: string | null;
  },
): ViewerProfileAvatarField {
  const hasPersistedAvatar = persistedAvatarUrl !== null;
  const resolvedPreviewUrl = previewUrl ?? persistedAvatarUrl;

  switch (state.kind) {
    case "empty":
      return {
        canClear: false,
        fileName: hasPersistedAvatar ? "現在の avatar" : null,
        inputAccept: viewerProfileAvatarAccept,
        isError: false,
        kind: hasPersistedAvatar ? "completed" : "empty",
        message: hasPersistedAvatar
          ? "現在の avatar を使用します。"
          : mode === "edit"
            ? "未設定のまま保存できます。"
            : "未設定でも登録できます。",
        previewUrl: resolvedPreviewUrl,
      };
    case "invalid":
      return {
        canClear: true,
        fileName: state.fileName,
        inputAccept: viewerProfileAvatarAccept,
        isError: true,
        kind: "invalid",
        message: state.message,
        previewUrl: resolvedPreviewUrl,
      };
    case "selected":
      return {
        canClear: true,
        fileName: state.file.name,
        inputAccept: viewerProfileAvatarAccept,
        isError: false,
        kind: "selected",
        message: mode === "edit" ? "保存時にアップロードします。" : "登録時にアップロードします。",
        previewUrl: resolvedPreviewUrl,
      };
    case "uploading":
      return {
        canClear: true,
        fileName: state.file.name,
        inputAccept: viewerProfileAvatarAccept,
        isError: false,
        kind: "uploading",
        message: "アップロードしています。",
        previewUrl: resolvedPreviewUrl,
      };
    case "completed":
      return {
        canClear: true,
        fileName: state.file.name,
        inputAccept: viewerProfileAvatarAccept,
        isError: false,
        kind: "completed",
        message: "アップロード準備ができています。",
        previewUrl: resolvedPreviewUrl,
      };
    case "failed":
      return {
        canClear: true,
        fileName: state.file.name,
        inputAccept: viewerProfileAvatarAccept,
        isError: true,
        kind: "failed",
        message: state.message,
        previewUrl: resolvedPreviewUrl,
      };
  }
}

export function useViewerProfileDraft({
  initialValues = defaultInitialValues,
  mode = "edit",
  onDirty,
}: UseViewerProfileDraftOptions = {}) {
  const [displayName, setDisplayNameState] = useState(initialValues.displayName);
  const [handle, setHandleState] = useState(initialValues.handle);
  const [avatarState, setAvatarState] = useState<ViewerProfileAvatarState>({ kind: "empty" });
  const [avatarInputKey, setAvatarInputKey] = useState(0);
  const previewFile = getAvatarPreviewFile(avatarState);
  const avatarPreviewUrl = useMemo(() => {
    if (previewFile === null) {
      return null;
    }

    return URL.createObjectURL(previewFile);
  }, [previewFile]);

  useEffect(() => {
    if (avatarPreviewUrl === null) {
      return;
    }

    return () => {
      URL.revokeObjectURL(avatarPreviewUrl);
    };
  }, [avatarPreviewUrl]);

  const markDirty = () => {
    onDirty?.();
  };

  const setDisplayName = (nextDisplayName: string) => {
    setDisplayNameState(nextDisplayName);
    markDirty();
  };

  const setHandle = (nextHandle: string) => {
    setHandleState(nextHandle);
    markDirty();
  };

  const clearAvatarSelection = () => {
    setAvatarState({ kind: "empty" });
    setAvatarInputKey((currentKey) => currentKey + 1);
    markDirty();
  };

  const selectAvatarFile = (file: File | null) => {
    if (file === null) {
      clearAvatarSelection();
      return;
    }

    const validationMessage = validateAvatarFile(file);
    if (validationMessage !== null) {
      setAvatarState({
        fileName: file.name,
        kind: "invalid",
        message: validationMessage,
      });
      markDirty();
      return;
    }

    setAvatarState({
      file,
      kind: "selected",
    });
    markDirty();
  };

  const getProfileValidationError = (): string | null => {
    if (displayName.trim() === "") {
      return "表示名を入力してください。";
    }
    if (handle.trim() === "") {
      return "handleを入力してください。";
    }

    return null;
  };

  const getAvatarSubmissionError = (): string | null => {
    switch (avatarState.kind) {
      case "invalid":
        return avatarState.message;
      case "uploading":
        return "avatar のアップロード完了を待ってから再度お試しください。";
      default:
        return null;
    }
  };

  const uploadAvatarFile = async (file: File): Promise<CompletedAvatarUpload> => {
    setAvatarState({
      file,
      kind: "uploading",
    });

    try {
      const createResult = await createViewerProfileAvatarUpload(file);
      await uploadViewerProfileAvatarTarget({
        file,
        target: createResult.uploadTarget,
      });
      const completedResult = await completeViewerProfileAvatarUpload(createResult.avatarUploadToken);
      const completedUpload = {
        avatarAssetID: completedResult.avatar.id,
        avatarUploadToken: completedResult.avatarUploadToken,
        avatarURL: completedResult.avatar.url,
      };

      setAvatarState({
        completedUpload,
        file,
        kind: "completed",
      });

      return completedUpload;
    } catch (error) {
      setAvatarState({
        file,
        kind: "failed",
        message: getAvatarUploadErrorMessage(error),
      });
      throw error;
    }
  };

  const uploadAvatarIfNeeded = async (): Promise<string | undefined> => {
    switch (avatarState.kind) {
      case "empty":
        return undefined;
      case "invalid":
        throw new Error(avatarState.message);
      case "completed":
        return avatarState.completedUpload.avatarUploadToken;
      case "selected":
      case "failed":
        return (await uploadAvatarFile(avatarState.file)).avatarUploadToken;
      case "uploading":
        return undefined;
    }
  };

  const resetCompletedAvatarUploadToken = () => {
    setAvatarState((currentAvatarState) => {
      if (currentAvatarState.kind !== "completed") {
        return currentAvatarState;
      }

      return {
        file: currentAvatarState.file,
        kind: "failed",
        message: invalidCompletedAvatarMessage,
      };
    });
  };

  const resetDraft = () => {
    setDisplayNameState(initialValues.displayName);
    setHandleState(initialValues.handle);
    setAvatarState({ kind: "empty" });
    setAvatarInputKey((currentKey) => currentKey + 1);
  };

  return {
    avatar: buildAvatarField(avatarState, avatarPreviewUrl, {
      mode,
      persistedAvatarUrl: initialValues.avatarUrl,
    }),
    avatarInputKey,
    clearAvatarSelection,
    displayName,
    getAvatarSubmissionError,
    getProfileValidationError,
    handle,
    resetDraft,
    resetCompletedAvatarUploadToken,
    selectAvatarFile,
    setDisplayName,
    setHandle,
    uploadAvatarIfNeeded,
  };
}

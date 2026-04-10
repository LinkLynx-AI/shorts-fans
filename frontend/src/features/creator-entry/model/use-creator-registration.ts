"use client";

import { useRouter } from "next/navigation";
import {
  useEffect,
  startTransition,
  useState,
} from "react";

import {
  getCurrentViewerBootstrap,
  useSetCurrentViewer,
  useSetViewerSession,
} from "@/entities/viewer";
import { ApiError } from "@/shared/api";

import {
  completeCreatorRegistrationAvatarUpload,
  createCreatorRegistrationAvatarUpload,
  uploadCreatorRegistrationAvatarTarget,
} from "../api/avatar-upload";
import { registerCreator } from "../api/register-creator";
import {
  getCreatorEntryErrorCode,
  getCreatorRegistrationErrorMessage,
} from "./creator-entry";

const creatorRegistrationAvatarMimeTypes = ["image/jpeg", "image/png", "image/webp"] as const;
const creatorRegistrationAvatarAccept = creatorRegistrationAvatarMimeTypes.join(",");
const creatorRegistrationAvatarMimeTypeSet = new Set<string>(creatorRegistrationAvatarMimeTypes);
const creatorRegistrationAvatarMaxFileSizeBytes = 5_242_880;
const creatorRegistrationInvalidCompletedAvatarMessage =
  "avatar upload の有効期限が切れました。再度申し込むともう一度アップロードします。";

export type CreatorRegistrationFormMode = "edit" | "register";

export type CreatorRegistrationInitialValues = {
  avatarUrl: string | null;
  bio: string;
  displayName: string;
  handle: string;
};

type CompletedAvatarUpload = {
  avatarAssetID: string;
  avatarUploadToken: string;
  avatarURL: string;
};

type CreatorRegistrationAvatarState =
  | { kind: "empty" }
  | { fileName: string; kind: "invalid"; message: string }
  | { file: File; kind: "selected" }
  | { file: File; kind: "uploading" }
  | { completedUpload: CompletedAvatarUpload; file: File; kind: "completed" }
  | { file: File; kind: "failed"; message: string };

type CreatorRegistrationAvatarField = {
  canClear: boolean;
  fileName: string | null;
  inputAccept: string;
  isError: boolean;
  kind: CreatorRegistrationAvatarState["kind"];
  message: string;
  previewUrl: string | null;
};

type UseCreatorRegistrationResult = {
  avatar: CreatorRegistrationAvatarField;
  avatarInputKey: number;
  bio: string;
  clearAvatarSelection: () => void;
  displayName: string;
  errorMessage: string | null;
  handle: string;
  isSubmitting: boolean;
  selectAvatarFile: (file: File | null) => void;
  setBio: (bio: string) => void;
  setDisplayName: (displayName: string) => void;
  setHandle: (handle: string) => void;
  submit: () => Promise<void>;
};

type UseCreatorRegistrationOptions = {
  initialValues?: CreatorRegistrationInitialValues;
  mode?: CreatorRegistrationFormMode;
};

const defaultCreatorRegistrationInitialValues: CreatorRegistrationInitialValues = {
  avatarUrl: null,
  bio: "",
  displayName: "",
  handle: "",
};

function validateAvatarFile(file: File): string | null {
  if (file.size <= 0) {
    return "avatar file を読み取れませんでした。別の画像を選択してください。";
  }
  if (file.size > creatorRegistrationAvatarMaxFileSizeBytes) {
    return "avatar は 5MB 以下の画像を選択してください。";
  }
  if (!creatorRegistrationAvatarMimeTypeSet.has(file.type)) {
    return "avatar は JPEG / PNG / WebP のみ選択できます。";
  }

  return null;
}

function getAvatarUploadErrorMessage(error: unknown): string {
  const creatorEntryErrorCode = getCreatorEntryErrorCode(error);
  if (creatorEntryErrorCode !== null) {
    return getCreatorRegistrationErrorMessage(error);
  }

  if (error instanceof ApiError && error.code === "network") {
    return "avatar のアップロードに失敗しました。通信状態を確認してから再度お試しください。";
  }
  if (error instanceof ApiError) {
    return "avatar のアップロードに失敗しました。再度お試しください。";
  }

  return "avatar のアップロードに失敗しました。少し時間を置いてから再度お試しください。";
}

function getAvatarPreviewFile(state: CreatorRegistrationAvatarState): File | null {
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
  state: CreatorRegistrationAvatarState,
  previewUrl: string | null,
  {
    mode,
    persistedAvatarUrl,
  }: {
    mode: CreatorRegistrationFormMode;
    persistedAvatarUrl: string | null;
  },
): CreatorRegistrationAvatarField {
  const hasPersistedAvatar = mode === "edit" && persistedAvatarUrl !== null;
  const resolvedPreviewUrl = previewUrl ?? (hasPersistedAvatar ? persistedAvatarUrl : null);

  switch (state.kind) {
    case "empty":
      return {
        canClear: false,
        fileName: hasPersistedAvatar ? "現在の avatar" : null,
        inputAccept: creatorRegistrationAvatarAccept,
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
        inputAccept: creatorRegistrationAvatarAccept,
        isError: true,
        kind: "invalid",
        message: state.message,
        previewUrl: resolvedPreviewUrl,
      };
    case "selected":
      return {
        canClear: true,
        fileName: state.file.name,
        inputAccept: creatorRegistrationAvatarAccept,
        isError: false,
        kind: "selected",
        message: mode === "edit" ? "保存時にアップロードします。" : "登録時にアップロードします。",
        previewUrl: resolvedPreviewUrl,
      };
    case "uploading":
      return {
        canClear: true,
        fileName: state.file.name,
        inputAccept: creatorRegistrationAvatarAccept,
        isError: false,
        kind: "uploading",
        message: "アップロードしています。",
        previewUrl: resolvedPreviewUrl,
      };
    case "completed":
      return {
        canClear: true,
        fileName: state.file.name,
        inputAccept: creatorRegistrationAvatarAccept,
        isError: false,
        kind: "completed",
        message: "アップロード済みです。",
        previewUrl: resolvedPreviewUrl,
      };
    case "failed":
      return {
        canClear: true,
        fileName: state.file.name,
        inputAccept: creatorRegistrationAvatarAccept,
        isError: true,
        kind: "failed",
        message: state.message,
        previewUrl: resolvedPreviewUrl,
      };
  }
}

/**
 * creator registration form の入力状態と submit を管理する。
 */
export function useCreatorRegistration({
  initialValues = defaultCreatorRegistrationInitialValues,
  mode = "register",
}: UseCreatorRegistrationOptions = {}): UseCreatorRegistrationResult {
  const router = useRouter();
  const setCurrentViewer = useSetCurrentViewer();
  const setViewerSession = useSetViewerSession();
  const [bio, setBioState] = useState(initialValues.bio);
  const [displayName, setDisplayNameState] = useState(initialValues.displayName);
  const [handle, setHandleState] = useState(initialValues.handle);
  const [avatarState, setAvatarState] = useState<CreatorRegistrationAvatarState>({ kind: "empty" });
  const [avatarPreviewUrl, setAvatarPreviewUrl] = useState<string | null>(null);
  const [avatarInputKey, setAvatarInputKey] = useState(0);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    const previewFile = getAvatarPreviewFile(avatarState);
    if (previewFile === null) {
      setAvatarPreviewUrl(null);
      return;
    }

    const nextPreviewUrl = URL.createObjectURL(previewFile);
    setAvatarPreviewUrl(nextPreviewUrl);

    return () => {
      URL.revokeObjectURL(nextPreviewUrl);
    };
  }, [avatarState]);

  const clearErrorMessage = () => {
    if (errorMessage !== null) {
      setErrorMessage(null);
    }
  };

  const setDisplayName = (nextDisplayName: string) => {
    setDisplayNameState(nextDisplayName);
    clearErrorMessage();
  };

  const setBio = (nextBio: string) => {
    setBioState(nextBio);
    clearErrorMessage();
  };

  const setHandle = (nextHandle: string) => {
    setHandleState(nextHandle);
    clearErrorMessage();
  };

  const clearAvatarSelection = () => {
    setAvatarState({ kind: "empty" });
    setAvatarInputKey((currentKey) => currentKey + 1);
    clearErrorMessage();
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
      clearErrorMessage();
      return;
    }

    setAvatarState({
      file,
      kind: "selected",
    });
    clearErrorMessage();
  };

  const uploadAvatarFile = async (file: File): Promise<CompletedAvatarUpload> => {
    setAvatarState({
      file,
      kind: "uploading",
    });

    try {
      const createResult = await createCreatorRegistrationAvatarUpload(file);
      await uploadCreatorRegistrationAvatarTarget({
        file,
        target: createResult.uploadTarget,
      });
      const completedResult = await completeCreatorRegistrationAvatarUpload(createResult.avatarUploadToken);
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
      const message = getAvatarUploadErrorMessage(error);

      setAvatarState({
        file,
        kind: "failed",
        message,
      });

      throw error;
    }
  };

  const resolveAvatarUploadToken = async (currentAvatarState: CreatorRegistrationAvatarState): Promise<string | undefined> => {
    switch (currentAvatarState.kind) {
      case "empty":
        return undefined;
      case "invalid":
        setErrorMessage(currentAvatarState.message);
        return undefined;
      case "completed":
        return currentAvatarState.completedUpload.avatarUploadToken;
      case "selected":
      case "failed":
        return (await uploadAvatarFile(currentAvatarState.file)).avatarUploadToken;
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
        message: creatorRegistrationInvalidCompletedAvatarMessage,
      };
    });
  };

  const submit = async () => {
    if (isSubmitting) {
      return;
    }

    if (displayName.trim() === "") {
      setErrorMessage("表示名を入力してください。");
      return;
    }
    if (handle.trim() === "") {
      setErrorMessage("handleを入力してください。");
      return;
    }

    setIsSubmitting(true);
    setErrorMessage(null);

    const currentAvatarState = avatarState;
    if (currentAvatarState.kind === "invalid") {
      setErrorMessage(currentAvatarState.message);
      setIsSubmitting(false);
      return;
    }

    let avatarUploadToken: string | undefined;

    try {
      avatarUploadToken = await resolveAvatarUploadToken(currentAvatarState);
      if (currentAvatarState.kind !== "empty" && avatarUploadToken === undefined) {
        setIsSubmitting(false);
        return;
      }
    } catch (error) {
      setErrorMessage(getAvatarUploadErrorMessage(error));
      setIsSubmitting(false);
      return;
    }

    try {
      await registerCreator({
        ...(avatarUploadToken ? { avatarUploadToken } : {}),
        bio,
        displayName,
        handle,
      });

      if (mode === "edit") {
        startTransition(() => {
          router.replace("/creator");
        });
        return;
      }

      const currentViewer = await getCurrentViewerBootstrap({
        credentials: "include",
      }).catch(() => null);

      if (currentViewer === null) {
        setErrorMessage("登録後の状態反映を確認できませんでした。画面を更新して確認してください。");
        return;
      }

      setCurrentViewer(currentViewer);
      setViewerSession(true);

      startTransition(() => {
        router.push("/fan/creator/success");
      });
    } catch (error) {
      if (getCreatorEntryErrorCode(error) === "invalid_avatar_upload_token") {
        resetCompletedAvatarUploadToken();
      }
      setErrorMessage(getCreatorRegistrationErrorMessage(error));
    } finally {
      setIsSubmitting(false);
    }
  };

  return {
    avatar: buildAvatarField(avatarState, avatarPreviewUrl, {
      mode,
      persistedAvatarUrl: initialValues.avatarUrl,
    }),
    avatarInputKey,
    bio,
    clearAvatarSelection,
    displayName,
    errorMessage,
    handle,
    isSubmitting,
    selectAvatarFile,
    setBio,
    setDisplayName,
    setHandle,
    submit,
  };
}

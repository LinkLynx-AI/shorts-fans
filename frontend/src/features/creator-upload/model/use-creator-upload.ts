"use client";

import { useState } from "react";

import { ApiError } from "@/shared/api";

import {
  completeCreatorUploadPackage,
  createCreatorUploadPackage,
  CreatorUploadApiError,
  uploadCreatorUploadTarget,
} from "../api";
import {
  addCreatorUploadShortSlot,
  createInitialCreatorUploadDraft,
  getCreatorUploadSelectedShorts,
  isCreatorUploadReady,
  isCreatorUploadSubmitting,
  removeCreatorUploadShortSlot,
  setCreatorUploadMainFile,
  setCreatorUploadMainTransferState,
  setCreatorUploadShortFile,
  setCreatorUploadShortTransferState,
  setCreatorUploadSubmissionError,
  setCreatorUploadSubmissionSuccess,
  startCreatorUploadCompletion,
  startCreatorUploadInitiation,
  startCreatorUploadTransfer,
  type CreatorUploadDraft,
} from "./creator-upload-draft";

type UseCreatorUploadResult = {
  addShortSlot: () => void;
  draft: CreatorUploadDraft;
  removeShortSlot: (index: number) => void;
  selectMainFile: (file: File | null) => void;
  selectShortFile: (index: number, file: File | null) => void;
  submit: () => Promise<void>;
};

function getCreatorUploadApiErrorMessage(error: CreatorUploadApiError): string {
  switch (error.code) {
    case "auth_required":
      return "ログイン状態を確認してから再度お試しください。";
    case "capability_required":
      return "approved creator のアカウントで再度お試しください。";
    case "storage_failure":
      return "アップロード先の準備に失敗しました。少し時間を置いて再度お試しください。";
    case "upload_expired":
      return "アップロード準備の有効期限が切れました。再試行してください。";
    case "upload_failure":
      return "アップロード済みファイルを確認できませんでした。再試行してください。";
    case "validation_error":
      return "選択した動画ファイル情報を確認できませんでした。動画を選び直して再度お試しください。";
    default:
      return "アップロードを完了できませんでした。少し時間を置いて再度お試しください。";
  }
}

function getCreatorUploadUnknownErrorMessage(error: unknown): string {
  if (error instanceof CreatorUploadApiError) {
    return getCreatorUploadApiErrorMessage(error);
  }

  if (error instanceof ApiError) {
    return "アップロードを完了できませんでした。通信状態を確認して再度お試しください。";
  }

  return "アップロードを完了できませんでした。少し時間を置いて再度お試しください。";
}

function getCreatorUploadFileErrorMessage(error: unknown, fileLabel: string): string {
  if (error instanceof ApiError) {
    return `${fileLabel}のアップロードに失敗しました。通信状態を確認して再度お試しください。`;
  }

  return `${fileLabel}のアップロードに失敗しました。再試行してください。`;
}

/**
 * creator upload form の file 選択と upload workflow を管理する。
 */
export function useCreatorUpload(): UseCreatorUploadResult {
  const [draft, setDraft] = useState(createInitialCreatorUploadDraft);

  const selectMainFile = (file: File | null) => {
    setDraft((currentDraft) => setCreatorUploadMainFile(currentDraft, file));
  };

  const selectShortFile = (index: number, file: File | null) => {
    setDraft((currentDraft) => setCreatorUploadShortFile(currentDraft, index, file));
  };

  const addShortSlot = () => {
    setDraft((currentDraft) => addCreatorUploadShortSlot(currentDraft));
  };

  const removeShortSlot = (index: number) => {
    setDraft((currentDraft) => removeCreatorUploadShortSlot(currentDraft, index));
  };

  const submit = async () => {
    if (!isCreatorUploadReady(draft) || isCreatorUploadSubmitting(draft) || draft.mainFile === null) {
      return;
    }

    const mainFile = draft.mainFile;
    const selectedShorts = getCreatorUploadSelectedShorts(draft);

    setDraft((currentDraft) => startCreatorUploadInitiation(currentDraft));

    try {
      const createResult = await createCreatorUploadPackage({
        mainFile,
        shortFiles: selectedShorts.map((short) => short.file),
      });

      setDraft((currentDraft) => startCreatorUploadTransfer(currentDraft));
      setDraft((currentDraft) => setCreatorUploadMainTransferState(currentDraft, { kind: "uploading" }));

      try {
        await uploadCreatorUploadTarget({
          file: mainFile,
          target: createResult.uploadTargets.main,
        });

        setDraft((currentDraft) => setCreatorUploadMainTransferState(currentDraft, { kind: "uploaded" }));
      } catch (error) {
        const message = getCreatorUploadFileErrorMessage(error, "本編動画");

        setDraft((currentDraft) =>
          setCreatorUploadMainTransferState(currentDraft, {
            kind: "failed",
            message,
          }),
        );
        setDraft((currentDraft) => setCreatorUploadSubmissionError(currentDraft, message));
        return;
      }

      const shortUploadEntryIds: string[] = [];

      for (const [selectedIndex, selectedShort] of selectedShorts.entries()) {
        const target = createResult.uploadTargets.shorts[selectedIndex];

        if (!target) {
          setDraft((currentDraft) =>
            setCreatorUploadSubmissionError(
              currentDraft,
              "アップロード対象の応答が不足しています。再試行してください。",
            ),
          );
          return;
        }

        setDraft((currentDraft) =>
          setCreatorUploadShortTransferState(currentDraft, selectedShort.slotIndex, { kind: "uploading" }),
        );

        try {
          await uploadCreatorUploadTarget({
            file: selectedShort.file,
            target,
          });

          shortUploadEntryIds.push(target.uploadEntryId);
          setDraft((currentDraft) =>
            setCreatorUploadShortTransferState(currentDraft, selectedShort.slotIndex, { kind: "uploaded" }),
          );
        } catch (error) {
          const message = getCreatorUploadFileErrorMessage(error, `ショート動画 ${selectedShort.slotIndex + 1}`);

          setDraft((currentDraft) =>
            setCreatorUploadShortTransferState(currentDraft, selectedShort.slotIndex, {
              kind: "failed",
              message,
            }),
          );
          setDraft((currentDraft) => setCreatorUploadSubmissionError(currentDraft, message));
          return;
        }
      }

      setDraft((currentDraft) => startCreatorUploadCompletion(currentDraft));

      const completionResult = await completeCreatorUploadPackage({
        mainUploadEntryId: createResult.uploadTargets.main.uploadEntryId,
        packageToken: createResult.packageToken,
        shortUploadEntryIds,
      });

      setDraft((currentDraft) =>
        setCreatorUploadSubmissionSuccess(
          currentDraft,
          completionResult.main.id,
          completionResult.shorts.map((short) => short.id),
        ),
      );
    } catch (error) {
      setDraft((currentDraft) => setCreatorUploadSubmissionError(currentDraft, getCreatorUploadUnknownErrorMessage(error)));
    }
  };

  return {
    addShortSlot,
    draft,
    removeShortSlot,
    selectMainFile,
    selectShortFile,
    submit,
  };
}
